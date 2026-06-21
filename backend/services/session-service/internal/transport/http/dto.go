package httptransport

import (
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/usecase"
)

type ingestEventRequest struct {
	EventType   string         `json:"event_type"`
	AnonymousID string         `json:"anonymous_id"`
	SessionID   string         `json:"session_id"`
	UserID      *string        `json:"user_id,omitempty"`
	OccurredAt  string         `json:"occurred_at"`
	Path        *string        `json:"path,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
}

type acceptedResponse struct {
	Accepted  bool   `json:"accepted"`
	RequestID string `json:"request_id"`
	EventID   string `json:"event_id,omitempty"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SessionResponse struct {
	SessionID             string                   `json:"session_id"`
	AnonymousID           string                   `json:"anonymous_id"`
	UserID                *string                  `json:"user_id,omitempty"`
	SchemaVersion         int                      `json:"schema_version"`
	Status                domain.SessionStatus     `json:"status"`
	Channel               domain.Channel           `json:"channel"`
	EntryPage             string                   `json:"entry_page"`
	ExitPage              *string                  `json:"exit_page,omitempty"`
	Referrer              *string                  `json:"referrer,omitempty"`
	UTM                   UTMResponse              `json:"utm"`
	UserAgent             string                   `json:"user_agent"`
	UserAgentHash         *string                  `json:"user_agent_hash,omitempty"`
	Device                DeviceResponse           `json:"device"`
	Client                ClientResponse           `json:"client"`
	IPHash                string                   `json:"ip_hash"`
	IPVersion             domain.IPVersion         `json:"ip_version,omitempty"`
	Geo                   GeoResponse              `json:"geo"`
	DeviceFingerprintHash *string                  `json:"device_fingerprint_hash,omitempty"`
	AuthSessionID         *string                  `json:"auth_session_id,omitempty"`
	RiskLevel             domain.RiskLevel         `json:"risk_level"`
	RiskReasons           []string                 `json:"risk_reasons"`
	StartedAt             time.Time                `json:"started_at"`
	LastSeenAt            time.Time                `json:"last_seen_at"`
	EndedAt               *time.Time               `json:"ended_at,omitempty"`
	RevokedAt             *time.Time               `json:"revoked_at,omitempty"`
	EndReason             *domain.SessionEndReason `json:"end_reason,omitempty"`
	CreatedAt             time.Time                `json:"created_at"`
	UpdatedAt             time.Time                `json:"updated_at"`
}

type DeviceResponse struct {
	Type           domain.DeviceType `json:"type"`
	Browser        *string           `json:"browser,omitempty"`
	BrowserVersion *string           `json:"browser_version,omitempty"`
	OS             *string           `json:"os,omitempty"`
	OSVersion      *string           `json:"os_version,omitempty"`
	Model          *string           `json:"model,omitempty"`
	Vendor         *string           `json:"vendor,omitempty"`
	IsBot          bool              `json:"is_bot"`
}

type ClientResponse struct {
	Channel        domain.Channel `json:"channel"`
	Locale         *string        `json:"locale,omitempty"`
	Timezone       *string        `json:"timezone,omitempty"`
	ScreenWidth    int            `json:"screen_width,omitempty"`
	ScreenHeight   int            `json:"screen_height,omitempty"`
	ViewportWidth  int            `json:"viewport_width,omitempty"`
	ViewportHeight int            `json:"viewport_height,omitempty"`
}

type GeoResponse struct {
	Country  *string          `json:"country,omitempty"`
	Region   *string          `json:"region,omitempty"`
	City     *string          `json:"city,omitempty"`
	Timezone *string          `json:"timezone,omitempty"`
	Source   domain.GeoSource `json:"source"`
}

type UTMResponse struct {
	Source   *string `json:"source,omitempty"`
	Medium   *string `json:"medium,omitempty"`
	Campaign *string `json:"campaign,omitempty"`
	Term     *string `json:"term,omitempty"`
	Content  *string `json:"content,omitempty"`
}

type JourneyResponse struct {
	Session JourneySessionResponse  `json:"session"`
	Events  []JourneyEventResponse  `json:"events"`
	Summary *JourneySummaryResponse `json:"summary,omitempty"`
}

type JourneySessionResponse struct {
	SessionID   string         `json:"session_id"`
	AnonymousID string         `json:"anonymous_id"`
	UserID      *string        `json:"user_id,omitempty"`
	StartedAt   time.Time      `json:"started_at"`
	LastSeenAt  time.Time      `json:"last_seen_at"`
	Device      DeviceResponse `json:"device"`
}

type JourneyEventResponse struct {
	EventType   domain.EventType `json:"event_type"`
	AnonymousID string           `json:"anonymous_id"`
	SessionID   string           `json:"session_id"`
	UserID      *string          `json:"user_id,omitempty"`
	OccurredAt  time.Time        `json:"occurred_at"`
	Path        *string          `json:"path,omitempty"`
	Properties  map[string]any   `json:"properties,omitempty"`
	Device      *DeviceResponse  `json:"device,omitempty"`
	Client      *ClientResponse  `json:"client,omitempty"`
	Geo         *GeoResponse     `json:"geo,omitempty"`
}

type JourneySummaryResponse struct {
	EntryPage        string                     `json:"entry_page,omitempty"`
	ExitPage         string                     `json:"exit_page,omitempty"`
	FirstEventAt     *time.Time                 `json:"first_event_at,omitempty"`
	LastEventAt      *time.Time                 `json:"last_event_at,omitempty"`
	DurationSeconds  int64                      `json:"duration_seconds"`
	TotalEvents      int                        `json:"total_events"`
	ProductsViewed   int                        `json:"products_viewed"`
	Searches         int                        `json:"searches"`
	CartActions      int                        `json:"cart_actions"`
	CheckoutStarted  bool                       `json:"checkout_started"`
	PaymentCompleted bool                       `json:"payment_completed"`
	Milestones       []JourneyMilestoneResponse `json:"milestones,omitempty"`
	TopPaths         []JourneyTopPathResponse   `json:"top_paths,omitempty"`
	CalculatedAt     *time.Time                 `json:"calculated_at,omitempty"`
}

type JourneyMilestoneResponse struct {
	Name       domain.JourneyMilestoneName `json:"name"`
	EventID    string                      `json:"event_id"`
	Path       string                      `json:"path,omitempty"`
	OccurredAt time.Time                   `json:"occurred_at"`
}

type JourneyTopPathResponse struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

type HeatmapResponse struct {
	Path        string                 `json:"path"`
	DeviceType  domain.DeviceType      `json:"device_type"`
	HeatmapType domain.HeatmapType     `json:"heatmap_type"`
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Points      []HeatmapPointResponse `json:"points"`
	MaxWeight   int                    `json:"max_weight"`
	TotalEvents int                    `json:"total_events"`
}

type HeatmapPointResponse struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Weight int `json:"weight"`
}

