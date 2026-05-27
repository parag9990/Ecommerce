package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type ChannelConfig struct {
	Enabled      bool
	ProviderName string
}

const (
	defaultMongoDatabase                 = "notification_db"
	defaultMongoTemplatesCollection      = "notification_templates"
	defaultMongoDeliveriesCollection     = "notification_deliveries"
	defaultMongoPreferencesCollection    = "notification_preferences"
	defaultMongoProviderEventsCollection = "provider_events"
	defaultGRPCAddress                   = ":9090"
	defaultStartupTimeout                = 5 * time.Second
	defaultShutdownTimeout               = 10 * time.Second
	defaultProviderTimeout               = 5 * time.Second
	defaultSMTPTLSMode                   = "starttls"
	defaultOrderEventsQueue              = "notification.order.events.v1"
	defaultPaymentEventsQueue            = "notification.payment.events.v1"
	defaultUserEventsQueue               = "notification.user.events.v1"
	defaultConsumerPrefetch              = 10
	defaultRetryMaxAttempts              = 4
	defaultRetryPrefetch                 = 10
	defaultRetryConfirmTimeout           = 5 * time.Second
	defaultRetryAttemptLease             = 30 * time.Second
	defaultAnalyticsHTTPAddress          = ":8081"
	defaultMetricsPath                   = "/metrics"
	defaultWebhookPathPrefix             = "/internal/provider-webhooks"
	defaultWebhookMaxBodyBytes           = 65536
	defaultWebhookReplayWindow           = 5 * time.Minute
)

var defaultRetryDelays = []time.Duration{30 * time.Second, 2 * time.Minute, 8 * time.Minute}

type MongoConfig struct {
	URI                      string
	Database                 string
	TemplatesCollection      string
	DeliveriesCollection     string
	PreferencesCollection    string
	ProviderEventsCollection string
}

type GRPCConfig struct {
	Address         string
	StartupTimeout  time.Duration
	ShutdownTimeout time.Duration
}

type RabbitMQConfig struct {
	Enabled            bool
	URL                string
	OrderEventsQueue   string
	PaymentEventsQueue string
	UserEventsQueue    string
	Prefetch           int
}

type RetryConfig struct {
	MaxAttempts           int
	Delays                []time.Duration
	Prefetch              int
	PublishConfirmTimeout time.Duration
	AttemptLease          time.Duration
	DeliveryEncryptionKey string
}

type SMTPConfig struct {
	Host     string
	Port     int
	From     string
	Username string
	Password string
	TLSMode  string
	Timeout  time.Duration
}

type SMSHTTPConfig struct {
	Endpoint    string
	BearerToken string
	Timeout     time.Duration
}

type AnalyticsConfig struct {
	MetricsEnabled        bool
	HTTPAddress           string
	MetricsPath           string
	WebhooksEnabled       bool
	WebhookPathPrefix     string
	WebhookMaxBodyBytes   int64
	WebhookReplayWindow   time.Duration
	WebhookSigningSecrets map[domain.Channel]string
	OpenTrackingEnabled   map[domain.Channel]bool
}

type Config struct {
	Mongo        MongoConfig
	GRPC         GRPCConfig
	RabbitMQ     RabbitMQConfig
	Retry        RetryConfig
	Email        ChannelConfig
	SMS          ChannelConfig
	Push         ChannelConfig
	WhatsAppLike ChannelConfig
	EmailSMTP    SMTPConfig
	SMSHTTP      SMSHTTPConfig
	Analytics    AnalyticsConfig
}

type LookupEnv func(key string) (string, bool)

func Load() (Config, error) {
	return LoadFromEnv(os.LookupEnv)
}

