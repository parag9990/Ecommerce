package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Attributes map[string]any

type AttributeType string

const (
	AttributeTypeText      AttributeType = "text"
	AttributeTypeNumber    AttributeType = "number"
	AttributeTypeBoolean   AttributeType = "boolean"
	AttributeTypeEnum      AttributeType = "enum"
	AttributeTypeMultiEnum AttributeType = "multi_enum"
)

type AttributeScope string

const (
	AttributeScopeProduct AttributeScope = "product"
	AttributeScopeVariant AttributeScope = "variant"
)

type AttributeDefinition struct {
	Key        string         `json:"key"`
	Label      string         `json:"label"`
	Type       AttributeType  `json:"type"`
	Scope      AttributeScope `json:"scope"`
	Required   bool           `json:"required"`
	Filterable bool           `json:"filterable"`
	Searchable bool           `json:"searchable"`
	Values     []string       `json:"values,omitempty"`
	Unit       string         `json:"unit,omitempty"`
}

func ValidateAttributeDefinitions(schema []AttributeDefinition, field string) ValidationReport {
	var report ValidationReport
	seen := make(map[string]struct{}, len(schema))
	for index, definition := range schema {
		itemField := fmt.Sprintf("%s[%d]", field, index)
		key := strings.TrimSpace(definition.Key)
		if key == "" {
			report.AddError(CodeAttributeKeyRequired, fieldPath(itemField, "key"), "attribute key is required")
		}
		if strings.TrimSpace(definition.Label) == "" {
			report.AddError(CodeAttributeLabelRequired, fieldPath(itemField, "label"), "attribute label is required")
		}
		if !definition.Type.Valid() {
			report.AddError(CodeInvalidAttributeType, fieldPath(itemField, "type"), "attribute type is not supported")
		}
		if !definition.Scope.Valid() {
			report.AddError(CodeInvalidAttributeScope, fieldPath(itemField, "scope"), "attribute scope is not supported")
		}
		seenKey := string(definition.Scope) + ":" + key
		if key != "" {
			if _, exists := seen[seenKey]; exists {
				report.AddError(CodeDuplicateAttributeDefinition, fieldPath(itemField, "key"), "attribute key must be unique per scope")
			}
			seen[seenKey] = struct{}{}
		}
		if (definition.Type == AttributeTypeEnum || definition.Type == AttributeTypeMultiEnum) && len(definition.Values) == 0 {
			report.AddError(CodeInvalidAttributeValue, fieldPath(itemField, "values"), "enum attributes must define allowed values")
		}
	}
	return report
}

func ValidateAttributes(attrs Attributes, schema []AttributeDefinition, scope AttributeScope, options ValidationOptions, field string) ValidationReport {
	var report ValidationReport
	definitions := definitionsForScope(schema, scope)
	known := make(map[string]AttributeDefinition, len(definitions))
	for _, definition := range definitions {
		known[definition.Key] = definition
		if definition.Required {
			if value, ok := attrs[definition.Key]; !ok || isEmptyAttributeValue(value) {
				report.AddError(CodeRequiredAttributeMissing, fieldPath(field, definition.Key), "required category attribute is missing")
			}
		}
	}

	if len(schema) == 0 {
		return report
	}

	if len(attrs) == 0 {
		return report
	}

	if options.StrictAttributeSchema {
		for key := range attrs {
			if _, ok := known[key]; !ok {
				report.AddError(CodeUnknownAttributeKey, fieldPath(field, key), "attribute is not allowed by category schema")
			}
		}
	}

	for key, value := range attrs {
		definition, ok := known[key]
		if !ok {
			continue
		}
		report.Merge(validateAttributeValue(value, definition, fieldPath(field, key)))
	}

	return report
}

func (t AttributeType) Valid() bool {
	switch t {
	case AttributeTypeText, AttributeTypeNumber, AttributeTypeBoolean, AttributeTypeEnum, AttributeTypeMultiEnum:
		return true
	default:
		return false
	}
}

func (s AttributeScope) Valid() bool {
	switch s {
	case AttributeScopeProduct, AttributeScopeVariant:
		return true
	default:
		return false
	}
}

