package usecase

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

const (
	defaultGeoLookupTimeout = 20 * time.Millisecond
	redactedUserAgent       = "redacted"
)

type DeviceEnrichmentService interface {
	Enrich(ctx context.Context, input DeviceEnrichmentInput) domain.EnrichedDeviceContext
}

type DeviceEnrichmentInput struct {
	SessionID             string
	AnonymousID           string
	UserID                *string
	UserAgent             string
	ClientIP              string
	IPHash                string
	DeviceFingerprint     string
	DeviceFingerprintHash string
	Channel               domain.Channel
	DeviceType            domain.DeviceType
	Locale                string
	Timezone              string
	ScreenWidth           int
	ScreenHeight          int
	ViewportWidth         int
	ViewportHeight        int
	ClientHints           ClientHints
	GeoHint               domain.Geo
	OccurredAt            time.Time
}

type ClientHints struct {
	UA       string
	Platform string
	Mobile   string
	Model    string
}

type ParsedUserAgent struct {
	Browser        string
	BrowserVersion string
	OS             string
	OSVersion      string
	DeviceType     domain.DeviceType
	Model          string
	Vendor         string
	IsBot          bool
}

type UserAgentParser interface {
	Parse(ctx context.Context, userAgent string) ParsedUserAgent
}

type GeoResolver interface {
	Resolve(ctx context.Context, clientIP string) domain.Geo
}

type PrivacyHasher interface {
	Hash(value string) string
}

type DeviceEnricherConfig struct {
	Enabled            bool
	StoreUserAgent     bool
	UserAgentMaxLength int
	GeoLookupTimeout   time.Duration
	DefaultChannel     domain.Channel
	DefaultDevice      domain.DeviceType
	DefaultIPVersion   domain.IPVersion
}

type DeviceEnricher struct {
	parser UserAgentParser
	geo    GeoResolver
	hasher PrivacyHasher
	cfg    DeviceEnricherConfig
	logger *slog.Logger
}

