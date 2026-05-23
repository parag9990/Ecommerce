package validation

import "net/http"

const CodeValidationError = "VALIDATION_ERROR"

type FieldError struct {
	Field   string `json:"field,omitempty"`
	Reason  string `json:"reason"`
	Message string `json:"message,omitempty"`
}

type Error struct {
	Status  int
	Code    string
	Message string
	Details []FieldError
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return "validation error"
}

func (e *Error) StatusCode() int {
	if e == nil || e.Status == 0 {
		return http.StatusBadRequest
	}
	return e.Status
}

func (e *Error) ErrorCode() string {
	if e == nil || e.Code == "" {
		return CodeValidationError
	}
	return e.Code
}

func (e *Error) PublicMessage() string {
	if e == nil || e.Message == "" {
		return "Invalid request"
	}
	return e.Message
}

func (e *Error) PublicDetails() any {
	if e == nil || len(e.Details) == 0 {
		return nil
	}
	return e.Details
}

func NewError(status int, details ...FieldError) *Error {
	if status == 0 {
		status = http.StatusBadRequest
	}
	return &Error{
		Status:  status,
		Code:    CodeValidationError,
		Message: "Invalid request",
		Details: compactDetails(details),
	}
}

func (e *Error) WithDetail(detail FieldError) *Error {
	if e == nil {
		return NewError(http.StatusBadRequest, detail)
	}
	e.Details = append(e.Details, detail)
	return e
}

func compactDetails(details []FieldError) []FieldError {
	out := make([]FieldError, 0, len(details))
	for _, detail := range details {
		if detail.Field == "" && detail.Reason == "" && detail.Message == "" {
			continue
		}
		out = append(out, detail)
	}
	return out
}
