package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
	"time"
)

const HeaderRequestID = "X-Request-Id"

type requestIDContextKey struct{}

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_:\-]{7,127}$`)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDContextKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func ValidRequestID(requestID string) bool {
	return validRequestID.MatchString(requestID)
}

func NewRequestID() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "req_" + time.Now().UTC().Format("20060102150405000000000")
	}
	return "req_" + hex.EncodeToString(buf[:])
}

func RequestIDMiddleware(header string) func(http.Handler) http.Handler {
	if header == "" {
		header = HeaderRequestID
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(header)
			if !ValidRequestID(requestID) {
				requestID = NewRequestID()
			}
			w.Header().Set(header, requestID)
			next.ServeHTTP(w, r.WithContext(WithRequestID(r.Context(), requestID)))
		})
	}
}
