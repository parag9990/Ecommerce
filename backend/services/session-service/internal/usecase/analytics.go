package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

const (
	defaultAnalyticsLiveWindow          = 5 * time.Minute
	defaultAnalyticsSessionLookback     = 24 * time.Hour
	defaultAnalyticsMaxSessionRangeDays = 90
	defaultAnalyticsDefaultPageSize     = 50
	defaultAnalyticsMaxPageSize         = 100
	defaultAnalyticsMaxFunnelRangeDays  = 90
	defaultAnalyticsRawFallbackRange    = 24 * time.Hour
	defaultAnalyticsMinFunnelSteps      = 2
	defaultAnalyticsMaxFunnelSteps      = 8
	defaultAnalyticsLiveCacheTTL        = 10 * time.Second
	defaultAnalyticsSessionsCacheTTL    = 30 * time.Second
	defaultAnalyticsFunnelsCacheTTL     = 2 * time.Minute
)

type AnalyticsConfig struct {
	LiveWindow          time.Duration
	SessionLookback     time.Duration
	MaxSessionRangeDays int
	DefaultPageSize     int
	MaxPageSize         int
	MaxFunnelRangeDays  int
	RawFallbackRange    time.Duration
	MinFunnelSteps      int
	MaxFunnelSteps      int
	DefaultFunnelMetric domain.AnalyticsMetric
	LiveCacheTTL        time.Duration
	SessionsCacheTTL    time.Duration
	FunnelsCacheTTL     time.Duration
}

type GetLiveMetricsInput struct {
	From time.Time
	To   time.Time
}

type ListSessionsInput struct {
	UserID      string
	AnonymousID string
	From        time.Time
	To          time.Time
	Page        int
	PageSize    int
	DeviceType  string
	Channel     string
	Status      string
}

type SessionListOutput struct {
	Sessions []domain.Session
	Total    int64
	Page     int
	PageSize int
	HasNext  bool
}

type GetFunnelReportInput struct {
	Metric     string
	From       time.Time
	To         time.Time
	Steps      []string
	DeviceType string
	Channel    string
	Country    string
	Campaign   string
}

type FunnelReportOutput struct {
	Steps             []domain.FunnelStep
	OverallConversion float64
	Source            string
	From              time.Time
	To                time.Time
}

type AnalyticsUsecase struct {
	live       LiveMetricsRepository
	sessions   AnalyticsSessionRepository
	aggregates AnalyticsAggregateRepository
	cfg        AnalyticsConfig
	logger     *slog.Logger
	clock      Clock
	cache      *analyticsCache
}

func NewAnalyticsUsecase(
	live LiveMetricsRepository,
	sessions AnalyticsSessionRepository,
	aggregates AnalyticsAggregateRepository,
	cfg AnalyticsConfig,
	logger *slog.Logger,
) (*AnalyticsUsecase, error) {
	if live == nil {
		return nil, errors.New("live metrics repository is required")
	}
	if sessions == nil {
		return nil, errors.New("analytics session repository is required")
	}
	if aggregates == nil {
		return nil, errors.New("analytics aggregate repository is required")
	}
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &AnalyticsUsecase{
		live:       live,
		sessions:   sessions,
		aggregates: aggregates,
		cfg:        cfg,
		logger:     logger,
		clock:      realClock{},
		cache:      newAnalyticsCache(),
	}, nil
}