func LoadFromEnv(lookup LookupEnv) (Config, error) {
	if lookup == nil {
		return Config{}, errors.New("environment lookup is required")
	}

	mongo, err := loadMongo(lookup)
	if err != nil {
		return Config{}, err
	}
	grpc, err := loadGRPC(lookup)
	if err != nil {
		return Config{}, err
	}
	rabbitMQ, err := loadRabbitMQ(lookup)
	if err != nil {
		return Config{}, err
	}
	retry, err := loadRetry(lookup)
	if err != nil {
		return Config{}, err
	}
	email, err := loadChannel(lookup, "NOTIFICATION_EMAIL_ENABLED", "NOTIFICATION_EMAIL_PROVIDER")
	if err != nil {
		return Config{}, err
	}
	sms, err := loadChannel(lookup, "NOTIFICATION_SMS_ENABLED", "NOTIFICATION_SMS_PROVIDER")
	if err != nil {
		return Config{}, err
	}
	push, err := loadChannel(lookup, "NOTIFICATION_PUSH_ENABLED", "NOTIFICATION_PUSH_PROVIDER")
	if err != nil {
		return Config{}, err
	}
	whatsAppLike, err := loadChannel(lookup, "NOTIFICATION_WHATSAPP_LIKE_ENABLED", "NOTIFICATION_WHATSAPP_LIKE_PROVIDER")
	if err != nil {
		return Config{}, err
	}

	emailSMTP, err := loadSMTP(lookup)
	if err != nil {
		return Config{}, err
	}
	smsHTTP, err := loadSMSHTTP(lookup)
	if err != nil {
		return Config{}, err
	}
	analytics, err := loadAnalytics(lookup)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Mongo:        mongo,
		GRPC:         grpc,
		RabbitMQ:     rabbitMQ,
		Retry:        retry,
		Email:        email,
		SMS:          sms,
		Push:         push,
		WhatsAppLike: whatsAppLike,
		EmailSMTP:    emailSMTP,
		SMSHTTP:      smsHTTP,
		Analytics:    analytics,
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if err := c.Mongo.Validate(); err != nil {
		return err
	}
	if err := c.GRPC.Validate(); err != nil {
		return err
	}
	if err := c.RabbitMQ.Validate(); err != nil {
		return err
	}
	if err := c.Retry.Validate(c.RabbitMQ.Enabled); err != nil {
		return err
	}
	for _, channel := range domain.SupportedChannels() {
		settings := c.channel(channel)
		if settings.Enabled && strings.TrimSpace(settings.ProviderName) == "" {
			return fmt.Errorf("provider name is required when notification channel %q is enabled", channel)
		}
	}
	if c.Email.Enabled {
		if err := c.EmailSMTP.Validate(); err != nil {
			return fmt.Errorf("email SMTP configuration: %w", err)
		}
	}
	if c.SMS.Enabled {
		if err := c.SMSHTTP.Validate(); err != nil {
			return fmt.Errorf("SMS HTTP configuration: %w", err)
		}
	}
	if c.RabbitMQ.Enabled && !c.Email.Enabled {
		return errors.New("notification event consumer requires the email channel to be enabled")
	}
	if err := c.Analytics.Validate(c); err != nil {
		return err
	}
	return nil
}

func (c MongoConfig) Validate() error {
	if strings.TrimSpace(c.URI) == "" {
		return errors.New("NOTIFICATION_MONGO_URI is required")
	}
	if err := validMongoDatabaseName(c.Database); err != nil {
		return fmt.Errorf("NOTIFICATION_MONGO_DATABASE: %w", err)
	}
	if err := validMongoCollectionName(c.TemplatesCollection); err != nil {
		return fmt.Errorf("NOTIFICATION_MONGO_TEMPLATES_COLLECTION: %w", err)
	}
	if err := validMongoCollectionName(c.DeliveriesCollection); err != nil {
		return fmt.Errorf("NOTIFICATION_MONGO_DELIVERIES_COLLECTION: %w", err)
	}
	if err := validMongoCollectionName(c.PreferencesCollection); err != nil {
		return fmt.Errorf("NOTIFICATION_MONGO_PREFERENCES_COLLECTION: %w", err)
	}
	if err := validMongoCollectionName(c.ProviderEventsCollection); err != nil {
		return fmt.Errorf("NOTIFICATION_MONGO_PROVIDER_EVENTS_COLLECTION: %w", err)
	}
	collections := map[string]struct{}{}
	for _, collection := range []string{c.TemplatesCollection, c.DeliveriesCollection, c.PreferencesCollection, c.ProviderEventsCollection} {
		if _, exists := collections[collection]; exists {
			return errors.New("Mongo notification collections must be distinct")
		}
		collections[collection] = struct{}{}
	}
	return nil
}

