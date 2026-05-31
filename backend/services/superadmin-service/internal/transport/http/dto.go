package http

import "ecommerce/superadmin-service/internal/domain"

type permissionDTO struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Risk        string `json:"risk"`
}

type permissionCatalogResponse struct {
	Permissions []permissionDTO `json:"permissions"`
}

type rolePermissionsResponse struct {
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type errorResponse struct {
	Code               string `json:"code"`
	Message            string `json:"message"`
	RequiredPermission string `json:"required_permission,omitempty"`
	RequestID          string `json:"request_id,omitempty"`
}

func permissionDTOs(definitions []domain.PermissionDefinition) []permissionDTO {
	out := make([]permissionDTO, 0, len(definitions))
	for _, definition := range definitions {
		out = append(out, permissionDTO{
			Key:         string(definition.Key),
			Description: definition.Description,
			Risk:        string(definition.Risk),
		})
	}
	return out
}

func permissionsToStrings(permissions []domain.Permission) []string {
	out := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		out = append(out, string(permission))
	}
	return out
}