func (u *AnalyticsUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *AnalyticsUsecase) GetLiveMetrics(ctx context.Context, input GetLiveMetricsInput) (domain.LiveMetrics, error) {
	if err := ctx.Err(); err != nil {
		return domain.LiveMetrics{}, err
	}
	now := u.clock.Now().UTC()
	if err := validateOptionalDateRange(input.From, input.To, 0); err != nil {
		return domain.LiveMetrics{}, err
	}
	cacheKey := analyticsCacheKey("live", map[string]string{
		"window": u.cfg.LiveWindow.String(),
	})
	if cached, ok := u.cache.Get(cacheKey, now); ok {
		if metrics, ok := cached.(domain.LiveMetrics); ok {
			return metrics, nil
		}
	}

	started := time.Now()
	activeUsers, err := u.live.CountActiveUsers(ctx, u.cfg.LiveWindow)
	if err != nil {
		return domain.LiveMetrics{}, fmt.Errorf("%w: count active users: %w", ErrAnalyticsStorageUnavailable, err)
	}
	activeSessions, err := u.live.CountActiveSessions(ctx, u.cfg.LiveWindow)
	if err != nil {
		return domain.LiveMetrics{}, fmt.Errorf("%w: count active sessions: %w", ErrAnalyticsStorageUnavailable, err)
	}
	eventsPerMinute, err := u.live.EventsPerMinute(ctx, u.cfg.LiveWindow)
	if err != nil {
		return domain.LiveMetrics{}, fmt.Errorf("%w: count events per minute: %w", ErrAnalyticsStorageUnavailable, err)
	}
	out := domain.LiveMetrics{
		ActiveUsers:     activeUsers,
		ActiveSessions:  activeSessions,
		EventsPerMinute: eventsPerMinute,
		WindowSeconds:   int64(u.cfg.LiveWindow.Seconds()),
		MeasuredAt:      now,
	}
	u.cache.Set(cacheKey, out, u.cfg.LiveCacheTTL, now)
	u.logger.InfoContext(ctx, "session.analytics.live_metrics_fetched",
		slog.Int64("active_users", out.ActiveUsers),
		slog.Int64("active_sessions", out.ActiveSessions),
		slog.Int64("duration_ms", time.Since(started).Milliseconds()),
	)
	return out, nil
}

func (u *AnalyticsUsecase) ListSessions(ctx context.Context, input ListSessionsInput) (SessionListOutput, error) {
	if err := ctx.Err(); err != nil {
		return SessionListOutput{}, err
	}
	now := u.clock.Now().UTC()
	filter, err := u.normalizeSessionListInput(input, now)
	if err != nil {
		return SessionListOutput{}, err
	}
	cacheKey := analyticsCacheKey("sessions", sessionFilterCacheParams(filter))
	if cached, ok := u.cache.Get(cacheKey, now); ok {
		if out, ok := cached.(SessionListOutput); ok {
			return cloneSessionListOutput(out), nil
		}
	}

	started := time.Now()
	sessions, total, err := u.sessions.ListSessions(ctx, filter)
	if err != nil {
		return SessionListOutput{}, fmt.Errorf("%w: list sessions: %w", ErrAnalyticsStorageUnavailable, err)
	}
	out := SessionListOutput{
		Sessions: sessions,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		HasNext:  int64(filter.Page*filter.PageSize) < total,
	}
	u.cache.Set(cacheKey, cloneSessionListOutput(out), u.cfg.SessionsCacheTTL, now)
	u.logger.InfoContext(ctx, "session.analytics.sessions_listed",
		slog.Int("page", out.Page),
		slog.Int("page_size", out.PageSize),
		slog.Int("sessions_count", len(out.Sessions)),
		slog.Int64("total", out.Total),
		slog.Int64("duration_ms", time.Since(started).Milliseconds()),
	)
	return out, nil
}

func (u *AnalyticsUsecase) GetFunnelReport(ctx context.Context, input GetFunnelReportInput) (FunnelReportOutput, error) {
	if err := ctx.Err(); err != nil {
		return FunnelReportOutput{}, err
	}
	now := u.clock.Now().UTC()
	filter, err := u.normalizeFunnelInput(input)
	if err != nil {
		return FunnelReportOutput{}, err
	}
	cacheKey := analyticsCacheKey("funnels", funnelFilterCacheParams(filter))
	if cached, ok := u.cache.Get(cacheKey, now); ok {
		if out, ok := cached.(FunnelReportOutput); ok {
			return cloneFunnelReportOutput(out), nil
		}
	}

	started := time.Now()
	steps, err := u.aggregates.GetFunnelAggregate(ctx, filter)
	if err != nil {
		return FunnelReportOutput{}, fmt.Errorf("%w: get funnel aggregate: %w", ErrAnalyticsStorageUnavailable, err)
	}
	source := "mongo_aggregate"
	if len(steps) == 0 {
		if filter.To.Sub(filter.From) > u.cfg.RawFallbackRange {
			return FunnelReportOutput{}, fmt.Errorf("%w: requested range has no aggregate and exceeds raw fallback window", ErrAnalyticsAggregateNotReady)
		}
		steps, err = u.aggregates.BuildFunnelFromRawEvents(ctx, filter)
		if err != nil {
			return FunnelReportOutput{}, fmt.Errorf("%w: build funnel from raw events: %w", ErrAnalyticsStorageUnavailable, err)
		}
		source = "raw_fallback"
	}
	steps = domain.CalculateFunnelRates(steps)
	out := FunnelReportOutput{
		Steps:             steps,
		OverallConversion: domain.OverallFunnelConversion(steps),
		Source:            source,
		From:              filter.From,
		To:                filter.To,
	}
	u.cache.Set(cacheKey, cloneFunnelReportOutput(out), u.cfg.FunnelsCacheTTL, now)
	u.logger.InfoContext(ctx, "session.analytics.funnel_report_fetched",
		slog.String("source", source),
		slog.Int("steps_count", len(out.Steps)),
		slog.Float64("overall_conversion", out.OverallConversion),
		slog.Int64("duration_ms", time.Since(started).Milliseconds()),
	)
	return out, nil
}

