package domain

type Contract struct {
	Project          string                   `json:"project"`
	Version          string                   `json:"version"`
	AuthRequirements map[AuthLevel]AuthPolicy `json:"auth_requirements"`
	RestEndpoints    []RouteDefinition        `json:"rest_endpoints"`
	Schemas          map[string]Schema        `json:"schemas"`
}

type AuthPolicy struct {
	Required    bool     `json:"required"`
	Roles       []string `json:"roles"`
	Description string   `json:"description"`
}

type Schema struct {
	Ref        string            `json:"$ref,omitempty"`
	Type       string            `json:"type,omitempty"`
	Format     string            `json:"format,omitempty"`
	Required   []string          `json:"required,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Enum       []string          `json:"enum,omitempty"`
	Minimum    *float64          `json:"minimum,omitempty"`
	Maximum    *float64          `json:"maximum,omitempty"`
	MinLength  *int              `json:"minLength,omitempty"`
	MaxLength  *int              `json:"maxLength,omitempty"`
	AllOf      []Schema          `json:"allOf,omitempty"`
}
