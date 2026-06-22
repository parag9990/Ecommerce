package httptransport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

func TestAnalyticsAdminTokenIsRequiredWhenConfigured(t *testing.T) {
	handler := newTestHandler(t, &fakeEventIngestUsecase{}, 64<<10)
	handler.SetAdminToken("test_session_admin_token_at_least_32_chars")
	request := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/live", nil)
	request.Header.Set("X-User-Roles", "admin")
	response := httptest.NewRecorder()
	NewRouter(handler).ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAnalyticsMaskingDirectiveCannotExposePII(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/sessions", nil)
	request.Header.Set("X-Admin-Mask-PII", "true")
	masking := domain.PrivacyMaskingSettings{UserIDMode: domain.MaskingModeFull, AnonymousIDMode: domain.MaskingModeFull, SessionIDMode: domain.MaskingModeFull, LocationGranularity: domain.LocationGranularityCity, ShowSearchQueries: true, ShowIPHash: true}
	result := analyticsMaskingForRequest(request, masking)
	if result.UserIDMode != domain.MaskingModeMasked || result.AnonymousIDMode != domain.MaskingModeMasked || result.SessionIDMode != domain.MaskingModeMasked || result.LocationGranularity != domain.LocationGranularityCountry || result.ShowSearchQueries || result.ShowIPHash {
		t.Fatalf("masking=%+v", result)
	}
}