func (u *AnalyticsUsecase) normalizeSessionListInput(input ListSessionsInput, now time.Time) (domain.SessionListFilter, error) {
	page := input.Page
	if page == 0 {
		page = 1
	}
	if page < 0 {
		return domain.SessionListFilter{}, fmt.Errorf("%w: page must be greater than zero", ErrInvalidSessionInput)
	}
	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = u.cfg.DefaultPageSize
	}
	if pageSize <= 0 {
		return domain.SessionListFilter{}, fmt.Errorf("%w: page_size must be greater than zero", ErrInvalidSessionInput)
	}
	if pageSize > u.cfg.MaxPageSize {
		return domain.SessionListFilter{}, fmt.Errorf("%w: page_size cannot exceed %d", ErrInvalidSessionInput, u.cfg.MaxPageSize)
	}

	from := input.From.UTC()
	to := input.To.UTC()
	if from.IsZero() && to.IsZero() {
		to = now
		from = now.Add(-u.cfg.SessionLookback)
	}
	if from.IsZero() {
		from = to.Add(-u.cfg.SessionLookback)
	}
	if to.IsZero() {
		to = now
	}
	if err := validateRequiredDateRange(from, to, daysDuration(u.cfg.MaxSessionRangeDays)); err != nil {
		return domain.SessionListFilter{}, err
	}

	filter := domain.SessionListFilter{
		UserID:      strings.TrimSpace(input.UserID),
		AnonymousID: strings.TrimSpace(input.AnonymousID),
		From:        from,
		To:          to,
		Page:        page,
		PageSize:    pageSize,
		DeviceType:  domain.DeviceType(strings.TrimSpace(input.DeviceType)),
		Channel:     domain.Channel(strings.TrimSpace(input.Channel)),
		Status:      domain.SessionStatus(strings.TrimSpace(input.Status)),
	}.Normalize()
	if filter.DeviceType != "" && !filter.DeviceType.Valid() {
		return domain.SessionListFilter{}, fmt.Errorf("%w: device_type must be desktop, mobile, tablet, bot, or unknown", ErrInvalidSessionInput)
	}
	if filter.Channel != "" && !filter.Channel.Valid() {
		return domain.SessionListFilter{}, fmt.Errorf("%w: channel must be supported", ErrInvalidSessionInput)
	}
	if filter.Status != "" && !filter.Status.Valid() {
		return domain.SessionListFilter{}, fmt.Errorf("%w: status must be active, ended, expired, or revoked", ErrInvalidSessionInput)
	}
	return filter, nil
}