func AttributeSignature(attrs Attributes) (string, error) {
	if len(attrs) == 0 {
		return "{}", nil
	}
	normalized := make(map[string]string, len(attrs))
	for key, value := range attrs {
		normalized[key] = canonicalAttributeValue(value)
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("marshal attribute signature: %w", err)
	}
	return string(payload), nil
}

func definitionsForScope(schema []AttributeDefinition, scope AttributeScope) []AttributeDefinition {
	definitions := make([]AttributeDefinition, 0, len(schema))
	for _, definition := range schema {
		if definition.Scope == scope {
			definitions = append(definitions, definition)
		}
	}
	return definitions
}

func validateAttributeValue(value any, definition AttributeDefinition, field string) ValidationReport {
	var report ValidationReport
	switch definition.Type {
	case AttributeTypeText:
		if _, ok := value.(string); !ok {
			report.AddError(CodeInvalidAttributeType, field, "text attribute must be a string")
		}
	case AttributeTypeNumber:
		if _, ok := numberAsFloat64(value); !ok {
			report.AddError(CodeInvalidAttributeType, field, "number attribute must be numeric")
		}
	case AttributeTypeBoolean:
		if _, ok := value.(bool); !ok {
			report.AddError(CodeInvalidAttributeType, field, "boolean attribute must be true or false")
		}
	case AttributeTypeEnum:
		candidate, ok := scalarAttributeString(value)
		if !ok {
			report.AddError(CodeInvalidAttributeType, field, "enum attribute must be a scalar value")
			return report
		}
		if !allowedValue(candidate, definition.Values) {
			report.AddError(CodeInvalidAttributeValue, field, "enum attribute value is not allowed")
		}
	case AttributeTypeMultiEnum:
		values, ok := stringSliceAttribute(value)
		if !ok {
			report.AddError(CodeInvalidAttributeType, field, "multi_enum attribute must be an array of strings")
			return report
		}
		for _, candidate := range values {
			if !allowedValue(candidate, definition.Values) {
				report.AddError(CodeInvalidAttributeValue, field, "multi_enum attribute contains a value that is not allowed")
				break
			}
		}
	default:
		report.AddError(CodeInvalidAttributeType, field, "attribute type is not supported")
	}
	return report
}

func isEmptyAttributeValue(value any) bool {
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) == ""
	}
	return false
}

func allowedValue(candidate string, allowed []string) bool {
	for _, value := range allowed {
		if candidate == value {
			return true
		}
	}
	return false
}

func scalarAttributeString(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case bool:
		return strconv.FormatBool(typed), true
	case int:
		return strconv.Itoa(typed), true
	case int8:
		return strconv.FormatInt(int64(typed), 10), true
	case int16:
		return strconv.FormatInt(int64(typed), 10), true
	case int32:
		return strconv.FormatInt(int64(typed), 10), true
	case int64:
		return strconv.FormatInt(typed, 10), true
	case uint:
		return strconv.FormatUint(uint64(typed), 10), true
	case uint8:
		return strconv.FormatUint(uint64(typed), 10), true
	case uint16:
		return strconv.FormatUint(uint64(typed), 10), true
	case uint32:
		return strconv.FormatUint(uint64(typed), 10), true
	case uint64:
		return strconv.FormatUint(typed, 10), true
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 64), true
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), true
	case json.Number:
		return typed.String(), true
	default:
		return "", false
	}
}

func numberAsFloat64(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case float32:
		v := float64(typed)
		return v, !math.IsNaN(v) && !math.IsInf(v, 0)
	case float64:
		return typed, !math.IsNaN(typed) && !math.IsInf(typed, 0)
	case json.Number:
		v, err := typed.Float64()
		return v, err == nil && !math.IsNaN(v) && !math.IsInf(v, 0)
	default:
		return 0, false
	}
}

func stringSliceAttribute(value any) ([]string, bool) {
	switch typed := value.(type) {
	case []string:
		return typed, true
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil, false
			}
			values = append(values, text)
		}
		return values, true
	default:
		return nil, false
	}
}

func canonicalAttributeValue(value any) string {
	if values, ok := stringSliceAttribute(value); ok {
		payload, _ := json.Marshal(values)
		return string(payload)
	}
	if text, ok := scalarAttributeString(value); ok {
		return text
	}
	payload, _ := json.Marshal(value)
	return string(payload)
}
