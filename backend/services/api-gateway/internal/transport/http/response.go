package httptransport

import (
	"encoding/json"
	"net/http"

	gatewayerrors "ecommerce/api-gateway/internal/errors"
	"ecommerce/api-gateway/internal/observability"
)

func writeSuccess(w http.ResponseWriter, r *http.Request, status int, data any) {
	writeJSON(w, status, ResponseEnvelope{
		Data:      data,
		RequestID: RequestIDFromContext(r.Context()),
		Error:     nil,
	})
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code string, message string, details ...string) {
	writeDetailedError(w, r, status, code, message, detailStrings(details)...)
}

func writeDetailedError(w http.ResponseWriter, r *http.Request, status int, code string, message string, details ...ErrorDetailDTO) {
	var bodyDetails any
	if len(details) > 0 {
		bodyDetails = details
	}
	writeMappedError(w, r, gatewayerrors.New(status, code, message, bodyDetails, nil))
}

func writeMappedError(w http.ResponseWriter, r *http.Request, err error) gatewayerrors.MappedError {
	mapped := gatewayerrors.Map(r.Context(), err)
	writeMappedResponse(w, r, mapped)
	return mapped
}

func writeMappedResponse(w http.ResponseWriter, r *http.Request, mapped gatewayerrors.MappedError) {
	if mapped.Status == 0 || mapped.Status == gatewayerrors.StatusClientClosedRequest {
		return
	}
	observability.RecordError(w, observability.ErrorInfo{
		Status:            mapped.Status,
		Code:              mapped.Code,
		GRPCCode:          mapped.GRPCCode,
		DownstreamService: mapped.DownstreamService,
	})
	for key, value := range mapped.Headers {
		if key != "" && value != "" {
			w.Header().Set(key, value)
		}
	}
	writeJSON(w, mapped.Status, ResponseEnvelope{
		Data:      nil,
		RequestID: RequestIDFromContext(r.Context()),
		Error: &ErrorDTO{
			Code:    mapped.Code,
			Message: mapped.Message,
			Details: mapped.Details,
		},
	})
}

func detailStrings(details []string) []ErrorDetailDTO {
	out := make([]ErrorDetailDTO, 0, len(details))
	for _, detail := range details {
		if detail == "" {
			continue
		}
		out = append(out, ErrorDetailDTO{Reason: detail})
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, payload ResponseEnvelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
