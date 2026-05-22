package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

const (
	defaultAllowedClockSkew = 5 * time.Minute
	defaultMaxEventAge      = 24 * time.Hour
	defaultUserAgent        = "unknown"
	unavailableIPHash       = "unavailable"
)

type IDGenerator interface {
	NewID(prefix string) (string, error)
}

type IngestConfig struct {
	ActiveSessionTTL time.Duration
	AllowedClockSkew time.Duration
	MaxEventAge      time.Duration
	EventValidation  domain.SessionEventValidationConfig
	DefaultChannel   domain.Channel
	DefaultDevice    domain.DeviceType
	DefaultIPVersion domain.IPVersion
	IPHashSalt       string
}

type IngestEventInput struct {
	EventType             string
	AnonymousID           string
	SessionID             string
	UserID                *string
	OccurredAt            time.Time
	Path                  *string
	Properties            map[string]any
	RequestID             string
	UserAgent             string
	IPAddress             string
	IPHash                string
	IPVersion             domain.IPVersion
	Channel               domain.Channel
	DeviceType            domain.DeviceType
	DeviceFingerprint     string
	DeviceFingerprintHash string
	Locale                string
	Timezone              string
	ScreenWidth           int
	ScreenHeight          int
	ViewportWidth         int
	ViewportHeight        int
	ClientHints           ClientHints
	GeoHint               domain.Geo
}

type IngestEventOutput struct {
	Accepted  bool
	RequestID string
	EventID   string
}

type IngestUsecase struct {
	events    EventRepository
	sessions  SessionTouchRepository
	active    ActiveSessionStore
	publisher SessionEventPublisher
	cfg       IngestConfig
	logger    *slog.Logger
	clock     Clock
	ids       IDGenerator
	enricher  DeviceEnrichmentService
}

type IngestOption func(*IngestUsecase)

func WithDeviceEnricher(enricher DeviceEnrichmentService) IngestOption {
	return func(u *IngestUsecase) {
		if enricher != nil {
			u.enricher = enricher
		}
	}
}

func NewIngestUsecase(
	events EventRepository,
	sessions SessionTouchRepository,
	active ActiveSessionStore,
	publisher SessionEventPublisher,
	cfg IngestConfig,
	logger *slog.Logger,
	options ...IngestOption,
) (*IngestUsecase, error) {
	if events == nil {
		return nil, errors.New("event repository is required")
	}
	if sessions == nil {
		return nil, errors.New("session touch repository is required")
	}
	if active == nil {
		return nil, errors.New("active session store is required")
	}
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	usecase := &IngestUsecase{
		events:    events,
		sessions:  sessions,
		active:    active,
		publisher: publisher,
		cfg:       cfg,
		logger:    logger,
		clock:     realClock{},
		ids:       randomIDGenerator{},
	}
	for _, option := range options {
		if option != nil {
			option(usecase)
		}
	}
	return usecase, nil
}

func (u *IngestUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *IngestUsecase) WithIDGenerator(ids IDGenerator) {
	if ids != nil {
		u.ids = ids
	}
}

