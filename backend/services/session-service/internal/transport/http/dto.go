package httptransport

import (
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

type apiEnvelope struct {
	Data      any       `json:"data,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Error     *apiError `json:"error,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type privacySettingsResponse struct {
	Masking     domain.PrivacyMaskingSettings `json:"masking"`
	Permissions domain.PrivacyPermissions     `json:"permissions"`
	UpdatedAt   time.Time                     `json:"updatedAt"`
	UpdatedBy   string                        `json:"updatedBy"`
}

type deletionRequestsResponse struct {
	Items []domain.DeletionRequest `json:"items"`
}

type updatePrivacySettingsRequest struct {
	Masking domain.PrivacyMaskingSettings `json:"masking"`
}

type updateRetentionSettingsRequest struct {
	Retention *domain.RetentionSettings `json:"retention,omitempty"`
	Reason    string                    `json:"reason"`

	RawEventsDays             int `json:"rawEventsDays,omitempty"`
	JourneySummariesDays      int `json:"journeySummariesDays,omitempty"`
	HeatmapAggregatesDays     int `json:"heatmapAggregatesDays,omitempty"`
	AnalyticsAggregatesMonths int `json:"analyticsAggregatesMonths,omitempty"`
	ActiveSessionTTLMinutes   int `json:"activeSessionTtlMinutes,omitempty"`
	DeletionRequestLogDays    int `json:"deletionRequestLogDays,omitempty"`
}

type deletionPreviewRequest struct {
	TargetType  domain.DeletionTargetType `json:"targetType"`
	TargetValue string                    `json:"targetValue"`
}

type createDeletionRequest struct {
	TargetType  domain.DeletionTargetType `json:"targetType"`
	TargetValue string                    `json:"targetValue"`
	Reason      string                    `json:"reason"`
	Confirmed   bool                      `json:"confirmed"`
}

func privacySettingsDTO(settings domain.PrivacySettings) privacySettingsResponse {
	return privacySettingsResponse{
		Masking:     settings.Masking,
		Permissions: settings.Permissions,
		UpdatedAt:   settings.UpdatedAt,
		UpdatedBy:   settings.UpdatedBy,
	}
}

func retentionFromRequest(req updateRetentionSettingsRequest) domain.RetentionSettings {
	if req.Retention != nil {
		return *req.Retention
	}
	return domain.RetentionSettings{
		RawEventsDays:             req.RawEventsDays,
		JourneySummariesDays:      req.JourneySummariesDays,
		HeatmapAggregatesDays:     req.HeatmapAggregatesDays,
		AnalyticsAggregatesMonths: req.AnalyticsAggregatesMonths,
		ActiveSessionTTLMinutes:   req.ActiveSessionTTLMinutes,
		DeletionRequestLogDays:    req.DeletionRequestLogDays,
	}
}