func (u *AnalyticsUsecase) normalizeFunnelInput(input GetFunnelReportInput) (domain.FunnelReportFilter, error) {
	from := input.From.UTC()
	to := input.To.UTC()
	if err := validateRequiredDateRange(from, to, daysDuration(u.cfg.MaxFunnelRangeDays)); err != nil {
		return domain.FunnelReportFilter{}, err
	}
	steps := make([]domain.EventType, 0, len(input.Steps))
	for _, raw := range input.Steps {
		step := domain.EventType(strings.TrimSpace(raw))
		if step != "" {
			steps = append(steps, step)
		}
	}
	if err := domain.ValidateFunnelStepTypes(steps, u.cfg.MinFunnelSteps, u.cfg.MaxFunnelSteps); err != nil {
		return domain.FunnelReportFilter{}, fmt.Errorf("%w: %w", ErrInvalidSessionInput, err)
	}
	metric := domain.AnalyticsMetric(strings.TrimSpace(input.Metric))
	if metric == "" {
		metric = u.cfg.DefaultFunnelMetric
	}
	filter := domain.FunnelReportFilter{
		Metric:     metric,
		From:       from,
		To:         to,
		Steps:      steps,
		DeviceType: domain.DeviceType(strings.TrimSpace(input.DeviceType)),
		Channel:    domain.Channel(strings.TrimSpace(input.Channel)),
		Country:    input.Country,
		Campaign:   input.Campaign,
	}.Normalize()
	if filter.DeviceType != "" && !filter.DeviceType.Valid() {
		return domain.FunnelReportFilter{}, fmt.Errorf("%w: device_type must be desktop, mobile, tablet, bot, or unknown", ErrInvalidSessionInput)
	}
	if filter.Channel != "" && !filter.Channel.Valid() {
		return domain.FunnelReportFilter{}, fmt.Errorf("%w: channel must be supported", ErrInvalidSessionInput)
	}
	return filter, nil
}

func validateOptionalDateRange(from time.Time, to time.Time, maxRange time.Duration) error {
	if from.IsZero() && to.IsZero() {
		return nil
	}
	return validateRequiredDateRange(from.UTC(), to.UTC(), maxRange)
}

func validateRequiredDateRange(from time.Time, to time.Time, maxRange time.Duration) error {
	if from.IsZero() || to.IsZero() {
		return fmt.Errorf("%w: from and to are required", ErrInvalidSessionInput)
	}
	if !to.After(from) {
		return fmt.Errorf("%w: to must be after from", ErrInvalidSessionInput)
	}
	if maxRange > 0 && to.Sub(from) > maxRange {
		return fmt.Errorf("%w: date range is too large", ErrInvalidSessionInput)
	}
	return nil
}

func daysDuration(days int) time.Duration {
	return time.Duration(days) * 24 * time.Hour
}

func sessionFilterCacheParams(filter domain.SessionListFilter) map[string]string {
	return map[string]string{
		"user_id":      filter.UserID,
		"anonymous_id": filter.AnonymousID,
		"from":         filter.From.Format(time.RFC3339Nano),
		"to":           filter.To.Format(time.RFC3339Nano),
		"page":         strconv.Itoa(filter.Page),
		"page_size":    strconv.Itoa(filter.PageSize),
		"device_type":  string(filter.DeviceType),
		"channel":      string(filter.Channel),
		"status":       string(filter.Status),
	}
}

func funnelFilterCacheParams(filter domain.FunnelReportFilter) map[string]string {
	return map[string]string{
		"metric":      string(filter.Metric),
		"from":        filter.From.Format(time.RFC3339Nano),
		"to":          filter.To.Format(time.RFC3339Nano),
		"steps":       domain.FunnelCacheStepKey(filter.Steps),
		"device_type": string(filter.DeviceType),
		"channel":     string(filter.Channel),
		"country":     filter.Country,
		"campaign":    filter.Campaign,
	}
}

func analyticsCacheKey(prefix string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := []string{prefix}
	for _, key := range keys {
		if params[key] != "" {
			parts = append(parts, key+"="+params[key])
		}
	}
	return strings.Join(parts, ":")
}

func cloneSessionListOutput(in SessionListOutput) SessionListOutput {
	out := in
	out.Sessions = append([]domain.Session(nil), in.Sessions...)
	return out
}

func cloneFunnelReportOutput(in FunnelReportOutput) FunnelReportOutput {
	out := in
	out.Steps = append([]domain.FunnelStep(nil), in.Steps...)
	return out
}

type analyticsCache struct {
	mu      sync.RWMutex
	entries map[string]analyticsCacheEntry
}

type analyticsCacheEntry struct {
	value     any
	expiresAt time.Time
}

func newAnalyticsCache() *analyticsCache {
	return &analyticsCache{entries: make(map[string]analyticsCacheEntry)}
}

func (c *analyticsCache) Get(key string, now time.Time) (any, bool) {
	if c == nil || strings.TrimSpace(key) == "" {
		return nil, false
	}
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || !entry.expiresAt.After(now) {
		if ok {
			c.mu.Lock()
			delete(c.entries, key)
			c.mu.Unlock()
		}
		return nil, false
	}
	return entry.value, true
}