func (c AnalyticsConfig) Validate(service Config) error {
	if c.WebhooksEnabled && !c.MetricsEnabled {
		return errors.New("NOTIFICATION_PROVIDER_WEBHOOKS_ENABLED requires NOTIFICATION_METRICS_ENABLED")
	}
	if !c.MetricsEnabled && !c.WebhooksEnabled {
		for channel, enabled := range c.OpenTrackingEnabled {
			if enabled {
				return fmt.Errorf("open tracking for channel %q requires notification webhooks", channel)
			}
		}
		return nil
	}
	if strings.TrimSpace(c.HTTPAddress) == "" {
		return errors.New("NOTIFICATION_ANALYTICS_HTTP_ADDRESS is required when analytics HTTP is enabled")
	}
	if !validHTTPPath(c.MetricsPath) || !validHTTPPath(c.WebhookPathPrefix) {
		return errors.New("notification metrics and webhook paths must be absolute HTTP paths")
	}
	if strings.TrimRight(c.WebhookPathPrefix, "/") == strings.TrimRight(c.MetricsPath, "/") {
		return errors.New("notification metrics and webhook paths must be different")
	}
	if c.WebhookMaxBodyBytes < 1 || c.WebhookReplayWindow <= 0 {
		return errors.New("notification webhook body limit and replay window must be positive")
	}
	if !c.WebhooksEnabled {
		for channel, enabled := range c.OpenTrackingEnabled {
			if enabled {
				return fmt.Errorf("open tracking for channel %q requires notification webhooks", channel)
			}
		}
		return nil
	}
	configuredEndpoints := 0
	for channel, secret := range c.WebhookSigningSecrets {
		if strings.TrimSpace(secret) == "" {
			continue
		}
		if !service.ChannelEnabled(channel) {
			return fmt.Errorf("webhook signing secret configured for disabled notification channel %q", channel)
		}
		configuredEndpoints++
	}
	if configuredEndpoints == 0 {
		return errors.New("notification webhooks require a signing secret for an enabled channel")
	}
	for channel, enabled := range c.OpenTrackingEnabled {
		if !enabled {
			continue
		}
		if channel != domain.ChannelEmail {
			return fmt.Errorf("open tracking is not enabled for notification channel %q", channel)
		}
		if strings.TrimSpace(c.WebhookSigningSecrets[channel]) == "" {
			return fmt.Errorf("open tracking for channel %q requires a webhook signing secret", channel)
		}
	}
	return nil
}

func (c AnalyticsConfig) SigningSecret(channel domain.Channel) string {
	return strings.TrimSpace(c.WebhookSigningSecrets[channel])
}

func (c AnalyticsConfig) AllowsOpenTracking(channel domain.Channel) bool {
	return c.OpenTrackingEnabled[channel]
}

func (c GRPCConfig) Validate() error {
	if strings.TrimSpace(c.Address) == "" {
		return errors.New("NOTIFICATION_GRPC_ADDRESS is required")
	}
	if c.StartupTimeout <= 0 {
		return errors.New("NOTIFICATION_STARTUP_TIMEOUT must be greater than zero")
	}
	if c.ShutdownTimeout <= 0 {
		return errors.New("NOTIFICATION_SHUTDOWN_TIMEOUT must be greater than zero")
	}
	return nil
}

func (c RabbitMQConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	parsed, err := url.Parse(strings.TrimSpace(c.URL))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "amqp" && parsed.Scheme != "amqps") {
		return errors.New("NOTIFICATION_RABBITMQ_URL must be an absolute amqp or amqps URL")
	}
	queues := []string{c.OrderEventsQueue, c.PaymentEventsQueue, c.UserEventsQueue}
	seen := make(map[string]struct{}, len(queues))
	for _, queue := range queues {
		queue = strings.TrimSpace(queue)
		if queue == "" {
			return errors.New("notification event queue names are required")
		}
		if _, exists := seen[queue]; exists {
			return errors.New("notification event queue names must be distinct")
		}
		seen[queue] = struct{}{}
	}
	if c.Prefetch < 1 {
		return errors.New("NOTIFICATION_CONSUMER_PREFETCH must be greater than zero")
	}
	return nil
}