func (u *IngestUsecase) IngestEvent(ctx context.Context, input IngestEventInput) (IngestEventOutput, error) {
	if err := ctx.Err(); err != nil {
		return IngestEventOutput{}, err
	}

	opStarted := time.Now()
	started := u.clock.Now().UTC()
	normalized := u.normalizeInput(input)
	if err := u.validateInput(normalized, started); err != nil {
		return IngestEventOutput{}, err
	}

	requestID := strings.TrimSpace(normalized.RequestID)
	if requestID == "" {
		generated, err := u.ids.NewID("req")
		if err != nil {
			return IngestEventOutput{}, fmt.Errorf("generate request id: %w", err)
		}
		requestID = generated
	}
	eventID, err := u.ids.NewID("evt")
	if err != nil {
		return IngestEventOutput{}, fmt.Errorf("generate event id: %w", err)
	}

	cleanProps := sanitizeProperties(normalized.Properties)
	enriched := u.enrichDevice(ctx, normalized, cleanProps)
	eventDevice := enriched.Device
	eventClient := enriched.Client
	eventGeo := enriched.Geo
	requestIDPtr := requestID
	event := domain.SessionEvent{
		EventID:               eventID,
		SessionID:             normalized.SessionID,
		AnonymousID:           normalized.AnonymousID,
		UserID:                normalized.UserID,
		EventType:             domain.EventType(normalized.EventType),
		Path:                  normalized.Path,
		Properties:            cleanProps,
		UserAgentHash:         enriched.UserAgentHash,
		IPHash:                stringPtrIfNotEmpty(enriched.IPHash),
		IPVersion:             enriched.IPVersion,
		DeviceFingerprintHash: enriched.DeviceFingerprintHash,
		Device:                &eventDevice,
		Client:                &eventClient,
		Geo:                   &eventGeo,
		OccurredAt:            normalized.OccurredAt,
		ReceivedAt:            started,
		SchemaVersion:         domain.CurrentSessionEventSchemaVersion,
		RequestID:             &requestIDPtr,
	}.Normalize()
	if err := event.ValidateWithConfig(u.cfg.EventValidation); err != nil {
		return IngestEventOutput{}, fmt.Errorf("%w: %w", ErrInvalidSessionInput, err)
	}

	if err := u.events.InsertEvent(ctx, event); err != nil {
		u.logger.ErrorContext(ctx, "session.ingest.event_insert_failed",
			slog.String("request_id", requestID),
			slog.String("event_type", string(event.EventType)),
			slog.String("session_id", event.SessionID),
			slog.String("error", err.Error()),
		)
		return IngestEventOutput{}, fmt.Errorf("%w: insert event: %w", ErrIngestStorageUnavailable, err)
	}

	touch := u.sessionTouchFromInput(normalized, event, enriched)
	if err := u.sessions.TouchSession(ctx, touch); err != nil {
		u.logger.ErrorContext(ctx, "session.ingest.session_touch_failed",
			slog.String("request_id", requestID),
			slog.String("event_id", event.EventID),
			slog.String("session_id", event.SessionID),
			slog.String("error", err.Error()),
		)
		return IngestEventOutput{}, fmt.Errorf("%w: touch session: %w", ErrIngestStorageUnavailable, err)
	}

	snapshot := domain.ActiveSessionSnapshotFromTouch(touch)
	if err := u.active.Touch(ctx, snapshot, u.cfg.ActiveSessionTTL); err != nil {
		u.logger.WarnContext(ctx, "session.ingest.active_touch_failed",
			slog.String("request_id", requestID),
			slog.String("event_id", event.EventID),
			slog.String("session_id", event.SessionID),
			slog.String("error", err.Error()),
		)
	}

	if u.publisher != nil {
		if err := u.publisher.PublishSessionEvent(ctx, event); err != nil {
			u.logger.WarnContext(ctx, "session.ingest.publish_failed",
				slog.String("request_id", requestID),
				slog.String("event_id", event.EventID),
				slog.String("session_id", event.SessionID),
				slog.String("error", err.Error()),
			)
		}
	}

	u.logger.InfoContext(ctx, "session.ingest.event_accepted",
		slog.String("request_id", requestID),
		slog.String("event_id", event.EventID),
		slog.String("event_type", string(event.EventType)),
		slog.String("session_id", event.SessionID),
		slog.Int64("duration_ms", time.Since(opStarted).Milliseconds()),
	)
	return IngestEventOutput{
		Accepted:  true,
		RequestID: requestID,
		EventID:   event.EventID,
	}, nil
}

