package validation

import (
	"mime"
	"net/http"
	"strings"
)

const IdempotencyKeyHeader = "Idempotency-Key"

func ValidateHeaders(r *http.Request, maxHeaderBytes int64) *Error {
	if maxHeaderBytes <= 0 {
		maxHeaderBytes = defaultHeaderMaxBytes
	}
	var total int64
	for name, values := range r.Header {
		total += int64(len(name))
		if strings.ContainsAny(name, " \t\r\n") {
			return NewError(http.StatusBadRequest, FieldError{
				Field:   name,
				Reason:  "invalid",
				Message: "Header name is invalid",
			})
		}
		for _, value := range values {
			total += int64(len(value))
			if len(value) > 8*1024 {
				return NewError(http.StatusBadRequest, FieldError{
					Field:   name,
					Reason:  "too_large",
					Message: "Header value is too large",
				})
			}
		}
	}
	if total > maxHeaderBytes {
		return NewError(http.StatusRequestHeaderFieldsTooLarge, FieldError{
			Field:   "headers",
			Reason:  "too_large",
			Message: "Request headers are too large",
		})
	}
	return nil
}

func RequireContentType(r *http.Request, allowed []string) *Error {
	if len(allowed) == 0 {
		return nil
	}
	raw := strings.TrimSpace(r.Header.Get("Content-Type"))
	if raw == "" {
		return NewError(http.StatusUnsupportedMediaType, FieldError{
			Field:   "Content-Type",
			Reason:  "required",
			Message: "Content-Type header is required",
		})
	}
	mediaType, _, err := mime.ParseMediaType(raw)
	if err != nil {
		return NewError(http.StatusUnsupportedMediaType, FieldError{
			Field:   "Content-Type",
			Reason:  "invalid",
			Message: "Content-Type header is invalid",
		})
	}
	for _, item := range allowed {
		if mediaType == item {
			return nil
		}
	}
	return NewError(http.StatusUnsupportedMediaType, FieldError{
		Field:   "Content-Type",
		Reason:  "unsupported",
		Message: "Only supported request content types are allowed",
	})
}

func ValidateIdempotencyKey(r *http.Request, bodyValue string, required bool) (string, *Error) {
	headerValue := strings.TrimSpace(r.Header.Get(IdempotencyKeyHeader))
	bodyValue = strings.TrimSpace(bodyValue)
	if headerValue != "" && bodyValue != "" && headerValue != bodyValue {
		return "", NewError(http.StatusBadRequest, FieldError{
			Field:   IdempotencyKeyHeader,
			Reason:  "mismatch",
			Message: "Header and body idempotency key must match",
		})
	}

	key := headerValue
	if key == "" {
		key = bodyValue
	}
	if required && key == "" {
		return "", NewError(http.StatusBadRequest, FieldError{
			Field:   IdempotencyKeyHeader,
			Reason:  "required",
			Message: "Idempotency key is required",
		})
	}
	if key != "" && !IsValidIdempotencyKey(key) {
		return "", NewError(http.StatusBadRequest, FieldError{
			Field:   IdempotencyKeyHeader,
			Reason:  "invalid",
			Message: "Idempotency key format is invalid",
		})
	}
	return key, nil
}
