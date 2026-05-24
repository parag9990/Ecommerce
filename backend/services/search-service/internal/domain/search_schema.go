package domain

import (
	"fmt"
	"strings"
)

const (
	ProductsCollectionName       = "products"
	DefaultProductSortingField   = ProductFieldPopularityScore
	TypesenseTypeString          = "string"
	TypesenseTypeStringArray     = "string[]"
	TypesenseTypeFloat           = "float"
	TypesenseTypeInt32           = "int32"
	TypesenseTypeInt64           = "int64"
	TypesenseTypeBool            = "bool"
	ProductSchemaContractVersion = "products-v1"
)

type CollectionSchema struct {
	Name                string        `json:"name"`
	Fields              []SchemaField `json:"fields"`
	DefaultSortingField string        `json:"default_sorting_field"`
}

type SchemaField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Facet    bool   `json:"facet,omitempty"`
	Sort     bool   `json:"sort,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

func (s CollectionSchema) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("%w: collection name is required", ErrInvalidSearchSchema)
	}
	if len(s.Fields) == 0 {
		return fmt.Errorf("%w: at least one field is required", ErrInvalidSearchSchema)
	}

	seen := make(map[string]SchemaField, len(s.Fields))
	for _, field := range s.Fields {
		if strings.TrimSpace(field.Name) == "" {
			return fmt.Errorf("%w: field name is required", ErrInvalidSearchSchema)
		}
		if strings.TrimSpace(field.Type) == "" {
			return fmt.Errorf("%w: field %q type is required", ErrInvalidSearchSchema, field.Name)
		}
		if _, ok := seen[field.Name]; ok {
			return fmt.Errorf("%w: duplicate field %q", ErrInvalidSearchSchema, field.Name)
		}
		seen[field.Name] = field
	}

	defaultField, ok := seen[s.DefaultSortingField]
	if strings.TrimSpace(s.DefaultSortingField) == "" || !ok {
		return fmt.Errorf("%w: default sorting field %q is not defined", ErrInvalidSearchSchema, s.DefaultSortingField)
	}
	if !defaultField.Sort {
		return fmt.Errorf("%w: default sorting field %q must be sortable", ErrInvalidSearchSchema, s.DefaultSortingField)
	}
	return nil
}

func (s CollectionSchema) FacetFields() []string {
	fields := make([]string, 0, len(s.Fields))
	for _, field := range s.Fields {
		if field.Facet {
			fields = append(fields, field.Name)
		}
	}
	return fields
}

func (s CollectionSchema) SortFields() []string {
	fields := make([]string, 0, len(s.Fields))
	for _, field := range s.Fields {
		if field.Sort {
			fields = append(fields, field.Name)
		}
	}
	return fields
}

func (s CollectionSchema) TypesenseDefinition() map[string]any {
	fields := make([]map[string]any, 0, len(s.Fields))
	for _, field := range s.Fields {
		entry := map[string]any{
			"name": field.Name,
			"type": field.Type,
		}
		if field.Facet {
			entry["facet"] = true
		}
		if field.Sort {
			entry["sort"] = true
		}
		if field.Optional {
			entry["optional"] = true
		}
		fields = append(fields, entry)
	}

	return map[string]any{
		"name":                  s.Name,
		"fields":                fields,
		"default_sorting_field": s.DefaultSortingField,
	}
}