func (u *IngestUsecase) normalizeInput(input IngestEventInput) IngestEventInput {
	out := input
	out.EventType = strings.TrimSpace(out.EventType)
	out.AnonymousID = strings.TrimSpace(out.AnonymousID)
	out.SessionID = strings.TrimSpace(out.SessionID)
	out.UserID = trimStringPtr(out.UserID)
	out.Path = trimStringPtr(out.Path)
	out.RequestID = strings.TrimSpace(out.RequestID)
	out.UserAgent = strings.TrimSpace(out.UserAgent)
	out.IPAddress = strings.TrimSpace(out.IPAddress)
	out.IPHash = strings.TrimSpace(out.IPHash)
	out.IPVersion = domain.IPVersion(strings.TrimSpace(string(out.IPVersion)))
	out.Channel = domain.Channel(strings.TrimSpace(string(out.Channel)))
	out.DeviceType = domain.DeviceType(strings.TrimSpace(string(out.DeviceType)))
	out.DeviceFingerprint = strings.TrimSpace(out.DeviceFingerprint)
	out.DeviceFingerprintHash = strings.TrimSpace(out.DeviceFingerprintHash)
	out.Locale = strings.TrimSpace(out.Locale)
	out.Timezone = strings.TrimSpace(out.Timezone)
	out.ClientHints = ClientHints{
		UA:       strings.TrimSpace(out.ClientHints.UA),
		Platform: strings.TrimSpace(out.ClientHints.Platform),
		Mobile:   strings.TrimSpace(out.ClientHints.Mobile),
		Model:    strings.TrimSpace(out.ClientHints.Model),
	}
	out.GeoHint = out.GeoHint.Normalize()
	if out.Channel == "" {
		out.Channel = u.cfg.DefaultChannel
	}
	if out.DeviceType == "" {
		out.DeviceType = u.cfg.DefaultDevice
	}
	if out.IPVersion == "" {
		out.IPVersion = u.detectIPVersion(out.IPAddress)
	}
	if out.IPVersion == "" {
		out.IPVersion = u.cfg.DefaultIPVersion
	}
	if out.UserAgent == "" {
		out.UserAgent = defaultUserAgent
	}
	return out
}

func (u *IngestUsecase) validateInput(input IngestEventInput, now time.Time) error {
	violations := make([]domain.FieldViolation, 0)
	eventType := domain.EventType(input.EventType)
	if input.EventType == "" {
		violations = append(violations, domain.FieldViolation{Field: "event_type", Message: "is required"})
	} else if !eventType.SupportedForIngestion() {
		violations = append(violations, domain.FieldViolation{Field: "event_type", Message: "is not supported"})
	}
	if input.AnonymousID == "" {
		violations = append(violations, domain.FieldViolation{Field: "anonymous_id", Message: "is required"})
	} else if !strings.HasPrefix(input.AnonymousID, "anon_") {
		violations = append(violations, domain.FieldViolation{Field: "anonymous_id", Message: "must start with anon_"})
	}
	if input.SessionID == "" {
		violations = append(violations, domain.FieldViolation{Field: "session_id", Message: "is required"})
	} else if !strings.HasPrefix(input.SessionID, "sess_") {
		violations = append(violations, domain.FieldViolation{Field: "session_id", Message: "must start with sess_"})
	}
	if input.OccurredAt.IsZero() {
		violations = append(violations, domain.FieldViolation{Field: "occurred_at", Message: "is required"})
	} else {
		occurredAt := input.OccurredAt.UTC()
		if occurredAt.After(now.Add(u.cfg.AllowedClockSkew)) {
			violations = append(violations, domain.FieldViolation{Field: "occurred_at", Message: "is too far in the future"})
		}
		if occurredAt.Before(now.Add(-u.cfg.MaxEventAge)) {
			violations = append(violations, domain.FieldViolation{Field: "occurred_at", Message: "is too old"})
		}
	}
	if input.Channel != "" && !input.Channel.Valid() {
		violations = append(violations, domain.FieldViolation{Field: "channel", Message: "must be a supported channel"})
	}
	if input.DeviceType != "" && !input.DeviceType.Valid() {
		violations = append(violations, domain.FieldViolation{Field: "device_type", Message: "must be a supported device type"})
	}
	if input.IPVersion != "" && !input.IPVersion.Valid() {
		violations = append(violations, domain.FieldViolation{Field: "ip_version", Message: "must be ipv4, ipv6, or unknown"})
	}
	if len(violations) > 0 {
		return fmt.Errorf("%w: %w", ErrInvalidSessionInput, domain.SessionEventValidationError{Violations: violations})
	}
	return nil
}

