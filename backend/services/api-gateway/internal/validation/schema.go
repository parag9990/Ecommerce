package validation

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ecommerce/api-gateway/internal/domain"
)

type SchemaValidator struct {
	schemas map[string]domain.Schema
}

func NewSchemaValidator(schemas map[string]domain.Schema) *SchemaValidator {
	out := make(map[string]domain.Schema, len(schemas))
	for name, schema := range schemas {
		out[name] = schema
	}
	return &SchemaValidator{schemas: out}
}

func (v *SchemaValidator) Validate(schemaName string, payload map[string]any) *Error {
	if schemaName == "" || schemaName == "Empty" {
		if len(payload) == 0 {
			return nil
		}
		return NewError(http.StatusBadRequest, FieldError{
			Field:   "body",
			Reason:  "unknown_field",
			Message: "Request body is not allowed for this route",
		})
	}
	schema, ok := v.ResolveByName(schemaName)
	if !ok {
		return nil
	}
	details := v.validateObject("", schema, payload)
	if len(details) == 0 {
		return nil
	}
	return NewError(http.StatusBadRequest, details...)
}

func (v *SchemaValidator) ResolveByName(schemaName string) (domain.Schema, bool) {
	schema, ok := v.schemas[schemaName]
	if !ok {
		return domain.Schema{}, false
	}
	return v.resolve(schema, map[string]bool{})
}

func (v *SchemaValidator) resolve(schema domain.Schema, seen map[string]bool) (domain.Schema, bool) {
	if schema.Ref != "" {
		name := schemaNameFromRef(schema.Ref)
		if name == "" || seen[name] {
			return domain.Schema{}, false
		}
		ref, ok := v.schemas[name]
		if !ok {
			return domain.Schema{}, false
		}
		seen[name] = true
		return v.resolve(ref, seen)
	}
	if len(schema.AllOf) > 0 {
		merged := domain.Schema{Type: "object", Properties: map[string]domain.Schema{}}
		required := make(map[string]struct{})
		for _, part := range schema.AllOf {
			resolved, ok := v.resolve(part, seen)
			if !ok {
				return domain.Schema{}, false
			}
			if resolved.Type != "" && merged.Type == "" {
				merged.Type = resolved.Type
			}
			for _, field := range resolved.Required {
				required[field] = struct{}{}
			}
			for name, property := range resolved.Properties {
				merged.Properties[name] = property
			}
		}
		for field := range required {
			merged.Required = append(merged.Required, field)
		}
		return merged, true
	}
	return schema, true
}

func (v *SchemaValidator) validateObject(prefix string, schema domain.Schema, payload map[string]any) []FieldError {
	schema, _ = v.resolve(schema, map[string]bool{})
	if schema.Type != "" && schema.Type != "object" {
		return []FieldError{{
			Field:   fieldName(prefix, "body"),
			Reason:  "type",
			Message: "must be an object",
		}}
	}

	var details []FieldError
	required := make(map[string]struct{}, len(schema.Required))
	for _, field := range schema.Required {
		required[field] = struct{}{}
		value, ok := payload[field]
		if !ok || value == nil || (isStringSchema(schema.Properties[field]) && strings.TrimSpace(safeStringValue(value)) == "") {
			details = append(details, FieldError{
				Field:   fieldName(prefix, field),
				Reason:  "required",
				Message: field + " is required",
			})
		}
	}

	if schema.Properties != nil {
		for field := range payload {
			if _, ok := schema.Properties[field]; !ok {
				details = append(details, FieldError{
					Field:   fieldName(prefix, field),
					Reason:  "unknown_field",
					Message: "Field is not allowed",
				})
			}
		}
	}

	for field, property := range schema.Properties {
		value, ok := payload[field]
		if !ok || value == nil {
			continue
		}
		_, isRequired := required[field]
		details = append(details, v.validateValue(fieldName(prefix, field), field, property, value, isRequired)...)
	}
	details = append(details, validateCrossFieldRules(prefix, payload)...)
	return details
}

