package gatewayerrors

import (
	"strconv"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type FieldDetail struct {
	Field   string `json:"field,omitempty"`
	Reason  string `json:"reason"`
	Message string `json:"message,omitempty"`
}

type PreconditionDetail struct {
	Type        string `json:"type"`
	Subject     string `json:"subject,omitempty"`
	Description string `json:"description,omitempty"`
}

type ErrorInfoDetail struct {
	Reason   string            `json:"reason"`
	Domain   string            `json:"domain,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func extractDetails(st *status.Status) (any, map[string]string) {
	headers := map[string]string{}
	fieldDetails := make([]FieldDetail, 0)
	preconditions := make([]PreconditionDetail, 0)
	errorInfos := make([]ErrorInfoDetail, 0)
	var retryAfterSeconds int64

	for _, detail := range st.Details() {
		switch d := detail.(type) {
		case *errdetails.BadRequest:
			for _, violation := range d.GetFieldViolations() {
				fieldDetails = append(fieldDetails, FieldDetail{
					Field:   sanitizeIdentifier(violation.GetField(), 128),
					Reason:  normalizeReason(violation.GetReason(), "invalid"),
					Message: sanitizePublicMessage(fieldViolationMessage(violation), "Invalid field"),
				})
			}
		case *errdetails.ErrorInfo:
			info := ErrorInfoDetail{
				Reason:   normalizeReason(d.GetReason(), "error"),
				Domain:   sanitizeIdentifier(d.GetDomain(), 128),
				Metadata: sanitizeMetadata(d.GetMetadata()),
			}
			errorInfos = append(errorInfos, info)
		case *errdetails.PreconditionFailure:
			for _, violation := range d.GetViolations() {
				preconditions = append(preconditions, PreconditionDetail{
					Type:        sanitizeIdentifier(violation.GetType(), 80),
					Subject:     sanitizePublicMessage(violation.GetSubject(), ""),
					Description: sanitizePublicMessage(violation.GetDescription(), "Precondition failed"),
				})
			}
		case *errdetails.RetryInfo:
			if delay := d.GetRetryDelay(); delay != nil {
				seconds := int64(delay.AsDuration().Seconds())
				if delay.AsDuration() > 0 && seconds == 0 {
					seconds = 1
				}
				if seconds > 0 {
					retryAfterSeconds = seconds
					headers["Retry-After"] = strconv.FormatInt(seconds, 10)
				}
			}
		}
	}

	if len(fieldDetails) > 0 {
		return fieldDetails, headers
	}

	details := make(map[string]any)
	if len(preconditions) > 0 {
		details["preconditions"] = preconditions
	}
	if len(errorInfos) > 0 {
		details["error_info"] = errorInfos
	}
	if retryAfterSeconds > 0 {
		details["retry_after_seconds"] = retryAfterSeconds
	}
	if len(details) > 0 {
		return details, headers
	}
	return nil, headers
}

func fieldViolationMessage(violation *errdetails.BadRequest_FieldViolation) string {
	if violation == nil {
		return ""
	}
	if localized := violation.GetLocalizedMessage(); localized != nil && strings.TrimSpace(localized.GetMessage()) != "" {
		return localized.GetMessage()
	}
	return violation.GetDescription()
}

func sanitizeMetadata(metadata map[string]string) map[string]string {
	if len(metadata) == 0 {
		return nil
	}
	out := make(map[string]string, len(metadata))
	for key, value := range metadata {
		key = sanitizeIdentifier(key, 80)
		if key == "" || unsafeKey(key) {
			continue
		}
		value = sanitizePublicMessage(value, "")
		if value == "" {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sanitizeDetails(details any) any {
	switch value := details.(type) {
	case nil:
		return nil
	case []FieldDetail:
		if len(value) == 0 {
			return nil
		}
		out := make([]FieldDetail, 0, len(value))
		for _, detail := range value {
			out = append(out, FieldDetail{
				Field:   sanitizeIdentifier(detail.Field, 128),
				Reason:  normalizeReason(detail.Reason, "invalid"),
				Message: sanitizePublicMessage(detail.Message, ""),
			})
		}
		return out
	default:
		return value
	}
}