func (c RetryConfig) Validate(enabled bool) error {
	if !enabled {
		return nil
	}
	if c.MaxAttempts < 1 || c.MaxAttempts > 10 {
		return errors.New("NOTIFICATION_RETRY_MAX_ATTEMPTS must be between 1 and 10")
	}
	if len(c.Delays) != c.MaxAttempts-1 {
		return errors.New("notification retry delay count must equal max attempts minus one")
	}
	for index, delay := range c.Delays {
		if delay <= 0 {
			return errors.New("notification retry delays must be positive")
		}
		if index > 0 && delay <= c.Delays[index-1] {
			return errors.New("notification retry delays must be strictly increasing")
		}
	}
	if c.Prefetch < 1 {
		return errors.New("NOTIFICATION_RETRY_PREFETCH must be greater than zero")
	}
	if c.PublishConfirmTimeout <= 0 {
		return errors.New("NOTIFICATION_RETRY_PUBLISH_CONFIRM_TIMEOUT must be greater than zero")
	}
	if c.AttemptLease <= c.PublishConfirmTimeout {
		return errors.New("NOTIFICATION_RETRY_ATTEMPT_LEASE must exceed publish confirmation timeout")
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(c.DeliveryEncryptionKey))
	if err != nil || len(key) != 32 {
		return errors.New("NOTIFICATION_DELIVERY_ENCRYPTION_KEY must be base64 encoding of 32 bytes when event consumption is enabled")
	}
	return nil
}

func (c SMTPConfig) Validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return errors.New("NOTIFICATION_EMAIL_SMTP_HOST is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("NOTIFICATION_EMAIL_SMTP_PORT must be a valid TCP port")
	}
	address, err := mail.ParseAddress(strings.TrimSpace(c.From))
	if err != nil || address.Address != strings.TrimSpace(c.From) {
		return errors.New("NOTIFICATION_EMAIL_FROM must be a valid email address")
	}
	switch c.TLSMode {
	case "tls", "starttls", "none":
	default:
		return errors.New("NOTIFICATION_EMAIL_SMTP_TLS_MODE must be tls, starttls, or none")
	}
	if (c.Username == "") != (c.Password == "") {
		return errors.New("SMTP username and password must be configured together")
	}
	if c.Username != "" && c.TLSMode == "none" {
		return errors.New("SMTP credentials require TLS")
	}
	if c.Timeout <= 0 {
		return errors.New("NOTIFICATION_EMAIL_TIMEOUT must be greater than zero")
	}
	return nil
}

func (c SMSHTTPConfig) Validate() error {
	endpoint := strings.TrimSpace(c.Endpoint)
	parsed, err := url.ParseRequestURI(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("NOTIFICATION_SMS_HTTP_ENDPOINT must be an absolute URL")
	}
	if parsed.Scheme != "https" && !loopbackHost(parsed.Hostname()) {
		return errors.New("NOTIFICATION_SMS_HTTP_ENDPOINT must use HTTPS outside local development")
	}
	if strings.TrimSpace(c.BearerToken) == "" && !loopbackHost(parsed.Hostname()) {
		return errors.New("NOTIFICATION_SMS_HTTP_BEARER_TOKEN is required outside local development")
	}
	if c.Timeout <= 0 {
		return errors.New("NOTIFICATION_SMS_TIMEOUT must be greater than zero")
	}
	return nil
}

func (c Config) ChannelEnabled(channel domain.Channel) bool {
	return c.channel(channel).Enabled
}

func (c Config) ProviderName(channel domain.Channel) string {
	return strings.TrimSpace(c.channel(channel).ProviderName)
}

func (c Config) channel(channel domain.Channel) ChannelConfig {
	switch channel {
	case domain.ChannelEmail:
		return c.Email
	case domain.ChannelSMS:
		return c.SMS
	case domain.ChannelPush:
		return c.Push
	case domain.ChannelWhatsAppLike:
		return c.WhatsAppLike
	default:
		return ChannelConfig{}
	}
}