func (v *SchemaValidator) validateValue(path string, field string, schema domain.Schema, value any, required bool) []FieldError {
	schema, ok := v.resolve(schema, map[string]bool{})
	if !ok {
		return nil
	}
	schemaType := schema.Type
	if schemaType == "" {
		switch {
		case len(schema.Properties) > 0:
			schemaType = "object"
		case schema.Items != nil:
			schemaType = "array"
		}
	}

	switch schemaType {
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			return []FieldError{{Field: path, Reason: "type", Message: "must be an object"}}
		}
		if schema.Properties == nil {
			return nil
		}
		return v.validateObject(path, schema, object)
	case "array":
		array, ok := value.([]any)
		if !ok {
			return []FieldError{{Field: path, Reason: "type", Message: "must be an array"}}
		}
		details := validateArrayRules(path, field, array)
		if schema.Items != nil {
			for i, item := range array {
				details = append(details, v.validateValue(fmt.Sprintf("%s[%d]", path, i), field, *schema.Items, item, false)...)
			}
		}
		return details
	case "integer":
		integer, ok := jsonInteger(value)
		if !ok {
			return []FieldError{{Field: path, Reason: "type", Message: "must be an integer"}}
		}
		return validateIntegerRules(path, field, integer, schema)
	case "number":
		number, ok := jsonNumber(value)
		if !ok {
			return []FieldError{{Field: path, Reason: "type", Message: "must be a number"}}
		}
		return validateNumberRules(path, field, number, schema)
	case "boolean":
		if _, ok := value.(bool); !ok {
			return []FieldError{{Field: path, Reason: "type", Message: "must be a boolean"}}
		}
		return nil
	case "string", "":
		text, ok := value.(string)
		if !ok {
			return []FieldError{{Field: path, Reason: "type", Message: "must be a string"}}
		}
		return validateStringRules(path, field, text, schema, required)
	default:
		return nil
	}
}

func validateStringRules(path string, field string, value string, schema domain.Schema, required bool) []FieldError {
	value = strings.TrimSpace(value)
	var details []FieldError
	if required && value == "" {
		details = append(details, FieldError{Field: path, Reason: "required", Message: field + " is required"})
		return details
	}
	if value == "" {
		return nil
	}
	if schema.MinLength != nil && len(value) < *schema.MinLength {
		details = append(details, FieldError{Field: path, Reason: "min", Message: fmt.Sprintf("%s is too short", field)})
	}
	if schema.MaxLength != nil && len(value) > *schema.MaxLength {
		details = append(details, FieldError{Field: path, Reason: "max", Message: fmt.Sprintf("%s is too long", field)})
	}
	switch schema.Format {
	case "email":
		if !IsValidEmail(value) {
			details = append(details, FieldError{Field: path, Reason: "email", Message: "must be a valid email address"})
		}
	case "date-time":
		if !IsValidDateTime(value) {
			details = append(details, FieldError{Field: path, Reason: "datetime", Message: "must be a valid RFC3339 date-time"})
		}
	case "uri", "url":
		if !isHTTPURL(value) {
			details = append(details, FieldError{Field: path, Reason: "url", Message: "must be a valid http or https URL"})
		}
	}
	if len(schema.Enum) > 0 && !contains(schema.Enum, value) {
		details = append(details, FieldError{Field: path, Reason: "oneof", Message: "value is not supported"})
	}
	details = append(details, validateNamedStringRules(path, field, value)...)
	return details
}

func validateNamedStringRules(path string, field string, value string) []FieldError {
	var details []FieldError
	switch {
	case field == "email" || field == "support_email":
		if !IsValidEmail(value) {
			details = append(details, FieldError{Field: path, Reason: "email", Message: "must be a valid email address"})
		}
	case strings.HasSuffix(field, "_id") || field == "id":
		if !IsValidPublicID(value) {
			details = append(details, FieldError{Field: path, Reason: "public_id", Message: field + " format is invalid"})
		}
	case field == "idempotency_key":
		if !IsValidIdempotencyKey(value) {
			details = append(details, FieldError{Field: path, Reason: "idempotency", Message: "idempotency key format is invalid"})
		}
	case field == "currency":
		if !IsValidCurrency(value) {
			details = append(details, FieldError{Field: path, Reason: "currency", Message: "currency must be a three-letter uppercase code"})
		}
	case field == "coupon_code" || field == "code":
		if !couponCodePattern.MatchString(value) {
			details = append(details, FieldError{Field: path, Reason: "format", Message: field + " format is invalid"})
		}
	case field == "sku":
		if !skuPattern.MatchString(value) {
			details = append(details, FieldError{Field: path, Reason: "format", Message: "sku format is invalid"})
		}
	case field == "payment_provider":
		if !contains([]string{"stripe", "razorpay", "cod"}, value) {
			details = append(details, FieldError{Field: path, Reason: "oneof", Message: "payment_provider value is not supported"})
		}
	case field == "sort":
		if !contains([]string{"relevance", "price_asc", "price_desc", "newest", "rating"}, value) {
			details = append(details, FieldError{Field: path, Reason: "oneof", Message: "sort value is not supported"})
		}
	case field == "q":
		if len(value) > 120 {
			details = append(details, FieldError{Field: path, Reason: "max", Message: "q must be at most 120 characters"})
		}
	case field == "reason":
		if len(value) < 3 || len(value) > 500 {
			details = append(details, FieldError{Field: path, Reason: "range", Message: "reason length is invalid"})
		}
	case field == "title":
		if len(value) < 3 || len(value) > 180 {
			details = append(details, FieldError{Field: path, Reason: "range", Message: "title length is invalid"})
		}
	case field == "description":
		if len(value) > 5000 {
			details = append(details, FieldError{Field: path, Reason: "max", Message: "description is too long"})
		}
	case field == "status":
		if !statusValuePattern.MatchString(value) {
			details = append(details, FieldError{Field: path, Reason: "format", Message: "status format is invalid"})
		}
	}
	return details
}