func (u *IngestUsecase) enrichDevice(ctx context.Context, input IngestEventInput, properties map[string]any) domain.EnrichedDeviceContext {
	enrichmentInput := DeviceEnrichmentInput{
		SessionID:             input.SessionID,
		AnonymousID:           input.AnonymousID,
		UserID:                input.UserID,
		UserAgent:             input.UserAgent,
		ClientIP:              input.IPAddress,
		IPHash:                input.IPHash,
		DeviceFingerprint:     input.DeviceFingerprint,
		DeviceFingerprintHash: input.DeviceFingerprintHash,
		Channel:               input.Channel,
		DeviceType:            input.DeviceType,
		Locale:                firstNonEmpty(input.Locale, stringFromProperty(properties, "locale"), stringFromProperty(properties, "language")),
		Timezone:              firstNonEmpty(input.Timezone, stringFromProperty(properties, "timezone")),
		ScreenWidth:           firstPositive(input.ScreenWidth, intFromProperty(properties["screen_width"])),
		ScreenHeight:          firstPositive(input.ScreenHeight, intFromProperty(properties["screen_height"])),
		ViewportWidth:         firstPositive(input.ViewportWidth, intFromProperty(properties["viewport_width"])),
		ViewportHeight:        firstPositive(input.ViewportHeight, intFromProperty(properties["viewport_height"])),
		ClientHints:           input.ClientHints,
		GeoHint:               input.GeoHint,
		OccurredAt:            input.OccurredAt,
	}
	if u.enricher != nil {
		return u.enricher.Enrich(ctx, enrichmentInput).Normalize()
	}
	ipHash := u.ipHash(input)
	deviceType := normalizeDeviceType(input.DeviceType, u.cfg.DefaultDevice)
	return domain.EnrichedDeviceContext{
		UserAgent: defaultUserAgentFromInput(input.UserAgent),
		IPHash:    ipHash,
		IPVersion: detectIPVersion(input.IPAddress, u.cfg.DefaultIPVersion),
		Device: domain.Device{
			Type: deviceType,
		},
		Client: domain.Client{
			Channel:        input.Channel,
			Locale:         stringPtrIfNotEmpty(enrichmentInput.Locale),
			Timezone:       stringPtrIfNotEmpty(enrichmentInput.Timezone),
			ScreenWidth:    normalizeDimension(enrichmentInput.ScreenWidth),
			ScreenHeight:   normalizeDimension(enrichmentInput.ScreenHeight),
			ViewportWidth:  normalizeDimension(enrichmentInput.ViewportWidth),
			ViewportHeight: normalizeDimension(enrichmentInput.ViewportHeight),
		},
		Geo: input.GeoHint,
	}.Normalize()
}

func (u *IngestUsecase) sessionTouchFromInput(input IngestEventInput, event domain.SessionEvent, enriched domain.EnrichedDeviceContext) domain.SessionTouch {
	return domain.SessionTouch{
		SessionID:             event.SessionID,
		AnonymousID:           event.AnonymousID,
		UserID:                event.UserID,
		EventType:             event.EventType,
		Path:                  event.Path,
		UserAgent:             enriched.UserAgent,
		UserAgentHash:         enriched.UserAgentHash,
		IPHash:                enriched.IPHash,
		IPVersion:             enriched.IPVersion,
		Channel:               enriched.Client.Channel,
		DeviceType:            enriched.Device.Type,
		Device:                enriched.Device,
		Client:                enriched.Client,
		Geo:                   enriched.Geo,
		DeviceFingerprintHash: enriched.DeviceFingerprintHash,
		OccurredAt:            event.OccurredAt,
		ReceivedAt:            event.ReceivedAt,
		SchemaVersion:         domain.CurrentSessionSchemaVersion,
	}.Normalize()
}

