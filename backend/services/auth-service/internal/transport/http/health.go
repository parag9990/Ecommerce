package httptransport

import (
	"context"
	"net/http"
	"time"
)

const readinessTimeout = 2 * time.Second

type ReadinessDependency struct {
	Name  string
	Check func(context.Context) error
}

func readinessHandler(dependencies []ReadinessDependency) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodGet) {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
		defer cancel()

		statuses := make(map[string]string, len(dependencies))
		ready := true
		for _, dependency := range dependencies {
			if dependency.Name == "" || dependency.Check == nil {
				continue
			}
			if err := dependency.Check(ctx); err != nil {
				statuses[dependency.Name] = "unavailable"
				ready = false
				continue
			}
			statuses[dependency.Name] = "ready"
		}

		if !ready {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status":       "unavailable",
				"dependencies": statuses,
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":       "ready",
			"dependencies": statuses,
		})
	}
}
