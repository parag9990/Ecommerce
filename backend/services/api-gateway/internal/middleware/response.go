package middleware

import (
	"encoding/json"
	"net/http"
)

type apiResponse struct {
	Data      any       `json:"data"`
	RequestID string    `json:"request_id"`
	Error     *apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeAPIError(w http.ResponseWriter, r *http.Request, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiResponse{
		Data:      nil,
		RequestID: RequestIDFromRequest(r),
		Error: &apiError{
			Code:    code,
			Message: message,
		},
	})
}