func loadMongo(lookup LookupEnv) (MongoConfig, error) {
	uri, _ := lookup("NOTIFICATION_MONGO_URI")
	database, _ := lookup("NOTIFICATION_MONGO_DATABASE")
	templates, _ := lookup("NOTIFICATION_MONGO_TEMPLATES_COLLECTION")
	deliveries, _ := lookup("NOTIFICATION_MONGO_DELIVERIES_COLLECTION")
	preferences, _ := lookup("NOTIFICATION_MONGO_PREFERENCES_COLLECTION")
	providerEvents, _ := lookup("NOTIFICATION_MONGO_PROVIDER_EVENTS_COLLECTION")

	cfg := MongoConfig{
		URI:                      strings.TrimSpace(uri),
		Database:                 defaultValue(database, defaultMongoDatabase),
		TemplatesCollection:      defaultValue(templates, defaultMongoTemplatesCollection),
		DeliveriesCollection:     defaultValue(deliveries, defaultMongoDeliveriesCollection),
		PreferencesCollection:    defaultValue(preferences, defaultMongoPreferencesCollection),
		ProviderEventsCollection: defaultValue(providerEvents, defaultMongoProviderEventsCollection),
	}
	if err := cfg.Validate(); err != nil {
		return MongoConfig{}, err
	}
	return cfg, nil
}

func loadGRPC(lookup LookupEnv) (GRPCConfig, error) {
	address, _ := lookup("NOTIFICATION_GRPC_ADDRESS")
	startup, err := durationValue(lookup, "NOTIFICATION_STARTUP_TIMEOUT", defaultStartupTimeout)
	if err != nil {
		return GRPCConfig{}, err
	}
	shutdown, err := durationValue(lookup, "NOTIFICATION_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	if err != nil {
		return GRPCConfig{}, err
	}
	return GRPCConfig{
		Address:         defaultValue(address, defaultGRPCAddress),
		StartupTimeout:  startup,
		ShutdownTimeout: shutdown,
	}, nil
}

func loadRabbitMQ(lookup LookupEnv) (RabbitMQConfig, error) {
	enabled, err := booleanValue(lookup, "NOTIFICATION_EVENT_CONSUMER_ENABLED")
	if err != nil {
		return RabbitMQConfig{}, err
	}
	urlValue, _ := lookup("NOTIFICATION_RABBITMQ_URL")
	orderQueue, _ := lookup("NOTIFICATION_ORDER_EVENTS_QUEUE")
	paymentQueue, _ := lookup("NOTIFICATION_PAYMENT_EVENTS_QUEUE")
	userQueue, _ := lookup("NOTIFICATION_USER_EVENTS_QUEUE")
	prefetchValue, _ := lookup("NOTIFICATION_CONSUMER_PREFETCH")
	prefetch := defaultConsumerPrefetch
	if strings.TrimSpace(prefetchValue) != "" {
		prefetch, err = strconv.Atoi(strings.TrimSpace(prefetchValue))
		if err != nil {
			return RabbitMQConfig{}, fmt.Errorf("NOTIFICATION_CONSUMER_PREFETCH must be an integer: %w", err)
		}
	}
	return RabbitMQConfig{
		Enabled:            enabled,
		URL:                strings.TrimSpace(urlValue),
		OrderEventsQueue:   defaultValue(orderQueue, defaultOrderEventsQueue),
		PaymentEventsQueue: defaultValue(paymentQueue, defaultPaymentEventsQueue),
		UserEventsQueue:    defaultValue(userQueue, defaultUserEventsQueue),
		Prefetch:           prefetch,
	}, nil
}

