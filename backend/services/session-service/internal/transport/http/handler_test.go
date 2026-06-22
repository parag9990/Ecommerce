package httptransport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/usecase"
)

func TestHandleIngestEventReturnsAcceptedAndUsesTrustedUserHeader(t *testing.T) {
	ingest := &fakeEventIngestUsecase{
		output: usecase.IngestEventOutput{
			Accepted:  true,
			RequestID: "req_out",
			EventID:   "evt_123",
		},
	}
	handler := newTestHandler(t, ingest, 64<<10)
	body := []byte(`{
		"event_type": "page_view",
		"anonymous_id": "anon_123",
		"session_id": "sess_123",
		"user_id": "user_body",
		"occurred_at": "2026-05-22T10:00:00Z",
		"path": "/",
		"properties": {"title": "Home"}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "req_header")
	req.Header.Set("X-User-ID", "user_trusted")
	rr := httptest.NewRecorder()

	handler.RegisterRoutes(http.NewServeMux())
	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rr.Code, rr.Body.String())
	}
	if ingest.input.UserID == nil || *ingest.input.UserID != "user_trusted" {
		t.Fatalf("expected trusted header user id, got %#v", ingest.input.UserID)
	}
	if ingest.input.RequestID != "req_header" {
		t.Fatalf("expected request id propagation, got %q", ingest.input.RequestID)
	}
	if !ingest.input.OccurredAt.Equal(time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected occurred_at %s", ingest.input.OccurredAt)
	}
	if ingest.input.UserAgent != "" {
		t.Fatalf("unexpected empty test user agent propagation: %q", ingest.input.UserAgent)
	}

	var response acceptedResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.Accepted || response.RequestID != "req_out" || response.EventID != "evt_123" {
		t.Fatalf("unexpected response %+v", response)
	}
}

func TestHandleIngestEventUsesTrustedProxyDeviceContext(t *testing.T) {
	ingest := &fakeEventIngestUsecase{output: usecase.IngestEventOutput{Accepted: true, RequestID: "req_out"}}
	requestCfg, err := NewRequestContextConfig([]string{"192.0.2.0/24"})
	if err != nil {
		t.Fatalf("new request context config: %v", err)
	}
	handler, err := NewHandler(map[string]ProbeFunc{
		"ok": func(ctx context.Context) error { return nil },
	}, ingest, &fakeSessionJourneyUsecase{}, &fakeSessionHeatmapUsecase{}, &fakeAnalyticsUsecase{}, 64<<10, slog.New(slog.NewTextHandler(io.Discard, nil)), requestCfg)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	body := []byte(`{
		"event_type": "page_view",
		"anonymous_id": "anon_123",
		"session_id": "sess_123",
		"occurred_at": "2026-05-22T10:00:00Z",
		"path": "/"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/events", bytes.NewReader(body))
	req.RemoteAddr = "192.0.2.10:12345"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 Chrome/125.0")
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	req.Header.Set("X-Client-Channel", string(domain.ChannelUserAppWeb))
	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	req.Header.Set("Sec-CH-UA-Platform", `"Android"`)
	req.Header.Set("X-Geo-Country", "IN")
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rr.Code, rr.Body.String())
	}
	if ingest.input.IPAddress != "203.0.113.10" {
		t.Fatalf("expected trusted forwarded client IP, got %q", ingest.input.IPAddress)
	}
	if ingest.input.Locale != "en-US" {
		t.Fatalf("expected locale propagation, got %q", ingest.input.Locale)
	}
	if ingest.input.ClientHints.Platform != `"Android"` {
		t.Fatalf("expected client hint propagation, got %+v", ingest.input.ClientHints)
	}
	if ingest.input.GeoHint.Country == nil || *ingest.input.GeoHint.Country != "IN" {
		t.Fatalf("expected trusted geo hint, got %+v", ingest.input.GeoHint)
	}
}

func TestHandleIngestEventRejectsUnsupportedContentType(t *testing.T) {
	ingest := &fakeEventIngestUsecase{}
	handler := newTestHandler(t, ingest, 64<<10)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/events", bytes.NewReader([]byte(`{}`)))
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rr.Code)
	}
	if ingest.called {
		t.Fatal("ingest usecase should not be called for unsupported content type")
	}
}

func TestHandleIngestEventRejectsLargePayload(t *testing.T) {
	ingest := &fakeEventIngestUsecase{}
	handler := newTestHandler(t, ingest, 16)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/events", bytes.NewReader([]byte(`{"event_type":"page_view","anonymous_id":"anon_123"}`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rr.Code)
	}
	if ingest.called {
		t.Fatal("ingest usecase should not be called for oversized payload")
	}
}