type LiveMetricsResponse struct {
	ActiveUsers     int64     `json:"active_users"`
	ActiveSessions  int64     `json:"active_sessions"`
	EventsPerMinute float64   `json:"events_per_minute"`
	WindowSeconds   int64     `json:"window_seconds,omitempty"`
	MeasuredAt      time.Time `json:"measured_at,omitempty"`
}

type SessionListResponse struct {
	Sessions []AnalyticsSessionResponse `json:"sessions"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"page_size"`
	Total    int64                      `json:"total"`
	HasNext  bool                       `json:"has_next"`
}

type AnalyticsSessionResponse struct {
	SessionID   string               `json:"session_id"`
	AnonymousID string               `json:"anonymous_id"`
	UserID      *string              `json:"user_id,omitempty"`
	Status      domain.SessionStatus `json:"status"`
	Channel     domain.Channel       `json:"channel"`
	EntryPage   string               `json:"entry_page"`
	ExitPage    *string              `json:"exit_page,omitempty"`
	StartedAt   time.Time            `json:"started_at"`
	LastSeenAt  time.Time            `json:"last_seen_at"`
	EndedAt     *time.Time           `json:"ended_at,omitempty"`
	Device      DeviceResponse       `json:"device"`
	Geo         GeoResponse          `json:"geo"`
}

