package gatewayerrors

import (
	"fmt"
	"net/http"
)

type MappedError struct {
	Status            int
	Code              string
	Message           string
	Details           any
	Headers           map[string]string
	GRPCCode          string
	DownstreamService string
	Internal          error
}

type LocalError interface {
	error
	StatusCode() int
	ErrorCode() string
	PublicMessage() string
	PublicDetails() any
}

type HeaderProvider interface {
	ResponseHeaders() map[string]string
}

type PublicError struct {
	status  int
	code    string
	message string
	details any
	headers map[string]string
	cause   error
}

func New(status int, code string, message string, details any, headers map[string]string) *PublicError {
	return Wrap(status, code, message, details, headers, nil)
}

func Wrap(status int, code string, message string, details any, headers map[string]string, cause error) *PublicError {
	return &PublicError{
		status:  normalizeStatus(status),
		code:    normalizeCode(code),
		message: normalizeMessage(message, fallbackMessage(code)),
		details: details,
		headers: copyHeaders(headers),
		cause:   cause,
	}
}

func (e *PublicError) Error() string {
	if e == nil {
		return ""
	}
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.code, e.cause)
	}
	if e.message != "" {
		return fmt.Sprintf("%s: %s", e.code, e.message)
	}
	return e.code
}

func (e *PublicError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *PublicError) StatusCode() int {
	if e == nil {
		return http.StatusInternalServerError
	}
	return normalizeStatus(e.status)
}

func (e *PublicError) ErrorCode() string {
	if e == nil {
		return CodeInternal
	}
	return normalizeCode(e.code)
}

func (e *PublicError) PublicMessage() string {
	if e == nil {
		return fallbackMessage(CodeInternal)
	}
	return normalizeMessage(e.message, fallbackMessage(e.code))
}

func (e *PublicError) PublicDetails() any {
	if e == nil {
		return nil
	}
	return e.details
}

func (e *PublicError) ResponseHeaders() map[string]string {
	if e == nil {
		return nil
	}
	return copyHeaders(e.headers)
}

func normalizeStatus(status int) int {
	if status <= 0 {
		return http.StatusInternalServerError
	}
	return status
}

func normalizeCode(code string) string {
	if code == "" {
		return CodeInternal
	}
	return code
}

func normalizeMessage(message string, fallback string) string {
	if message == "" {
		return fallback
	}
	return message
}

func copyHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	out := make(map[string]string, len(headers))
	for key, value := range headers {
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
