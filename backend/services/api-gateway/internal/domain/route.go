package domain

import (
	"fmt"
	"net/http"
	"strings"
	"unicode"
)

type HTTPMethod string

const (
	MethodGet    HTTPMethod = http.MethodGet
	MethodPost   HTTPMethod = http.MethodPost
	MethodPatch  HTTPMethod = http.MethodPatch
	MethodDelete HTTPMethod = http.MethodDelete
)

type AuthLevel string

const (
	AuthPublic     AuthLevel = "public"
	AuthBuyer      AuthLevel = "buyer"
	AuthSeller     AuthLevel = "seller"
	AuthAdmin      AuthLevel = "admin"
	AuthSuperadmin AuthLevel = "superadmin"
	AuthWebhook    AuthLevel = "webhook"
)

type RouteDefinition struct {
	ID             string     `json:"id"`
	Method         HTTPMethod `json:"method"`
	Path           string     `json:"path"`
	Service        string     `json:"service"`
	GRPCMethod     string     `json:"grpc"`
	AuthLevel      AuthLevel  `json:"auth"`
	RequestSchema  string     `json:"request_schema"`
	ResponseSchema string     `json:"response_schema"`
}

func (r RouteDefinition) Key() string {
	return string(r.Method) + " " + r.Path
}

func (r RouteDefinition) Pattern() string {
	return r.Key()
}

func (r RouteDefinition) RequiresToken() bool {
	return r.AuthLevel != AuthPublic && r.AuthLevel != AuthWebhook
}

func (r RouteDefinition) AllowedRoles() []string {
	roles := rolesByAuthLevel[r.AuthLevel]
	out := make([]string, len(roles))
	copy(out, roles)
	return out
}

func (r RouteDefinition) Group(basePath string) string {
	path := strings.TrimPrefix(r.Path, basePath)
	path = strings.Trim(path, "/")
	if path == "" {
		return "root"
	}
	segment := strings.Split(path, "/")[0]
	switch segment {
	case "sellers", "seller":
		return "seller"
	case "webhooks":
		return "webhook"
	default:
		return segment
	}
}

func (r RouteDefinition) Validate(basePath string) error {
	var errs []error
	if strings.TrimSpace(r.ID) == "" {
		errs = append(errs, ValidationError{Field: "id", Message: "route id is required"})
	} else if strings.HasPrefix(r.ID, ".") || strings.HasSuffix(r.ID, ".") || !strings.Contains(r.ID, ".") {
		errs = append(errs, ValidationError{Field: "id", Message: "route id must use domain.action format"})
	}
	if !IsSupportedHTTPMethod(r.Method) {
		errs = append(errs, ValidationError{Field: "method", Message: fmt.Sprintf("unsupported method %q", r.Method)})
	}
	if err := ValidatePathTemplate(r.Path, basePath); err != nil {
		errs = append(errs, err)
	}
	if strings.TrimSpace(r.Service) == "" {
		errs = append(errs, ValidationError{Field: "service", Message: "target service is required"})
	} else if !strings.HasSuffix(r.Service, "-service") {
		errs = append(errs, ValidationError{Field: "service", Message: "target service must use service-name-service format"})
	}
	if !isValidGRPCMethod(r.GRPCMethod) {
		errs = append(errs, ValidationError{Field: "grpc", Message: "grpc method must use Service.Method format"})
	}
	if !IsSupportedAuthLevel(r.AuthLevel) {
		errs = append(errs, ValidationError{Field: "auth", Message: fmt.Sprintf("unsupported auth level %q", r.AuthLevel)})
	}
	if strings.TrimSpace(r.RequestSchema) == "" {
		errs = append(errs, ValidationError{Field: "request_schema", Message: "request schema is required"})
	}
	if strings.TrimSpace(r.ResponseSchema) == "" {
		errs = append(errs, ValidationError{Field: "response_schema", Message: "response schema is required"})
	}
	if len(errs) > 0 {
		return ContractValidationError{Errors: errs}
	}
	return nil
}

func IsSupportedHTTPMethod(method HTTPMethod) bool {
	switch method {
	case MethodGet, MethodPost, MethodPatch, MethodDelete:
		return true
	default:
		return false
	}
}

func IsSupportedAuthLevel(level AuthLevel) bool {
	_, ok := rolesByAuthLevel[level]
	return ok
}

func ValidatePathTemplate(path string, basePath string) error {
	if strings.TrimSpace(path) == "" {
		return ValidationError{Field: "path", Message: "path is required"}
	}
	if !strings.HasPrefix(path, "/") {
		return ValidationError{Field: "path", Message: "path must start with /"}
	}
	if strings.ContainsAny(path, " \t\r\n") {
		return ValidationError{Field: "path", Message: "path must not contain whitespace"}
	}
	if basePath != "" && !strings.HasPrefix(path, basePath) {
		return ValidationError{Field: "path", Message: fmt.Sprintf("path must start with %s", basePath)}
	}
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		return ValidationError{Field: "path", Message: "path must not end with /"}
	}
	seenParams := make(map[string]struct{})
	for _, segment := range strings.Split(strings.Trim(path, "/"), "/") {
		if segment == "" {
			return ValidationError{Field: "path", Message: "path must not contain empty segments"}
		}
		if strings.Contains(segment, "{") || strings.Contains(segment, "}") {
			if !strings.HasPrefix(segment, "{") || !strings.HasSuffix(segment, "}") || strings.Count(segment, "{") != 1 || strings.Count(segment, "}") != 1 {
				return ValidationError{Field: "path", Message: "path parameters must occupy the full segment"}
			}
			name := strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")
			if !isValidPathParam(name) {
				return ValidationError{Field: "path", Message: fmt.Sprintf("invalid path parameter %q", name)}
			}
			if _, exists := seenParams[name]; exists {
				return ValidationError{Field: "path", Message: fmt.Sprintf("duplicate path parameter %q", name)}
			}
			seenParams[name] = struct{}{}
		}
	}
	return nil
}

func isValidPathParam(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if r == '_' {
			continue
		}
		if i == 0 && unicode.IsDigit(r) {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isValidGRPCMethod(method string) bool {
	parts := strings.Split(method, ".")
	return len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != ""
}

var rolesByAuthLevel = map[AuthLevel][]string{
	AuthPublic:     {},
	AuthBuyer:      {"buyer", "seller", "admin", "superadmin"},
	AuthSeller:     {"seller", "seller_manager", "seller_catalog_editor", "seller_order_manager", "superadmin"},
	AuthAdmin:      {"admin", "operations_admin", "finance_admin", "catalog_admin", "superadmin"},
	AuthSuperadmin: {"superadmin"},
	AuthWebhook:    {},
}