func validateIntegerRules(path string, field string, value int64, schema domain.Schema) []FieldError {
	var details []FieldError
	if schema.Minimum != nil && float64(value) < *schema.Minimum {
		details = append(details, FieldError{Field: path, Reason: "min", Message: fmt.Sprintf("%s is below the minimum", field)})
	}
	if schema.Maximum != nil && float64(value) > *schema.Maximum {
		details = append(details, FieldError{Field: path, Reason: "max", Message: fmt.Sprintf("%s is above the maximum", field)})
	}
	switch field {
	case "page":
		if value < 1 {
			details = append(details, FieldError{Field: path, Reason: "min", Message: "page must be a positive integer"})
		}
	case "page_size":
		if value < 1 || value > 100 {
			details = append(details, FieldError{Field: path, Reason: "range", Message: "page_size must be between 1 and 100"})
		}
	case "limit":
		if value < 1 || value > 100 {
			details = append(details, FieldError{Field: path, Reason: "range", Message: "limit must be between 1 and 100"})
		}
	case "quantity":
		if value < 1 || value > 99 {
			details = append(details, FieldError{Field: path, Reason: "range", Message: "quantity must be between 1 and 99"})
		}
	case "stock_quantity":
		if value < 0 || value > 1000000 {
			details = append(details, FieldError{Field: path, Reason: "range", Message: "stock_quantity is out of range"})
		}
	case "amount", "discount_value", "usage_limit":
		if value < 0 {
			details = append(details, FieldError{Field: path, Reason: "min", Message: field + " must not be negative"})
		}
	}
	return details
}

func validateNumberRules(path string, field string, value float64, schema domain.Schema) []FieldError {
	var details []FieldError
	if schema.Minimum != nil && value < *schema.Minimum {
		details = append(details, FieldError{Field: path, Reason: "min", Message: fmt.Sprintf("%s is below the minimum", field)})
	}
	if schema.Maximum != nil && value > *schema.Maximum {
		details = append(details, FieldError{Field: path, Reason: "max", Message: fmt.Sprintf("%s is above the maximum", field)})
	}
	return details
}

func validateArrayRules(path string, field string, value []any) []FieldError {
	switch field {
	case "images":
		if len(value) > 20 {
			return []FieldError{{Field: path, Reason: "max", Message: "images must contain at most 20 items"}}
		}
	case "variants":
		if len(value) < 1 || len(value) > 100 {
			return []FieldError{{Field: path, Reason: "range", Message: "variants must contain between 1 and 100 items"}}
		}
	}
	return nil
}

func validateCrossFieldRules(prefix string, payload map[string]any) []FieldError {
	from := safeStringValue(payload["from"])
	to := safeStringValue(payload["to"])
	if from == "" || to == "" {
		return nil
	}
	fromTime, fromErr := time.Parse(time.RFC3339, from)
	toTime, toErr := time.Parse(time.RFC3339, to)
	if fromErr != nil || toErr != nil {
		return nil
	}
	if fromTime.After(toTime) {
		return []FieldError{{
			Field:   fieldName(prefix, "from"),
			Reason:  "date_range",
			Message: "from must be before or equal to to",
		}}
	}
	return nil
}

func jsonInteger(value any) (int64, bool) {
	switch typed := value.(type) {
	case json.Number:
		if strings.Contains(typed.String(), ".") {
			return 0, false
		}
		parsed, err := typed.Int64()
		return parsed, err == nil
	case float64:
		if math.Trunc(typed) != typed {
			return 0, false
		}
		return int64(typed), true
	case int:
		return int64(typed), true
	case int64:
		return typed, true
	default:
		return 0, false
	}
}

func jsonNumber(value any) (float64, bool) {
	switch typed := value.(type) {
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	default:
		return 0, false
	}
}

func isStringSchema(schema domain.Schema) bool {
	return schema.Type == "string"
}

func schemaNameFromRef(ref string) string {
	const prefix = "#/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return ""
	}
	return strings.TrimPrefix(ref, prefix)
}

func fieldName(prefix string, field string) string {
	if prefix == "" {
		return field
	}
	if field == "" {
		return prefix
	}
	return prefix + "." + field
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func isHTTPURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func queryInteger(raw string) (json.Number, bool) {
	if _, err := strconv.ParseInt(raw, 10, 64); err != nil {
		return "", false
	}
	return json.Number(raw), true
}
