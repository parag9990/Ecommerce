package validation

import (
	"net/http"
	"strings"

	"ecommerce/api-gateway/internal/domain"
)

func ValidatePathParams(r *http.Request, route domain.RouteDefinition) *Error {
	for _, name := range pathParamNames(route.Path) {
		value := strings.TrimSpace(r.PathValue(name))
		if value == "" {
			return NewError(http.StatusBadRequest, FieldError{
				Field:   name,
				Reason:  "required",
				Message: name + " is required",
			})
		}
		if !validPathParamValue(name, value) {
			return NewError(http.StatusBadRequest, FieldError{
				Field:   name,
				Reason:  "invalid",
				Message: name + " format is invalid",
			})
		}
	}
	return nil
}

func pathParamNames(path string) []string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	out := make([]string, 0, len(segments))
	for _, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")
			if name != "" {
				out = append(out, name)
			}
		}
	}
	return out
}

func validPathParamValue(name string, value string) bool {
	switch name {
	case "provider":
		return IsValidProvider(value)
	case "key":
		return IsValidSettingKey(value)
	default:
		if strings.HasSuffix(name, "_id") || name == "id" {
			return IsValidPublicID(value)
		}
		return IsValidPublicID(value)
	}
}