func NewDeviceEnricher(parser UserAgentParser, geo GeoResolver, hasher PrivacyHasher, cfg DeviceEnricherConfig, logger *slog.Logger) (*DeviceEnricher, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if parser == nil {
		parser = unknownUserAgentParser{}
	}
	if geo == nil {
		geo = disabledGeoResolver{}
	}
	if hasher == nil {
		hasher = noopHasher{}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &DeviceEnricher{
		parser: parser,
		geo:    geo,
		hasher: hasher,
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (e *DeviceEnricher) Enrich(ctx context.Context, input DeviceEnrichmentInput) domain.EnrichedDeviceContext {
	if err := ctx.Err(); err != nil {
		return e.fallback(input).Normalize()
	}

	userAgentForParsing := trimMax(input.UserAgent, e.cfg.UserAgentMaxLength)
	if userAgentForParsing == "" {
		userAgentForParsing = defaultUserAgent
	}
	storedUserAgent := userAgentForParsing
	if !e.cfg.StoreUserAgent {
		storedUserAgent = redactedUserAgent
	}

	parsed := ParsedUserAgent{DeviceType: normalizeDeviceType(input.DeviceType, e.cfg.DefaultDevice)}
	if e.cfg.Enabled {
		parsed = e.parser.Parse(ctx, userAgentForParsing)
	}
	device := e.deviceFromParsed(parsed, input)
	client := e.clientFromInput(input)
	geo := e.geoFromInput(ctx, input)
	userAgentHash := stringPtrIfNotEmpty(e.hasher.Hash(userAgentForParsing))
	ipHash := e.ipHash(input)
	fingerprintHash := strings.TrimSpace(input.DeviceFingerprintHash)
	if fingerprintHash == "" {
		fingerprintHash = e.hasher.Hash(input.DeviceFingerprint)
	}

	enriched := domain.EnrichedDeviceContext{
		UserAgent:             storedUserAgent,
		UserAgentHash:         userAgentHash,
		IPHash:                ipHash,
		IPVersion:             detectIPVersion(input.ClientIP, e.cfg.DefaultIPVersion),
		DeviceFingerprintHash: stringPtrIfNotEmpty(fingerprintHash),
		Device:                device,
		Client:                client,
		Geo:                   geo,
	}.Normalize()

	e.logger.DebugContext(ctx, "session.device.enriched",
		slog.String("session_id", strings.TrimSpace(input.SessionID)),
		slog.String("device_type", string(enriched.Device.Type)),
		slog.String("browser", safeString(enriched.Device.Browser)),
		slog.String("os", safeString(enriched.Device.OS)),
		slog.String("geo_country", safeString(enriched.Geo.Country)),
	)
	return enriched
}

func (e *DeviceEnricher) fallback(input DeviceEnrichmentInput) domain.EnrichedDeviceContext {
	deviceType := normalizeDeviceType(input.DeviceType, e.cfg.DefaultDevice)
	device := domain.Device{Type: deviceType}
	return domain.EnrichedDeviceContext{
		UserAgent: defaultUserAgent,
		IPHash:    e.ipHash(input),
		IPVersion: detectIPVersion(input.ClientIP, e.cfg.DefaultIPVersion),
		Device:    device,
		Client:    e.clientFromInput(input),
		Geo:       domain.Geo{Source: domain.GeoSourceUnknown},
	}
}

func (e *DeviceEnricher) deviceFromParsed(parsed ParsedUserAgent, input DeviceEnrichmentInput) domain.Device {
	deviceType := normalizeDeviceType(parsed.DeviceType, normalizeDeviceType(input.DeviceType, e.cfg.DefaultDevice))
	if parsed.IsBot {
		deviceType = domain.DeviceTypeBot
	}
	device := domain.Device{
		Type:           deviceType,
		Browser:        stringPtrIfNotEmpty(parsed.Browser),
		BrowserVersion: stringPtrIfNotEmpty(parsed.BrowserVersion),
		OS:             stringPtrIfNotEmpty(parsed.OS),
		OSVersion:      stringPtrIfNotEmpty(parsed.OSVersion),
		Model:          stringPtrIfNotEmpty(parsed.Model),
		Vendor:         stringPtrIfNotEmpty(parsed.Vendor),
		IsBot:          parsed.IsBot || deviceType == domain.DeviceTypeBot,
	}

	if platform := cleanClientHint(input.ClientHints.Platform); platform != "" {
		device.OS = &platform
	}
	if model := cleanClientHint(input.ClientHints.Model); model != "" {
		device.Model = &model
	}
	if mobile, ok := parseClientHintMobile(input.ClientHints.Mobile); ok && mobile && device.Type != domain.DeviceTypeTablet && !device.IsBot {
		device.Type = domain.DeviceTypeMobile
	}
	if brand, version := parseClientHintBrand(input.ClientHints.UA); brand != "" {
		device.Browser = &brand
		if version != "" {
			device.BrowserVersion = &version
		}
	}
	return device.Normalize()
}

func (e *DeviceEnricher) clientFromInput(input DeviceEnrichmentInput) domain.Client {
	channel := domain.Channel(strings.TrimSpace(string(input.Channel)))
	if !channel.Valid() {
		channel = e.cfg.DefaultChannel
	}
	locale := firstLanguage(input.Locale)
	timezone := strings.TrimSpace(input.Timezone)
	return domain.Client{
		Channel:        channel,
		Locale:         stringPtrIfNotEmpty(locale),
		Timezone:       stringPtrIfNotEmpty(timezone),
		ScreenWidth:    normalizeDimension(input.ScreenWidth),
		ScreenHeight:   normalizeDimension(input.ScreenHeight),
		ViewportWidth:  normalizeDimension(input.ViewportWidth),
		ViewportHeight: normalizeDimension(input.ViewportHeight),
	}.Normalize()
}

func (e *DeviceEnricher) geoFromInput(ctx context.Context, input DeviceEnrichmentInput) domain.Geo {
	if hinted := normalizeGeoHint(input.GeoHint); hinted != nil {
		return *hinted
	}
	if !e.cfg.Enabled {
		return domain.Geo{Source: domain.GeoSourceDisabled}.Normalize()
	}
	geoCtx, cancel := context.WithTimeout(ctx, e.cfg.GeoLookupTimeout)
	defer cancel()
	geo := e.geo.Resolve(geoCtx, input.ClientIP).Normalize()
	if geo.Source == "" {
		geo.Source = domain.GeoSourceUnknown
	}
	return geo
}

func (e *DeviceEnricher) ipHash(input DeviceEnrichmentInput) string {
	if hash := strings.TrimSpace(input.IPHash); hash != "" && net.ParseIP(hash) == nil {
		return hash
	}
	if hashed := e.hasher.Hash(input.ClientIP); hashed != "" {
		return hashed
	}
	return unavailableIPHash
}

func normalizeGeoHint(geo domain.Geo) *domain.Geo {
	normalized := geo.Normalize()
	if normalized.Country == nil && normalized.Region == nil && normalized.City == nil && normalized.Timezone == nil {
		return nil
	}
	normalized.Source = domain.GeoSourceHeader
	if normalized.Country != nil {
		country := strings.ToUpper(strings.TrimSpace(*normalized.Country))
		normalized.Country = &country
	}
	return &normalized
}

func normalizeDeviceType(value domain.DeviceType, fallback domain.DeviceType) domain.DeviceType {
	deviceType := domain.DeviceType(strings.TrimSpace(string(value)))
	if deviceType.Valid() {
		return deviceType
	}
	if fallback.Valid() {
		return fallback
	}
	return domain.DeviceTypeUnknown
}

func detectIPVersion(value string, fallback domain.IPVersion) domain.IPVersion {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		if fallback.Valid() {
			return fallback
		}
		return domain.IPVersionUnknown
	}
	if ip.To4() != nil {
		return domain.IPVersionIPv4
	}
	return domain.IPVersionIPv6
}

func trimMax(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}

func normalizeDimension(value int) int {
	if value <= 0 || value > 10000 {
		return 0
	}
	return value
}

func firstLanguage(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	first, _, _ := strings.Cut(value, ",")
	first = strings.TrimSpace(first)
	first, _, _ = strings.Cut(first, ";")
	first = strings.ReplaceAll(strings.TrimSpace(first), "_", "-")
	if len(first) > 64 {
		return ""
	}
	return first
}

func parseClientHintMobile(value string) (bool, bool) {
	value = strings.Trim(strings.TrimSpace(value), `"'`)
	switch value {
	case "?1", "1", "true", "True":
		return true, true
	case "?0", "0", "false", "False":
		return false, true
	default:
		return false, false
	}
}

func cleanClientHint(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	if len(value) > 128 {
		return ""
	}
	return value
}

func parseClientHintBrand(value string) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}
	type brandVersion struct {
		brand   string
		version string
	}
	brands := make([]brandVersion, 0)
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var brand string
		var version string
		for _, token := range strings.Split(part, ";") {
			key, raw, ok := strings.Cut(strings.TrimSpace(token), "=")
			if !ok {
				brand = cleanClientHint(token)
				continue
			}
			if strings.EqualFold(strings.TrimSpace(key), "v") {
				version = cleanClientHint(raw)
			}
		}
		if brand != "" && !strings.Contains(strings.ToLower(brand), "not") {
			brands = append(brands, brandVersion{brand: brand, version: version})
		}
	}
	for _, candidate := range brands {
		switch candidate.brand {
		case "Google Chrome", "Microsoft Edge", "Safari", "Firefox", "Opera", "Samsung Internet":
			return candidate.brand, candidate.version
		}
	}
	if len(brands) > 0 {
		return brands[0].brand, brands[0].version
	}
	return "", ""
}