func (c *analyticsCache) Set(key string, value any, ttl time.Duration, now time.Time) {
	if c == nil || strings.TrimSpace(key) == "" || ttl <= 0 {
		return
	}
	c.mu.Lock()
	c.entries[key] = analyticsCacheEntry{value: value, expiresAt: now.Add(ttl)}
	c.mu.Unlock()
}

func (c AnalyticsConfig) Validate() error {
	if c.LiveWindow <= 0 {
		return errors.New("SESSION_ANALYTICS_LIVE_WINDOW must be greater than zero")
	}
	if c.SessionLookback <= 0 {
		return errors.New("SESSION_ANALYTICS_SESSION_LOOKBACK must be greater than zero")
	}
	if c.MaxSessionRangeDays <= 0 {
		return errors.New("SESSION_ANALYTICS_MAX_SESSION_RANGE_DAYS must be greater than zero")
	}
	if c.DefaultPageSize <= 0 {
		return errors.New("SESSION_ANALYTICS_DEFAULT_PAGE_SIZE must be greater than zero")
	}
	if c.MaxPageSize < c.DefaultPageSize {
		return errors.New("SESSION_ANALYTICS_MAX_PAGE_SIZE must be greater than or equal to default page size")
	}
	if c.MaxFunnelRangeDays <= 0 {
		return errors.New("SESSION_ANALYTICS_MAX_FUNNEL_RANGE_DAYS must be greater than zero")
	}
	if c.RawFallbackRange <= 0 {
		return errors.New("SESSION_ANALYTICS_RAW_FALLBACK_RANGE must be greater than zero")
	}
	if c.MinFunnelSteps <= 0 {
		return errors.New("SESSION_ANALYTICS_MIN_FUNNEL_STEPS must be greater than zero")
	}
	if c.MaxFunnelSteps < c.MinFunnelSteps {
		return errors.New("SESSION_ANALYTICS_MAX_FUNNEL_STEPS must be greater than or equal to min funnel steps")
	}
	if strings.TrimSpace(string(c.DefaultFunnelMetric)) == "" {
		return errors.New("SESSION_ANALYTICS_DEFAULT_FUNNEL_METRIC cannot be empty")
	}
	if c.LiveCacheTTL < 0 || c.SessionsCacheTTL < 0 || c.FunnelsCacheTTL < 0 {
		return errors.New("SESSION_ANALYTICS_CACHE_TTL values cannot be negative")
	}
	return nil
}

func (c AnalyticsConfig) withDefaults() AnalyticsConfig {
	if c.LiveWindow == 0 {
		c.LiveWindow = defaultAnalyticsLiveWindow
	}
	if c.SessionLookback == 0 {
		c.SessionLookback = defaultAnalyticsSessionLookback
	}
	if c.MaxSessionRangeDays == 0 {
		c.MaxSessionRangeDays = defaultAnalyticsMaxSessionRangeDays
	}
	if c.DefaultPageSize == 0 {
		c.DefaultPageSize = defaultAnalyticsDefaultPageSize
	}
	if c.MaxPageSize == 0 {
		c.MaxPageSize = defaultAnalyticsMaxPageSize
	}
	if c.MaxFunnelRangeDays == 0 {
		c.MaxFunnelRangeDays = defaultAnalyticsMaxFunnelRangeDays
	}
	if c.RawFallbackRange == 0 {
		c.RawFallbackRange = defaultAnalyticsRawFallbackRange
	}
	if c.MinFunnelSteps == 0 {
		c.MinFunnelSteps = defaultAnalyticsMinFunnelSteps
	}
	if c.MaxFunnelSteps == 0 {
		c.MaxFunnelSteps = defaultAnalyticsMaxFunnelSteps
	}
	if c.DefaultFunnelMetric == "" {
		c.DefaultFunnelMetric = domain.AnalyticsMetricFunnelCheckout
	}
	if c.LiveCacheTTL == 0 {
		c.LiveCacheTTL = defaultAnalyticsLiveCacheTTL
	}
	if c.SessionsCacheTTL == 0 {
		c.SessionsCacheTTL = defaultAnalyticsSessionsCacheTTL
	}
	if c.FunnelsCacheTTL == 0 {
		c.FunnelsCacheTTL = defaultAnalyticsFunnelsCacheTTL
	}
	return c
}