func (u *IngestUsecase) ipHash(input IngestEventInput) string {
	if input.IPHash != "" && net.ParseIP(input.IPHash) == nil {
		return input.IPHash
	}
	ip := strings.TrimSpace(input.IPAddress)
	if ip == "" || strings.TrimSpace(u.cfg.IPHashSalt) == "" {
		return unavailableIPHash
	}
	mac := hmac.New(sha256.New, []byte(u.cfg.IPHashSalt))
	_, _ = mac.Write([]byte(ip))
	return hex.EncodeToString(mac.Sum(nil))
}

func (u *IngestUsecase) detectIPVersion(value string) domain.IPVersion {
	return detectIPVersion(value, u.cfg.DefaultIPVersion)
}

func sanitizeProperties(input map[string]any) map[string]any {
	if len(input) == 0 {
		return nil
	}
	clean := make(map[string]any, len(input))
	for key, value := range input {
		key = strings.TrimSpace(key)
		if key == "" || sensitivePropertyName(key) {
			continue
		}
		clean[key] = sanitizePropertyValue(value)
	}
	if len(clean) == 0 {
		return nil
	}
	return clean
}

func sanitizePropertyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return sanitizeProperties(typed)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, sanitizePropertyValue(item))
		}
		return out
	case string:
		return strings.TrimSpace(typed)
	default:
		return typed
	}
}

func stringFromProperty(properties map[string]any, key string) string {
	if len(properties) == 0 {
		return ""
	}
	value, ok := properties[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func defaultUserAgentFromInput(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultUserAgent
	}
	return value
}

func sensitivePropertyName(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	normalized = strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(normalized)
	sensitiveFragments := []string{
		"password",
		"otp",
		"card_number",
		"credit_card",
		"cvv",
		"pin",
		"token",
		"authorization",
		"cookie",
		"raw_ip",
		"private_message",
		"secret",
		"keystroke",
		"field_value",
		"input_value",
	}
	for _, fragment := range sensitiveFragments {
		if normalized == fragment || strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}

func trimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (c IngestConfig) Validate() error {
	if c.ActiveSessionTTL <= 0 {
		return errors.New("SESSION_ACTIVE_TTL must be greater than zero")
	}
	if c.AllowedClockSkew <= 0 {
		return errors.New("SESSION_ALLOWED_CLOCK_SKEW must be greater than zero")
	}
	if c.MaxEventAge <= 0 {
		return errors.New("SESSION_MAX_EVENT_AGE must be greater than zero")
	}
	if !c.DefaultChannel.Valid() {
		return errors.New("SESSION_DEFAULT_CHANNEL must be a supported channel")
	}
	if !c.DefaultDevice.Valid() {
		return errors.New("SESSION_DEFAULT_DEVICE_TYPE must be a supported device type")
	}
	if !c.DefaultIPVersion.Valid() {
		return errors.New("SESSION_DEFAULT_IP_VERSION must be a supported IP version")
	}
	return nil
}

func (c IngestConfig) withDefaults() IngestConfig {
	if c.ActiveSessionTTL == 0 {
		c.ActiveSessionTTL = defaultActiveSessionTTL
	}
	if c.AllowedClockSkew == 0 {
		c.AllowedClockSkew = defaultAllowedClockSkew
	}
	if c.MaxEventAge == 0 {
		c.MaxEventAge = defaultMaxEventAge
	}
	c.EventValidation = c.EventValidation.WithDefaults()
	if c.DefaultChannel == "" {
		c.DefaultChannel = domain.ChannelUnknown
	}
	if c.DefaultDevice == "" {
		c.DefaultDevice = domain.DeviceTypeUnknown
	}
	if c.DefaultIPVersion == "" {
		c.DefaultIPVersion = domain.IPVersionUnknown
	}
	c.IPHashSalt = strings.TrimSpace(c.IPHashSalt)
	return c
}

type randomIDGenerator struct{}

func (randomIDGenerator) NewID(prefix string) (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes[:])
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return strings.ToLower(encoded), nil
	}
	return prefix + "_" + strings.ToLower(encoded), nil
}
