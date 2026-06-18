package validation

import (
	"encoding/json"
	"net/http"
	"strings"

	"ecommerce/api-gateway/internal/domain"
)

func ValidateQuery(r *http.Request, route domain.RouteDefinition, schemaValidator *SchemaValidator, maxQueryBytes int) (map[string]any, *Error) {
	if maxQueryBytes <= 0 {
		maxQueryBytes = defaultQueryMaxBytes
	}
	if len(r.URL.RawQuery) > maxQueryBytes {
		return nil, NewError(http.StatusBadRequest, FieldError{
			Field:   "query",
			Reason:  "too_large",
			Message: "Query string is too large",
		})
	}
	values := r.URL.Query()
	if len(values) == 0 {
		return map[string]any{}, nil
	}
	if route.Method != domain.MethodGet {
		return nil, NewError(http.StatusBadRequest, FieldError{
			Field:   "query",
			Reason:  "not_allowed",
			Message: "Query parameters are not allowed for this route",
		})
	}
	if schemaValidator == nil || route.RequestSchema == "" || route.RequestSchema == "Empty" {
		return nil, NewError(http.StatusBadRequest, FieldError{
			Field:   firstQueryKey(values),
			Reason:  "unknown_field",
			Message: "Query parameter is not allowed",
		})
	}
	schema, ok := schemaValidator.ResolveByName(route.RequestSchema)
	if !ok {
		return map[string]any{}, nil
	}
	payload, err := queryPayload(values, schema)
	if err != nil {
		return nil, err
	}
	if verr := schemaValidator.Validate(route.RequestSchema, payload); verr != nil {
		return nil, verr
	}
	return payload, nil
}

func queryPayload(values map[string][]string, schema domain.Schema) (map[string]any, *Error) {
	payload := make(map[string]any, len(values))
	for field, rawValues := range values {
		property, ok := schema.Properties[field]
		if !ok {
			return nil, NewError(http.StatusBadRequest, FieldError{
				Field:   field,
				Reason:  "unknown_field",
				Message: "Query parameter is not allowed",
			})
		}
		value, err := queryValue(field, rawValues, property)
		if err != nil {
			return nil, err
		}
		payload[field] = value
	}
	return payload, nil
}

func queryValue(field string, rawValues []string, schema domain.Schema) (any, *Error) {
	if len(rawValues) == 0 {
		return "", nil
	}
	raw := strings.TrimSpace(rawValues[0])
	switch schema.Type {
	case "integer":
		parsed, ok := queryInteger(raw)
		if !ok {
			return nil, NewError(http.StatusBadRequest, FieldError{
				Field:   field,
				Reason:  "invalid",
				Message: field + " must be an integer",
			})
		}
		return parsed, nil
	case "boolean":
		switch strings.ToLower(raw) {
		case "true", "1", "yes":
			return true, nil
		case "false", "0", "no":
			return false, nil
		default:
			return nil, NewError(http.StatusBadRequest, FieldError{
				Field:   field,
				Reason:  "invalid",
				Message: field + " must be a boolean",
			})
		}
	case "array":
		out := make([]any, 0, len(rawValues))
		for _, rawItem := range rawValues {
			out = append(out, strings.TrimSpace(rawItem))
		}
		return out, nil
	case "object":
		if raw == "" {
			return map[string]any{}, nil
		}
		var object map[string]any
		if err := json.Unmarshal([]byte(raw), &object); err != nil {
			return nil, NewError(http.StatusBadRequest, FieldError{
				Field:   field,
				Reason:  "invalid_json",
				Message: field + " must be a JSON object",
			})
		}
		return object, nil
	default:
		return raw, nil
	}
}

func firstQueryKey(values map[string][]string) string {
	for key := range values {
		return key
	}
	return "query"
}