type FunnelReportResponse struct {
	Steps             []FunnelStepResponse `json:"steps"`
	OverallConversion float64              `json:"overall_conversion"`
	Source            string               `json:"source,omitempty"`
	From              time.Time            `json:"from"`
	To                time.Time            `json:"to"`
}

type FunnelStepResponse struct {
	Name                   string           `json:"name"`
	EventType              domain.EventType `json:"event_type,omitempty"`
	Count                  int64            `json:"count"`
	UniqueSessions         int64            `json:"unique_sessions"`
	UniqueUsers            int64            `json:"unique_users,omitempty"`
	ConversionFromPrevious float64          `json:"conversion_from_previous"`
	DropoffFromPrevious    float64          `json:"dropoff_from_previous"`
}

type deleteUserSessionDataRequest struct {
	RequestID    string   `json:"request_id,omitempty"`
	UserID       string   `json:"user_id,omitempty"`
	AnonymousIDs []string `json:"anonymous_ids,omitempty"`
	Reason       string   `json:"reason"`
	HardDelete   bool     `json:"hard_delete"`
	RequestedBy  string   `json:"requested_by,omitempty"`
}

type runRetentionCleanupRequest struct {
	RequestID string `json:"request_id,omitempty"`
	DryRun    *bool  `json:"dry_run,omitempty"`
}

type setLegalHoldRequest struct {
	Hold    bool   `json:"hold"`
	Reason  string `json:"reason,omitempty"`
	ActorID string `json:"actor_id,omitempty"`
}

type DeleteUserSessionDataResponse struct {
	RequestID          string    `json:"request_id"`
	RawEventsDeleted   int64     `json:"raw_events_deleted"`
	SessionsAnonymized int64     `json:"sessions_anonymized"`
	SessionsDeleted    int64     `json:"sessions_deleted,omitempty"`
	JourneysAnonymized int64     `json:"journeys_anonymized"`
	JourneysDeleted    int64     `json:"journeys_deleted,omitempty"`
	RedisKeysDeleted   int64     `json:"redis_keys_deleted"`
	RedisErrors        int64     `json:"redis_errors,omitempty"`
	HardDelete         bool      `json:"hard_delete"`
	CompletedAt        time.Time `json:"completed_at"`
}

type RetentionCleanupResponse struct {
	RequestID                   string    `json:"request_id"`
	DryRun                      bool      `json:"dry_run"`
	SessionsAnonymized          int64     `json:"sessions_anonymized"`
	JourneysAnonymized          int64     `json:"journeys_anonymized"`
	HeatmapPointsPurged         int64     `json:"heatmap_points_purged"`
	HeatmapSessionMarkersPurged int64     `json:"heatmap_session_markers_purged"`
	AnalyticsAggregatesPurged   int64     `json:"analytics_aggregates_purged"`
	AnalyticsAggregatesWithPII  int64     `json:"analytics_aggregates_with_pii"`
	LegalHoldRecordsSkipped     int64     `json:"legal_hold_records_skipped"`
	StartedAt                   time.Time `json:"started_at"`
	CompletedAt                 time.Time `json:"completed_at"`
}

