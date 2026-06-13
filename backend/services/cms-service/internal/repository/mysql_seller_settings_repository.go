package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

type MySQLSellerSettingsRepository struct {
	db *sql.DB
}

func NewMySQLSellerSettingsRepository(db *sql.DB) (*MySQLSellerSettingsRepository, error) {
	if db == nil {
		return nil, errors.New("mysql db is required")
	}
	return &MySQLSellerSettingsRepository{db: db}, nil
}

func (r *MySQLSellerSettingsRepository) GetSellerSettings(ctx context.Context, sellerID string) (domain.SellerSettings, error) {
	const query = `
SELECT seller_id, return_policy, shipping_policy, support_email, settings_json, created_at, updated_at
FROM seller_settings
WHERE seller_id = ?
LIMIT 1`
	return scanSellerSettings(r.db.QueryRowContext(ctx, query, strings.TrimSpace(sellerID)))
}

type sellerSettingsScanner interface {
	Scan(dest ...any) error
}

func scanSellerSettings(scanner sellerSettingsScanner) (domain.SellerSettings, error) {
	var settings domain.SellerSettings
	var returnPolicy sql.NullString
	var shippingPolicy sql.NullString
	var supportEmail sql.NullString
	var rawSettings []byte

	err := scanner.Scan(
		&settings.SellerID,
		&returnPolicy,
		&shippingPolicy,
		&supportEmail,
		&rawSettings,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SellerSettings{}, domain.ErrSellerSettingsNotFound
		}
		return domain.SellerSettings{}, err
	}
	if returnPolicy.Valid {
		settings.ReturnPolicy = returnPolicy.String
	}
	if shippingPolicy.Valid {
		settings.ShippingPolicy = shippingPolicy.String
	}
	if supportEmail.Valid {
		settings.SupportEmail = supportEmail.String
	}
	if len(rawSettings) > 0 {
		if err := json.Unmarshal(rawSettings, &settings.Settings); err != nil {
			return domain.SellerSettings{}, err
		}
	}
	return settings.Normalized(), nil
}
