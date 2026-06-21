package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"
)

type Check func(context.Context) error

func Handler(checks map[string]Check, timeout time.Duration) http.Handler {
	return ReadinessHandler("", "", checks, timeout)
}

func LivenessHandler(service, environment string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, response{Status: "live", Service: service, Environment: environment})
	})
}

func ReadinessHandler(service, environment string, checks map[string]Check, timeout time.Duration) http.Handler {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		results := make(map[string]string, len(checks))
		names := make([]string, 0, len(checks))
		for name := range checks {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			check := checks[name]
			if check != nil {
				if err := check(ctx); err != nil {
					results[name] = "unavailable"
					continue
				}
			}
			results[name] = "ok"
		}
		status, state := http.StatusOK, "ready"
		for _, result := range results {
			if result != "ok" {
				status, state = http.StatusServiceUnavailable, "not_ready"
				break
			}
		}
		writeJSON(w, status, response{Status: state, Service: service, Environment: environment, Checks: results})
	})
}

type response struct {
	Status      string            `json:"status"`
	Service     string            `json:"service,omitempty"`
	Environment string            `json:"environment,omitempty"`
	Checks      map[string]string `json:"checks,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, value response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
