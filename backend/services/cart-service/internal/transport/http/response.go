package httptransport

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const maxRequestBodyBytes = 1 << 20

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, fmt.Sprintf("encode response: %v", err), http.StatusInternalServerError)
	}
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	writeAPIErrorDetails(w, status, code, message, nil)
}

func writeAPIErrorDetails(w http.ResponseWriter, status int, code string, message string, details map[string]any) {
	writeJSON(w, status, errorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	return false
}