func stringPtrIfNotEmpty(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func safeString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (c DeviceEnricherConfig) Validate() error {
	if c.UserAgentMaxLength <= 0 {
		return errors.New("SESSION_USER_AGENT_MAX_LENGTH must be greater than zero")
	}
	if c.GeoLookupTimeout <= 0 {
		return errors.New("SESSION_GEOIP_TIMEOUT must be greater than zero")
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

func (c DeviceEnricherConfig) withDefaults() DeviceEnricherConfig {
	if c.UserAgentMaxLength == 0 {
		c.UserAgentMaxLength = domain.DefaultMaxUserAgentLength
	}
	if c.GeoLookupTimeout == 0 {
		c.GeoLookupTimeout = defaultGeoLookupTimeout
	}
	if c.DefaultChannel == "" {
		c.DefaultChannel = domain.ChannelUnknown
	}
	if c.DefaultDevice == "" {
		c.DefaultDevice = domain.DeviceTypeUnknown
	}
	if c.DefaultIPVersion == "" {
		c.DefaultIPVersion = domain.IPVersionUnknown
	}
	return c
}

type unknownUserAgentParser struct{}

func (unknownUserAgentParser) Parse(ctx context.Context, userAgent string) ParsedUserAgent {
	return ParsedUserAgent{DeviceType: domain.DeviceTypeUnknown}
}

type disabledGeoResolver struct{}

func (disabledGeoResolver) Resolve(ctx context.Context, clientIP string) domain.Geo {
	return domain.Geo{Source: domain.GeoSourceDisabled}
}

type noopHasher struct{}

func (noopHasher) Hash(value string) string {
	return ""
}

func intFromProperty(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int8:
		return int(typed)
	case int16:
		return int(typed)
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case uint:
		return int(typed)
	case uint8:
		return int(typed)
	case uint16:
		return int(typed)
	case uint32:
		return int(typed)
	case uint64:
		if typed > uint64(^uint(0)>>1) {
			return 0
		}
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}
