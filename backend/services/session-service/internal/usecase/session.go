package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

const defaultInactivityTimeout = 30 * time.Minute

type Config struct {
	SchemaVersion     int
	Validation        domain.ValidationConfig
	InactivityTimeout time.Duration
	DefaultStatus     domain.SessionStatus
	DefaultChannel    domain.Channel
	DefaultDeviceType domain.DeviceType
	DefaultIPVersion  domain.IPVersion
	DefaultRiskLevel  domain.RiskLevel
}

func DefaultConfig() Config {
	return Config{
		SchemaVersion:     domain.CurrentSessionSchemaVersion,
		Validation:        domain.DefaultValidationConfig(),
		InactivityTimeout: defaultInactivityTimeout,
		DefaultStatus:     domain.SessionStatusActive,
		DefaultChannel:    domain.ChannelUnknown,
		DefaultDeviceType: domain.DeviceTypeUnknown,
		DefaultIPVersion:  domain.IPVersionUnknown,
		DefaultRiskLevel:  domain.RiskLevelUnknown,
	}
}

type SessionUsecase struct {
	cfg    Config
	logger *slog.Logger
	clock  Clock
}

func NewSessionUsecase(cfg Config, logger *slog.Logger) (*SessionUsecase, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SessionUsecase{
		cfg:    cfg,
		logger: logger,
		clock:  realClock{},
	}, nil
}

func (u *SessionUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *SessionUsecase) PrepareSession(ctx context.Context, session domain.Session) (domain.Session, error) {
	if err := ctx.Err(); err != nil {
		return domain.Session{}, err
	}

	now := u.clock.Now().UTC()
	prepared := session.Normalize()
	if prepared.SchemaVersion == 0 {
		prepared.SchemaVersion = u.cfg.SchemaVersion
	}
	if prepared.Status == "" {
		prepared.Status = u.cfg.DefaultStatus
	}
	if prepared.Channel == "" {
		prepared.Channel = u.cfg.DefaultChannel
	}
	if prepared.Device.Type == "" {
		prepared.Device.Type = u.cfg.DefaultDeviceType
	}
	if prepared.IPVersion == "" {
		prepared.IPVersion = u.cfg.DefaultIPVersion
	}
	if prepared.RiskLevel == "" {
		prepared.RiskLevel = u.cfg.DefaultRiskLevel
	}
	if prepared.StartedAt.IsZero() {
		prepared.StartedAt = now
	}
	if prepared.LastSeenAt.IsZero() {
		prepared.LastSeenAt = prepared.StartedAt
	}
	if prepared.CreatedAt.IsZero() {
		prepared.CreatedAt = prepared.StartedAt
	}
	if prepared.UpdatedAt.IsZero() {
		prepared.UpdatedAt = prepared.LastSeenAt
	}

	prepared = prepared.Normalize()
	if err := prepared.ValidateWithConfig(u.cfg.Validation); err != nil {
		u.logger.WarnContext(ctx, "session.model.validation_failed",
			slog.String("session_id", prepared.SessionID),
			slog.String("anonymous_id", prepared.AnonymousID),
			slog.String("error", err.Error()),
		)
		return domain.Session{}, fmt.Errorf("%w: %w", ErrInvalidSessionInput, err)
	}

	u.logger.DebugContext(ctx, "session.model.validated",
		slog.String("session_id", prepared.SessionID),
		slog.String("anonymous_id", prepared.AnonymousID),
		slog.String("status", string(prepared.Status)),
		slog.String("channel", string(prepared.Channel)),
	)
	return prepared, nil
}

func (c Config) Validate() error {
	if c.SchemaVersion != domain.CurrentSessionSchemaVersion {
		return fmt.Errorf("SESSION_SCHEMA_VERSION must be %d", domain.CurrentSessionSchemaVersion)
	}
	if c.InactivityTimeout <= 0 {
		return errors.New("SESSION_INACTIVITY_TIMEOUT must be greater than zero")
	}
	if !c.DefaultStatus.Valid() {
		return errors.New("SESSION_DEFAULT_STATUS must be a supported status")
	}
	if !c.DefaultChannel.Valid() {
		return errors.New("SESSION_DEFAULT_CHANNEL must be a supported channel")
	}
	if !c.DefaultDeviceType.Valid() {
		return errors.New("SESSION_DEFAULT_DEVICE_TYPE must be a supported device type")
	}
	if !c.DefaultIPVersion.Valid() {
		return errors.New("SESSION_DEFAULT_IP_VERSION must be a supported IP version")
	}
	if !c.DefaultRiskLevel.Valid() {
		return errors.New("SESSION_DEFAULT_RISK_LEVEL must be a supported risk level")
	}
	return nil
}

func (c Config) withDefaults() Config {
	defaults := DefaultConfig()
	if c.SchemaVersion == 0 {
		c.SchemaVersion = defaults.SchemaVersion
	}
	c.Validation = c.Validation.WithDefaults()
	if c.InactivityTimeout == 0 {
		c.InactivityTimeout = defaults.InactivityTimeout
	}
	if c.DefaultStatus == "" {
		c.DefaultStatus = defaults.DefaultStatus
	}
	if c.DefaultChannel == "" {
		c.DefaultChannel = defaults.DefaultChannel
	}
	if c.DefaultDeviceType == "" {
		c.DefaultDeviceType = defaults.DefaultDeviceType
	}
	if c.DefaultIPVersion == "" {
		c.DefaultIPVersion = defaults.DefaultIPVersion
	}
	if c.DefaultRiskLevel == "" {
		c.DefaultRiskLevel = defaults.DefaultRiskLevel
	}
	return c
}