func loadRetry(lookup LookupEnv) (RetryConfig, error) {
	maxAttempts, err := integerValue(lookup, "NOTIFICATION_RETRY_MAX_ATTEMPTS", defaultRetryMaxAttempts)
	if err != nil {
		return RetryConfig{}, err
	}
	if maxAttempts < 1 || maxAttempts > 10 {
		return RetryConfig{}, errors.New("NOTIFICATION_RETRY_MAX_ATTEMPTS must be between 1 and 10")
	}
	delays := make([]time.Duration, 0, maxAttempts-1)
	for tier := 1; tier < maxAttempts; tier++ {
		key := fmt.Sprintf("NOTIFICATION_RETRY_DELAY_%d", tier)
		fallback := time.Duration(0)
		if tier <= len(defaultRetryDelays) {
			fallback = defaultRetryDelays[tier-1]
		} else if value, exists := lookup(key); !exists || strings.TrimSpace(value) == "" {
			return RetryConfig{}, fmt.Errorf("%s is required for configured max attempts", key)
		}
		delay, err := durationValue(lookup, key, fallback)
		if err != nil {
			return RetryConfig{}, err
		}
		delays = append(delays, delay)
	}
	prefetch, err := integerValue(lookup, "NOTIFICATION_RETRY_PREFETCH", defaultRetryPrefetch)
	if err != nil {
		return RetryConfig{}, err
	}
	confirmTimeout, err := durationValue(lookup, "NOTIFICATION_RETRY_PUBLISH_CONFIRM_TIMEOUT", defaultRetryConfirmTimeout)
	if err != nil {
		return RetryConfig{}, err
	}
	attemptLease, err := durationValue(lookup, "NOTIFICATION_RETRY_ATTEMPT_LEASE", defaultRetryAttemptLease)
	if err != nil {
		return RetryConfig{}, err
	}
	encryptionKey, _ := lookup("NOTIFICATION_DELIVERY_ENCRYPTION_KEY")
	return RetryConfig{
		MaxAttempts: maxAttempts, Delays: delays, Prefetch: prefetch,
		PublishConfirmTimeout: confirmTimeout, AttemptLease: attemptLease,
		DeliveryEncryptionKey: strings.TrimSpace(encryptionKey),
	}, nil
}

func loadSMTP(lookup LookupEnv) (SMTPConfig, error) {
	host, _ := lookup("NOTIFICATION_EMAIL_SMTP_HOST")
	portValue, _ := lookup("NOTIFICATION_EMAIL_SMTP_PORT")
	from, _ := lookup("NOTIFICATION_EMAIL_FROM")
	username, _ := lookup("NOTIFICATION_EMAIL_SMTP_USERNAME")
	password, _ := lookup("NOTIFICATION_EMAIL_SMTP_PASSWORD")
	tlsMode, _ := lookup("NOTIFICATION_EMAIL_SMTP_TLS_MODE")
	timeout, err := durationValue(lookup, "NOTIFICATION_EMAIL_TIMEOUT", defaultProviderTimeout)
	if err != nil {
		return SMTPConfig{}, err
	}
	port := 0
	if strings.TrimSpace(portValue) != "" {
		port, err = strconv.Atoi(strings.TrimSpace(portValue))
		if err != nil {
			return SMTPConfig{}, fmt.Errorf("NOTIFICATION_EMAIL_SMTP_PORT must be an integer: %w", err)
		}
	}
	return SMTPConfig{
		Host:     strings.TrimSpace(host),
		Port:     port,
		From:     strings.TrimSpace(from),
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
		TLSMode:  defaultValue(tlsMode, defaultSMTPTLSMode),
		Timeout:  timeout,
	}, nil
}

func loadSMSHTTP(lookup LookupEnv) (SMSHTTPConfig, error) {
	endpoint, _ := lookup("NOTIFICATION_SMS_HTTP_ENDPOINT")
	token, _ := lookup("NOTIFICATION_SMS_HTTP_BEARER_TOKEN")
	timeout, err := durationValue(lookup, "NOTIFICATION_SMS_TIMEOUT", defaultProviderTimeout)
	if err != nil {
		return SMSHTTPConfig{}, err
	}
	return SMSHTTPConfig{
		Endpoint:    strings.TrimSpace(endpoint),
		BearerToken: strings.TrimSpace(token),
		Timeout:     timeout,
	}, nil
}