func TestHandleGetJourneyReturnsStrictContractForAdmin(t *testing.T) {
	startedAt := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	path := "/"
	journey := &fakeSessionJourneyUsecase{
		output: usecase.JourneyOutput{
			Session: domain.Session{
				SessionID:   "sess_123",
				AnonymousID: "anon_123",
				UserID:      stringPtr("user_123"),
				StartedAt:   startedAt,
				LastSeenAt:  startedAt.Add(5 * time.Minute),
				Device:      domain.Device{Type: domain.DeviceTypeMobile},
			},
			Events: []domain.SessionEvent{
				{
					EventType:   domain.EventPageView,
					AnonymousID: "anon_123",
					SessionID:   "sess_123",
					OccurredAt:  startedAt,
					Path:        &path,
					Properties:  map[string]any{"title": "Home"},
				},
			},
		},
	}
	handler := newTestHandlerWithJourney(t, &fakeEventIngestUsecase{}, journey, 64<<10)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/sessions/sess_123/journey?limit=200&cursor=2026-05-22T10:00:00Z", nil)
	req.Header.Set("X-User-Roles", "admin")
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if journey.input.SessionID != "sess_123" || journey.input.Limit != 200 {
		t.Fatalf("unexpected journey input: %+v", journey.input)
	}
	if journey.input.CursorOccurredAt == nil || !journey.input.CursorOccurredAt.Equal(startedAt) {
		t.Fatalf("expected cursor propagation, got %#v", journey.input.CursorOccurredAt)
	}
	var response JourneyResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Session.SessionID != "sess_123" || len(response.Events) != 1 || response.Events[0].EventType != domain.EventPageView {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestHandleGetJourneyRequiresAdminRole(t *testing.T) {
	handler := newTestHandler(t, &fakeEventIngestUsecase{}, 64<<10)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/sessions/sess_123/journey", nil)
	req.Header.Set("X-User-Roles", "buyer")
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestHandleGetJourneyMapsSessionNotFound(t *testing.T) {
	journey := &fakeSessionJourneyUsecase{err: domain.ErrSessionNotFound}
	handler := newTestHandlerWithJourney(t, &fakeEventIngestUsecase{}, journey, 64<<10)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/sessions/sess_missing/journey", nil)
	req.Header.Set("X-User-Roles", "admin")
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandleGetHeatmapReturnsPointsForAdmin(t *testing.T) {
	heatmap := &fakeSessionHeatmapUsecase{
		output: usecase.HeatmapOutput{
			Path:        "/products/prod_123",
			DeviceType:  domain.DeviceTypeMobile,
			HeatmapType: domain.HeatmapTypeClick,
			From:        "2026-05-22",
			To:          "2026-05-22",
			Points: []usecase.HeatmapPointOutput{
				{X: 55, Y: 70, Weight: 42},
			},
			MaxWeight:   42,
			TotalEvents: 42,
		},
	}
	handler := newTestHandlerWithHeatmap(t, &fakeEventIngestUsecase{}, &fakeSessionJourneyUsecase{}, heatmap, 64<<10)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/heatmaps?path=/products/prod_123&device_type=mobile&from=2026-05-22&to=2026-05-22", nil)
	req.Header.Set("X-User-Roles", "admin")
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if heatmap.input.Path != "/products/prod_123" || heatmap.input.DeviceType != "mobile" {
		t.Fatalf("unexpected heatmap input: %+v", heatmap.input)
	}
	var response HeatmapResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Points) != 1 || response.Points[0].Weight != 42 || response.MaxWeight != 42 {
		t.Fatalf("unexpected heatmap response: %+v", response)
	}
}

func TestHandleGetHeatmapAcceptsRFC3339Range(t *testing.T) {
	heatmap := &fakeSessionHeatmapUsecase{
		output: usecase.HeatmapOutput{
			Path:        "/products/prod_123",
			DeviceType:  domain.DeviceTypeMobile,
			HeatmapType: domain.HeatmapTypeClick,
			From:        "2026-05-22",
			To:          "2026-05-23",
			Points:      []usecase.HeatmapPointOutput{},
		},
	}
	handler := newTestHandlerWithHeatmap(t, &fakeEventIngestUsecase{}, &fakeSessionJourneyUsecase{}, heatmap, 64<<10)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/heatmaps?path=/products/prod_123&device_type=mobile&from=2026-05-22T10:30:00Z&to=2026-05-23T11:45:00Z", nil)
	req.Header.Set("X-User-Roles", "admin")
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !heatmap.input.From.Equal(time.Date(2026, 5, 22, 10, 30, 0, 0, time.UTC)) {
		t.Fatalf("unexpected from timestamp: %s", heatmap.input.From)
	}
	if !heatmap.input.To.Equal(time.Date(2026, 5, 23, 11, 45, 0, 0, time.UTC)) {
		t.Fatalf("unexpected to timestamp: %s", heatmap.input.To)
	}
}

func TestHandleGetHeatmapRequiresAdminRole(t *testing.T) {
	handler := newTestHandler(t, &fakeEventIngestUsecase{}, 64<<10)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/heatmaps?path=/&device_type=desktop&from=2026-05-22&to=2026-05-22", nil)
	req.Header.Set("X-User-Roles", "buyer")
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestHandleGetHeatmapValidatesDate(t *testing.T) {
	handler := newTestHandler(t, &fakeEventIngestUsecase{}, 64<<10)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/heatmaps?path=/&device_type=desktop&from=bad&to=2026-05-22", nil)
	req.Header.Set("X-User-Roles", "admin")
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleGetLiveMetricsReturnsMetricsForAdmin(t *testing.T) {
	analytics := &fakeAnalyticsUsecase{
		liveOutput: domain.LiveMetrics{ActiveUsers: 42, ActiveSessions: 51, EventsPerMinute: 18.5},
	}
	handler := newTestHandlerWithAnalytics(t, &fakeEventIngestUsecase{}, &fakeSessionJourneyUsecase{}, &fakeSessionHeatmapUsecase{}, analytics, 64<<10)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/live", nil)
	req.Header.Set("X-User-Roles", "admin")
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var response LiveMetricsResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ActiveUsers != 42 || response.ActiveSessions != 51 || response.EventsPerMinute != 18.5 {
		t.Fatalf("unexpected live metrics response: %+v", response)
	}
}

func TestHandleListSessionsParsesFiltersForAdmin(t *testing.T) {
	startedAt := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	analytics := &fakeAnalyticsUsecase{
		sessionsOutput: usecase.SessionListOutput{
			Sessions: []domain.Session{{
				SessionID:     "sess_1234567890",
				AnonymousID:   "anon_1234567890",
				UserID:        stringPtr("user_1234567890"),
				Status:        domain.SessionStatusActive,
				Channel:       domain.ChannelUserAppWeb,
				EntryPage:     "/",
				StartedAt:     startedAt,
				LastSeenAt:    startedAt.Add(time.Minute),
				Device:        domain.Device{Type: domain.DeviceTypeMobile},
				Geo:           domain.Geo{Source: domain.GeoSourceUnknown},
				SchemaVersion: domain.CurrentSessionSchemaVersion,
			}},
			Page:     1,
			PageSize: 50,
			Total:    1,
		},
	}
	handler := newTestHandlerWithAnalytics(t, &fakeEventIngestUsecase{}, &fakeSessionJourneyUsecase{}, &fakeSessionHeatmapUsecase{}, analytics, 64<<10)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/sessions?user_id=user_123&from=2026-05-22T00:00:00Z&to=2026-05-23T00:00:00Z&page=1&page_size=50&device_type=mobile", nil)
	req.Header.Set("X-User-Roles", "admin")
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if analytics.sessionsInput.UserID != "user_123" || analytics.sessionsInput.DeviceType != "mobile" {
		t.Fatalf("unexpected sessions input: %+v", analytics.sessionsInput)
	}
	var response SessionListResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Total != 1 || len(response.Sessions) != 1 || response.Sessions[0].AnonymousID == "anon_1234567890" {
		t.Fatalf("unexpected session list response: %+v", response)
	}
}

func TestHandleGetFunnelReportReturnsStepsForAdmin(t *testing.T) {
	from := time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	analytics := &fakeAnalyticsUsecase{
		funnelOutput: usecase.FunnelReportOutput{
			From:   from,
			To:     to,
			Source: "raw_fallback",
			Steps: []domain.FunnelStep{
				{Name: "product_view", EventType: domain.EventProductView, Count: 100, UniqueSessions: 80, ConversionFromPrevious: 100},
				{Name: "add_to_cart", EventType: domain.EventAddToCart, Count: 20, UniqueSessions: 16, ConversionFromPrevious: 20, DropoffFromPrevious: 80},
			},
			OverallConversion: 20,
		},
	}
	handler := newTestHandlerWithAnalytics(t, &fakeEventIngestUsecase{}, &fakeSessionJourneyUsecase{}, &fakeSessionHeatmapUsecase{}, analytics, 64<<10)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/funnels?from=2026-05-22T00:00:00Z&to=2026-05-23T00:00:00Z&steps=product_view,add_to_cart", nil)
	req.Header.Set("X-User-Roles", "admin")
	rr := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if len(analytics.funnelInput.Steps) != 2 || analytics.funnelInput.Steps[0] != "product_view" {
		t.Fatalf("unexpected funnel input: %+v", analytics.funnelInput)
	}
	var response FunnelReportResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Steps) != 2 || response.OverallConversion != 20 || response.Source != "raw_fallback" {
		t.Fatalf("unexpected funnel response: %+v", response)
	}
}

func newTestHandler(t *testing.T, ingest EventIngestUsecase, maxBodyBytes int64) *Handler {
	return newTestHandlerWithJourney(t, ingest, &fakeSessionJourneyUsecase{}, maxBodyBytes)
}

func newTestHandlerWithJourney(t *testing.T, ingest EventIngestUsecase, journey SessionJourneyUsecase, maxBodyBytes int64) *Handler {
	return newTestHandlerWithHeatmap(t, ingest, journey, &fakeSessionHeatmapUsecase{}, maxBodyBytes)
}

func newTestHandlerWithHeatmap(t *testing.T, ingest EventIngestUsecase, journey SessionJourneyUsecase, heatmap SessionHeatmapUsecase, maxBodyBytes int64) *Handler {
	return newTestHandlerWithAnalytics(t, ingest, journey, heatmap, &fakeAnalyticsUsecase{}, maxBodyBytes)
}

func newTestHandlerWithAnalytics(t *testing.T, ingest EventIngestUsecase, journey SessionJourneyUsecase, heatmap SessionHeatmapUsecase, analytics AnalyticsUsecase, maxBodyBytes int64) *Handler {
	t.Helper()
	handler, err := NewHandler(map[string]ProbeFunc{
		"ok": func(ctx context.Context) error { return nil },
	}, ingest, journey, heatmap, analytics, maxBodyBytes, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	return handler
}

type fakeEventIngestUsecase struct {
	input  usecase.IngestEventInput
	output usecase.IngestEventOutput
	err    error
	called bool
}

func (f *fakeEventIngestUsecase) IngestEvent(ctx context.Context, input usecase.IngestEventInput) (usecase.IngestEventOutput, error) {
	f.called = true
	f.input = input
	return f.output, f.err
}

type fakeSessionJourneyUsecase struct {
	input  usecase.GetJourneyInput
	output usecase.JourneyOutput
	err    error
	called bool
}

func (f *fakeSessionJourneyUsecase) GetJourney(ctx context.Context, input usecase.GetJourneyInput) (usecase.JourneyOutput, error) {
	f.called = true
	f.input = input
	return f.output, f.err
}

type fakeSessionHeatmapUsecase struct {
	input  usecase.GetHeatmapInput
	output usecase.HeatmapOutput
	err    error
	called bool
}

func (f *fakeSessionHeatmapUsecase) GetHeatmap(ctx context.Context, input usecase.GetHeatmapInput) (usecase.HeatmapOutput, error) {
	f.called = true
	f.input = input
	return f.output, f.err
}

type fakeAnalyticsUsecase struct {
	liveInput       usecase.GetLiveMetricsInput
	liveOutput      domain.LiveMetrics
	liveErr         error
	sessionsInput   usecase.ListSessionsInput
	sessionsOutput  usecase.SessionListOutput
	sessionsErr     error
	funnelInput     usecase.GetFunnelReportInput
	funnelOutput    usecase.FunnelReportOutput
	funnelErr       error
	retentionInput  usecase.GetRetentionReportInput
	retentionOutput domain.RetentionReport
	retentionErr    error
}

func (f *fakeAnalyticsUsecase) GetLiveMetrics(ctx context.Context, input usecase.GetLiveMetricsInput) (domain.LiveMetrics, error) {
	f.liveInput = input
	return f.liveOutput, f.liveErr
}

func (f *fakeAnalyticsUsecase) ListSessions(ctx context.Context, input usecase.ListSessionsInput) (usecase.SessionListOutput, error) {
	f.sessionsInput = input
	return f.sessionsOutput, f.sessionsErr
}

func (f *fakeAnalyticsUsecase) GetFunnelReport(ctx context.Context, input usecase.GetFunnelReportInput) (usecase.FunnelReportOutput, error) {
	f.funnelInput = input
	return f.funnelOutput, f.funnelErr
}

func (f *fakeAnalyticsUsecase) GetRetentionReport(ctx context.Context, input usecase.GetRetentionReportInput) (domain.RetentionReport, error) {
	f.retentionInput = input
	return f.retentionOutput, f.retentionErr
}

func stringPtr(value string) *string {
	return &value
}
