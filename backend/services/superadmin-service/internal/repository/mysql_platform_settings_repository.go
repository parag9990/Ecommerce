package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"ecommerce/superadmin-service/internal/domain"
)

type MySQLPlatformSettingsRepository struct {
	db *sql.DB
}

func NewMySQLPlatformSettingsRepository(db *sql.DB) (*MySQLPlatformSettingsRepository, error) {
	if db == nil {
		return nil, errors.New("mysql platform settings repository requires db")
	}
	return &MySQLPlatformSettingsRepository{db: db}, nil
}

func (r *MySQLPlatformSettingsRepository) ListSettings(ctx context.Context) ([]domain.PlatformSetting, error) {
	const query = `
		SELECT setting_key,
		       setting_type,
		       value_json,
		       risk_level,
		       version,
		       updated_by_admin_id,
		       update_reason,
		       created_at,
		       updated_at
		FROM platform_settings
		ORDER BY setting_key`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list platform settings: %w", err)
	}
	defer rows.Close()

	settings := make([]domain.PlatformSetting, 0)
	for rows.Next() {
		setting, err := scanPlatformSetting(rows)
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate platform settings: %w", err)
	}
	return settings, nil
}

func (r *MySQLPlatformSettingsRepository) GetSetting(ctx context.Context, key domain.PlatformSettingKey) (domain.PlatformSetting, error) {
	if !key.Valid() {
		return domain.PlatformSetting{}, domain.NewValidationError(fmt.Sprintf("unsupported platform setting key %q", key))
	}

	const query = `
		SELECT setting_key,
		       setting_type,
		       value_json,
		       risk_level,
		       version,
		       updated_by_admin_id,
		       update_reason,
		       created_at,
		       updated_at
		FROM platform_settings
		WHERE setting_key = ?`

	setting, err := scanPlatformSetting(r.db.QueryRowContext(ctx, query, string(key)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.PlatformSetting{}, domain.NewPlatformSettingNotFound(key)
		}
		return domain.PlatformSetting{}, err
	}
	return setting, nil
}

func (r *MySQLPlatformSettingsRepository) UpdateSetting(ctx context.Context, setting domain.PlatformSetting, expectedVersion uint64) (domain.PlatformSetting, error) {
	if !setting.Key.Valid() {
		return domain.PlatformSetting{}, domain.NewValidationError(fmt.Sprintf("unsupported platform setting key %q", setting.Key))
	}
	if expectedVersion == 0 {
		return domain.PlatformSetting{}, domain.NewValidationError("expected setting version is required")
	}
	if strings.TrimSpace(setting.UpdatedByAdminID) == "" {
		return domain.PlatformSetting{}, domain.NewValidationError("updated_by_admin_id is required")
	}
	if strings.TrimSpace(setting.UpdateReason) == "" {
		return domain.PlatformSetting{}, domain.NewReasonRequired()
	}

	valueJSON, err := json.Marshal(setting.Value)
	if err != nil {
		return domain.PlatformSetting{}, domain.NewValidationError("setting value must be JSON serializable")
	}

	const query = `
		UPDATE platform_settings
		SET value_json = ?,
		    version = version + 1,
		    updated_by_admin_id = ?,
		    update_reason = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE setting_key = ?
		  AND version = ?`

	result, err := r.db.ExecContext(
		ctx,
		query,
		string(valueJSON),
		setting.UpdatedByAdminID,
		setting.UpdateReason,
		string(setting.Key),
		expectedVersion,
	)
	if err != nil {
		return domain.PlatformSetting{}, fmt.Errorf("update platform setting: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return domain.PlatformSetting{}, fmt.Errorf("check platform setting update result: %w", err)
	}
	if affected == 0 {
		latest, latestErr := r.GetSetting(ctx, setting.Key)
		if latestErr != nil {
			return domain.PlatformSetting{}, latestErr
		}
		return domain.PlatformSetting{}, domain.NewSettingVersionConflict(setting.Key, expectedVersion, latest.Version)
	}

	return r.GetSetting(ctx, setting.Key)
}

type UnavailablePlatformSettingsRepository struct {
	message string
}

func NewUnavailablePlatformSettingsRepository(message string) *UnavailablePlatformSettingsRepository {
	if strings.TrimSpace(message) == "" {
		message = "platform settings storage is not configured"
	}
	return &UnavailablePlatformSettingsRepository{message: message}
}

func (r *UnavailablePlatformSettingsRepository) ListSettings(ctx context.Context) ([]domain.PlatformSetting, error) {
	return nil, domain.NewDownstreamUnavailable(r.message, nil)
}

func (r *UnavailablePlatformSettingsRepository) GetSetting(ctx context.Context, key domain.PlatformSettingKey) (domain.PlatformSetting, error) {
	return domain.PlatformSetting{}, domain.NewDownstreamUnavailable(r.message, nil)
}

func (r *UnavailablePlatformSettingsRepository) UpdateSetting(ctx context.Context, setting domain.PlatformSetting, expectedVersion uint64) (domain.PlatformSetting, error) {
	return domain.PlatformSetting{}, domain.NewDownstreamUnavailable(r.message, nil)
}

type platformSettingScanner interface {
	Scan(dest ...any) error
}

func scanPlatformSetting(scanner platformSettingScanner) (domain.PlatformSetting, error) {
	var setting domain.PlatformSetting
	var key string
	var settingType string
	var rawValue []byte
	var risk string
	var createdAt sql.NullTime
	var updatedAt sql.NullTime

	if err := scanner.Scan(
		&key,
		&settingType,
		&rawValue,
		&risk,
		&setting.Version,
		&setting.UpdatedByAdminID,
		&setting.UpdateReason,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.PlatformSetting{}, err
	}

	setting.Key = domain.PlatformSettingKey(key)
	if !setting.Key.Valid() {
		return domain.PlatformSetting{}, domain.NewInternal(fmt.Sprintf("database contains unknown platform setting key %q", key), nil)
	}
	setting.Type = domain.SettingType(settingType)
	if !setting.Type.Valid() {
		return domain.PlatformSetting{}, domain.NewInternal(fmt.Sprintf("database contains unknown platform setting type %q", settingType), nil)
	}
	setting.Risk = domain.RiskLevel(risk)
	if _, ok := domain.PlatformSettingDefinitionByKey(setting.Key); !ok {
		return domain.PlatformSetting{}, domain.NewInternal(fmt.Sprintf("database contains unregistered platform setting key %q", key), nil)
	}
	definition, _ := domain.PlatformSettingDefinitionByKey(setting.Key)
	if definition.Type != setting.Type || definition.Risk != setting.Risk {
		return domain.PlatformSetting{}, domain.NewInternal(fmt.Sprintf("database platform setting metadata mismatch for key %q", key), nil)
	}

	if err := json.Unmarshal(rawValue, &setting.Value); err != nil {
		return domain.PlatformSetting{}, domain.NewInternal("platform setting value_json is invalid JSON", err)
	}
	if createdAt.Valid {
		setting.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		setting.UpdatedAt = updatedAt.Time
	}
	return setting, nil
}
