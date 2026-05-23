package validation

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func ReadLimitedBody(w http.ResponseWriter, r *http.Request, maxBytes int64) ([]byte, *Error) {
	if maxBytes <= 0 {
		maxBytes = defaultBodyMaxBytes
	}
	if r.Body == nil {
		return nil, nil
	}
	limited := http.MaxBytesReader(w, r.Body, maxBytes)
	data, err := io.ReadAll(limited)
	_ = limited.Close()
	if err != nil {
		if isBodyTooLarge(err) {
			return nil, NewError(http.StatusRequestEntityTooLarge, FieldError{
				Field:   "body",
				Reason:  "too_large",
				Message: "Request body is too large",
			})
		}
		return nil, NewError(http.StatusBadRequest, FieldError{
			Field:   "body",
			Reason:  "read_failed",
			Message: "Request body could not be read",
		})
	}
	r.Body = io.NopCloser(bytes.NewReader(data))
	return data, nil
}

func DecodeJSONObject(data []byte, required bool) (map[string]any, *Error) {
	payload := make(map[string]any)
	if len(bytes.TrimSpace(data)) == 0 {
		if required {
			return nil, NewError(http.StatusBadRequest, FieldError{
				Field:   "body",
				Reason:  "required",
				Message: "Request body is required",
			})
		}
		return payload, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		if isBodyTooLarge(err) {
			return nil, NewError(http.StatusRequestEntityTooLarge, FieldError{
				Field:   "body",
				Reason:  "too_large",
				Message: "Request body is too large",
			})
		}
		if errors.Is(err, io.EOF) {
			return nil, NewError(http.StatusBadRequest, FieldError{
				Field:   "body",
				Reason:  "required",
				Message: "Request body is required",
			})
		}
		if strings.Contains(err.Error(), "unknown field") {
			return nil, NewError(http.StatusBadRequest, FieldError{
				Field:   "body",
				Reason:  "unknown_field",
				Message: "Request body contains an unknown field",
			})
		}
		return nil, NewError(http.StatusBadRequest, FieldError{
			Field:   "body",
			Reason:  "invalid_json",
			Message: "Request body must be a valid JSON object",
		})
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, NewError(http.StatusBadRequest, FieldError{
			Field:   "body",
			Reason:  "multiple_json_values",
			Message: "Request body must contain only one JSON value",
		})
	}
	return payload, nil
}

func isBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

func HasRequestBody(r *http.Request) bool {
	return r.Body != nil && r.ContentLength != 0
}