func loadAnalytics(lookup LookupEnv) (AnalyticsConfig, error) {
	metricsEnabled, err := booleanValue(lookup, "NOTIFICATION_METRICS_ENABLED")
	if err != nil {
		return AnalyticsConfig{}, err
	}
	webhooksEnabled, err := booleanValue(lookup, "NOTIFICATION_PROVIDER_WEBHOOKS_ENABLED")
	if err != nil {
		return AnalyticsConfig{}, err
	}
	httpAddress, _ := lookup("NOTIFICATION_ANALYTICS_HTTP_ADDRESS")
	metricsPath, _ := lookup("NOTIFICATION_METRICS_PATH")
	webhookPath, _ := lookup("NOTIFICATION_WEBHOOK_PATH_PREFIX")
	maxBodyBytes, err := integerValue(lookup, "NOTIFICATION_WEBHOOK_MAX_BODY_BYTES", defaultWebhookMaxBodyBytes)
	if err != nil {
		return AnalyticsConfig{}, err
	}
	replaySeconds, err := integerValue(lookup, "NOTIFICATION_WEBHOOK_REPLAY_WINDOW_SECONDS", int(defaultWebhookReplayWindow/time.Second))
	if err != nil {
		return AnalyticsConfig{}, err
	}
	secrets := make(map[domain.Channel]string, len(domain.SupportedChannels()))
	openTracking := make(map[domain.Channel]bool, len(domain.SupportedChannels()))
	keys := map[domain.Channel]string{
		domain.ChannelEmail:        "EMAIL",
		domain.ChannelSMS:          "SMS",
		domain.ChannelPush:         "PUSH",
		domain.ChannelWhatsAppLike: "WHATSAPP_LIKE",
	}
	for channel, key := range keys {
		secret, _ := lookup("NOTIFICATION_" + key + "_WEBHOOK_SIGNING_SECRET")
		secrets[channel] = strings.TrimSpace(secret)
		enabled, parseErr := booleanValue(lookup, "NOTIFICATION_"+key+"_OPEN_TRACKING_ENABLED")
		if parseErr != nil {
			return AnalyticsConfig{}, parseErr
		}
		openTracking[channel] = enabled
	}
	return AnalyticsConfig{
		MetricsEnabled: metricsEnabled, HTTPAddress: defaultValue(httpAddress, defaultAnalyticsHTTPAddress),
		MetricsPath: defaultValue(metricsPath, defaultMetricsPath), WebhooksEnabled: webhooksEnabled,
		WebhookPathPrefix:   strings.TrimRight(defaultValue(webhookPath, defaultWebhookPathPrefix), "/"),
		WebhookMaxBodyBytes: int64(maxBodyBytes), WebhookReplayWindow: time.Duration(replaySeconds) * time.Second,
		WebhookSigningSecrets: secrets, OpenTrackingEnabled: openTracking,
	}, nil
}

func loadChannel(lookup LookupEnv, enabledKey, providerKey string) (ChannelConfig, error) {
	enabled, err := booleanValue(lookup, enabledKey)
	if err != nil {
		return ChannelConfig{}, err
	}
	provider, _ := lookup(providerKey)
	return ChannelConfig{
		Enabled:      enabled,
		ProviderName: strings.TrimSpace(provider),
	}, nil
}

func booleanValue(lookup LookupEnv, key string) (bool, error) {
	value, _ := lookup(key)
	if strings.TrimSpace(value) == "" {
		return false, nil
	}
	enabled, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return enabled, nil
}

func integerValue(lookup LookupEnv, key string, fallback int) (int, error) {
	value, _ := lookup(key)
	if strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	number, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return number, nil
}

func defaultValue(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func durationValue(lookup LookupEnv, key string, fallback time.Duration) (time.Duration, error) {
	value, _ := lookup(key)
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return duration, nil
}

func loopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func validHTTPPath(path string) bool {
	path = strings.TrimSpace(path)
	return strings.HasPrefix(path, "/") && !strings.ContainsAny(path, "?#")
}

func validMongoDatabaseName(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("database name is required")
	}
	if strings.ContainsAny(value, `/\."$*<>:|?`) {
		return errors.New("database name contains an invalid character")
	}
	return nil
}

func validMongoCollectionName(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("collection name is required")
	}
	if strings.ContainsRune(value, '$') || strings.HasPrefix(value, "system.") {
		return errors.New("collection name is reserved or contains an invalid character")
	}
	return nil
}