type SetLegalHoldResponse struct {
	SessionID       string    `json:"session_id"`
	LegalHold       bool      `json:"legal_hold"`
	SessionsMatched int64     `json:"sessions_matched"`
	JourneysMatched int64     `json:"journeys_matched"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func SessionResponseFromDomain(session domain.Session) SessionResponse {
	normalized := session.Normalize()
	return SessionResponse{
		SessionID:             normalized.SessionID,
		AnonymousID:           normalized.AnonymousID,
		UserID:                normalized.UserID,
		SchemaVersion:         normalized.SchemaVersion,
		Status:                normalized.Status,
		Channel:               normalized.Channel,
		EntryPage:             normalized.EntryPage,
		ExitPage:              normalized.ExitPage,
		Referrer:              normalized.Referrer,
		UTM:                   utmResponseFromDomain(normalized.UTM),
		UserAgent:             normalized.UserAgent,
		UserAgentHash:         normalized.UserAgentHash,
		Device:                deviceResponseFromDomain(normalized.Device),
		Client:                clientResponseFromDomain(normalized.Client),
		IPHash:                normalized.IPHash,
		IPVersion:             normalized.IPVersion,
		Geo:                   geoResponseFromDomain(normalized.Geo),
		DeviceFingerprintHash: normalized.DeviceFingerprintHash,
		AuthSessionID:         normalized.AuthSessionID,
		RiskLevel:             normalized.RiskLevel,
		RiskReasons:           append([]string(nil), normalized.RiskReasons...),
		StartedAt:             normalized.StartedAt,
		LastSeenAt:            normalized.LastSeenAt,
		EndedAt:               normalized.EndedAt,
		RevokedAt:             normalized.RevokedAt,
		EndReason:             normalized.EndReason,
		CreatedAt:             normalized.CreatedAt,
		UpdatedAt:             normalized.UpdatedAt,
	}
}

func DeleteUserSessionDataResponseFromDomain(result domain.DeleteUserSessionDataResult) DeleteUserSessionDataResponse {
	return DeleteUserSessionDataResponse{
		RequestID:          result.RequestID,
		RawEventsDeleted:   result.RawEventsDeleted,
		SessionsAnonymized: result.SessionsAnonymized,
		SessionsDeleted:    result.SessionsDeleted,
		JourneysAnonymized: result.JourneysAnonymized,
		JourneysDeleted:    result.JourneysDeleted,
		RedisKeysDeleted:   result.RedisKeysDeleted,
		RedisErrors:        result.RedisErrors,
		HardDelete:         result.HardDelete,
		CompletedAt:        result.CompletedAt,
	}
}

func RetentionCleanupResponseFromDomain(result domain.RetentionCleanupResult) RetentionCleanupResponse {
	return RetentionCleanupResponse{
		RequestID:                   result.RequestID,
		DryRun:                      result.DryRun,
		SessionsAnonymized:          result.SessionsAnonymized,
		JourneysAnonymized:          result.JourneysAnonymized,
		HeatmapPointsPurged:         result.HeatmapPointsPurged,
		HeatmapSessionMarkersPurged: result.HeatmapSessionMarkersPurged,
		AnalyticsAggregatesPurged:   result.AnalyticsAggregatesPurged,
		AnalyticsAggregatesWithPII:  result.AnalyticsAggregatesWithPII,
		LegalHoldRecordsSkipped:     result.LegalHoldRecordsSkipped,
		StartedAt:                   result.StartedAt,
		CompletedAt:                 result.CompletedAt,
	}
}

func SetLegalHoldResponseFromDomain(result domain.SetLegalHoldResult) SetLegalHoldResponse {
	return SetLegalHoldResponse{
		SessionID:       result.SessionID,
		LegalHold:       result.LegalHold,
		SessionsMatched: result.SessionsMatched,
		JourneysMatched: result.JourneysMatched,
		UpdatedAt:       result.UpdatedAt,
	}
}

func JourneyResponseFromDomain(session domain.Session, events []domain.SessionEvent, summary domain.JourneySummary) JourneyResponse {
	normalized := session.Normalize()
	response := JourneyResponse{
		Session: JourneySessionResponse{
			SessionID:   normalized.SessionID,
			AnonymousID: normalized.AnonymousID,
			UserID:      normalized.UserID,
			StartedAt:   normalized.StartedAt,
			LastSeenAt:  normalized.LastSeenAt,
			Device:      deviceResponseFromDomain(normalized.Device),
		},
		Events: make([]JourneyEventResponse, 0, len(events)),
	}
	for _, event := range events {
		response.Events = append(response.Events, journeyEventResponseFromDomain(event))
	}
	if normalizedSummary := summary.Normalize(); normalizedSummary.SessionID != "" {
		response.Summary = journeySummaryResponseFromDomain(normalizedSummary)
	}
	return response
}

func LiveMetricsResponseFromDomain(metrics domain.LiveMetrics) LiveMetricsResponse {
	return LiveMetricsResponse{
		ActiveUsers:     metrics.ActiveUsers,
		ActiveSessions:  metrics.ActiveSessions,
		EventsPerMinute: metrics.EventsPerMinute,
		WindowSeconds:   metrics.WindowSeconds,
		MeasuredAt:      metrics.MeasuredAt,
	}
}

func SessionListResponseFromOutput(output usecase.SessionListOutput) SessionListResponse {
	response := SessionListResponse{
		Sessions: make([]AnalyticsSessionResponse, 0, len(output.Sessions)),
		Page:     output.Page,
		PageSize: output.PageSize,
		Total:    output.Total,
		HasNext:  output.HasNext,
	}
	for _, session := range output.Sessions {
		response.Sessions = append(response.Sessions, analyticsSessionResponseFromDomain(session))
	}
	return response
}

func FunnelReportResponseFromOutput(output usecase.FunnelReportOutput) FunnelReportResponse {
	response := FunnelReportResponse{
		Steps:             make([]FunnelStepResponse, 0, len(output.Steps)),
		OverallConversion: output.OverallConversion,
		Source:            output.Source,
		From:              output.From,
		To:                output.To,
	}
	for _, step := range output.Steps {
		normalized := step.Normalize()
		response.Steps = append(response.Steps, FunnelStepResponse{
			Name:                   normalized.Name,
			EventType:              normalized.EventType,
			Count:                  normalized.Count,
			UniqueSessions:         normalized.UniqueSessions,
			UniqueUsers:            normalized.UniqueUsers,
			ConversionFromPrevious: normalized.ConversionFromPrevious,
			DropoffFromPrevious:    normalized.DropoffFromPrevious,
		})
	}
	return response
}

func HeatmapResponseFromOutput(output usecase.HeatmapOutput) HeatmapResponse {
	response := HeatmapResponse{
		Path:        output.Path,
		DeviceType:  output.DeviceType,
		HeatmapType: output.HeatmapType,
		From:        output.From,
		To:          output.To,
		Points:      make([]HeatmapPointResponse, 0, len(output.Points)),
		MaxWeight:   output.MaxWeight,
		TotalEvents: output.TotalEvents,
	}
	for _, point := range output.Points {
		response.Points = append(response.Points, HeatmapPointResponse{
			X:      point.X,
			Y:      point.Y,
			Weight: point.Weight,
		})
	}
	return response
}

func journeySummaryResponseFromDomain(summary domain.JourneySummary) *JourneySummaryResponse {
	response := &JourneySummaryResponse{
		EntryPage:        summary.EntryPage,
		ExitPage:         summary.ExitPage,
		DurationSeconds:  summary.DurationSeconds,
		TotalEvents:      summary.TotalEvents,
		ProductsViewed:   summary.ProductsViewed,
		Searches:         summary.Searches,
		CartActions:      summary.CartActions,
		CheckoutStarted:  summary.CheckoutStarted,
		PaymentCompleted: summary.PaymentCompleted,
		Milestones:       make([]JourneyMilestoneResponse, 0, len(summary.Milestones)),
		TopPaths:         make([]JourneyTopPathResponse, 0, len(summary.TopPaths)),
	}
	if !summary.FirstEventAt.IsZero() {
		response.FirstEventAt = &summary.FirstEventAt
	}
	if !summary.LastEventAt.IsZero() {
		response.LastEventAt = &summary.LastEventAt
	}
	if !summary.CalculatedAt.IsZero() {
		response.CalculatedAt = &summary.CalculatedAt
	}
	for _, milestone := range summary.Milestones {
		response.Milestones = append(response.Milestones, JourneyMilestoneResponse{
			Name:       milestone.Name,
			EventID:    milestone.EventID,
			Path:       milestone.Path,
			OccurredAt: milestone.OccurredAt,
		})
	}
	for _, path := range summary.TopPaths {
		response.TopPaths = append(response.TopPaths, JourneyTopPathResponse{
			Path:  path.Path,
			Count: path.Count,
		})
	}
	return response
}

func analyticsSessionResponseFromDomain(session domain.Session) AnalyticsSessionResponse {
	normalized := session.Normalize()
	return AnalyticsSessionResponse{
		SessionID:   normalized.SessionID,
		AnonymousID: maskAnalyticsID(normalized.AnonymousID),
		UserID:      maskAnalyticsIDPtr(normalized.UserID),
		Status:      normalized.Status,
		Channel:     normalized.Channel,
		EntryPage:   normalized.EntryPage,
		ExitPage:    normalized.ExitPage,
		StartedAt:   normalized.StartedAt,
		LastSeenAt:  normalized.LastSeenAt,
		EndedAt:     normalized.EndedAt,
		Device:      deviceResponseFromDomain(normalized.Device),
		Geo:         geoResponseFromDomain(normalized.Geo),
	}
}

func journeyEventResponseFromDomain(event domain.SessionEvent) JourneyEventResponse {
	normalized := event.Normalize()
	response := JourneyEventResponse{
		EventType:   normalized.EventType,
		AnonymousID: normalized.AnonymousID,
		SessionID:   normalized.SessionID,
		UserID:      normalized.UserID,
		OccurredAt:  normalized.OccurredAt,
		Path:        normalized.Path,
		Properties:  normalized.Properties,
	}
	if normalized.Device != nil {
		device := deviceResponseFromDomain(*normalized.Device)
		response.Device = &device
	}
	if normalized.Client != nil {
		client := clientResponseFromDomain(*normalized.Client)
		response.Client = &client
	}
	if normalized.Geo != nil {
		geo := geoResponseFromDomain(*normalized.Geo)
		response.Geo = &geo
	}
	return response
}

func deviceResponseFromDomain(device domain.Device) DeviceResponse {
	return DeviceResponse{
		Type:           device.Type,
		Browser:        device.Browser,
		BrowserVersion: device.BrowserVersion,
		OS:             device.OS,
		OSVersion:      device.OSVersion,
		Model:          device.Model,
		Vendor:         device.Vendor,
		IsBot:          device.IsBot,
	}
}

func clientResponseFromDomain(client domain.Client) ClientResponse {
	normalized := client.Normalize()
	return ClientResponse{
		Channel:        normalized.Channel,
		Locale:         normalized.Locale,
		Timezone:       normalized.Timezone,
		ScreenWidth:    normalized.ScreenWidth,
		ScreenHeight:   normalized.ScreenHeight,
		ViewportWidth:  normalized.ViewportWidth,
		ViewportHeight: normalized.ViewportHeight,
	}
}

func geoResponseFromDomain(geo domain.Geo) GeoResponse {
	normalized := geo.Normalize()
	return GeoResponse{
		Country:  normalized.Country,
		Region:   normalized.Region,
		City:     normalized.City,
		Timezone: normalized.Timezone,
		Source:   normalized.Source,
	}
}

func utmResponseFromDomain(utm domain.UTM) UTMResponse {
	return UTMResponse{
		Source:   utm.Source,
		Medium:   utm.Medium,
		Campaign: utm.Campaign,
		Term:     utm.Term,
		Content:  utm.Content,
	}
}

func maskAnalyticsIDPtr(value *string) *string {
	if value == nil {
		return nil
	}
	masked := maskAnalyticsID(*value)
	if masked == "" {
		return nil
	}
	return &masked
}

func maskAnalyticsID(value string) string {
	if len(value) <= 10 {
		return value
	}
	return value[:10] + "****"
}
