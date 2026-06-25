package httptransport

import (
	"net/http"
	"strings"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/domain"
)

type sellerSessionResponse struct {
	Authenticated bool            `json:"authenticated"`
	ActiveSeller  *sellerSummary  `json:"active_seller"`
	Sellers       []sellerSummary `json:"sellers"`
}

type sellerSummary struct {
	SellerID    string   `json:"seller_id"`
	DisplayName string   `json:"display_name"`
	Status      string   `json:"status"`
	Role        string   `json:"role,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	StaffRole   string   `json:"staff_role,omitempty"`
}

func sellerSessionRouteEndpoint(route domain.RouteDefinition, fallback http.HandlerFunc) http.HandlerFunc {
	if string(route.Method) != http.MethodGet || route.Path != "/api/v1/seller/session" {
		return fallback
	}
	return handleSellerSession
}

func handleSellerSession(w http.ResponseWriter, r *http.Request) {
	claims, ok := gatewayauth.ClaimsFromContext(r.Context())
	if !ok {
		writeSuccess(w, r, http.StatusOK, sellerSessionResponse{
			Authenticated: false,
			ActiveSeller:  nil,
			Sellers:       []sellerSummary{},
		})
		return
	}

	sellerID := strings.TrimSpace(claims.SellerID)
	if sellerID == "" {
		writeError(w, r, http.StatusForbidden, "SELLER_CONTEXT_REQUIRED", "Seller context is required")
		return
	}

	seller := sellerSummary{
		SellerID:    sellerID,
		DisplayName: sellerID,
		Status:      "active",
		Role:        primarySellerRole(claims.Roles),
		Roles:       cloneSellerSessionRoles(claims.Roles),
		StaffRole:   primarySellerRole(claims.Roles),
	}
	writeSuccess(w, r, http.StatusOK, sellerSessionResponse{
		Authenticated: true,
		ActiveSeller:  &seller,
		Sellers:       []sellerSummary{seller},
	})
}

func primarySellerRole(roles []string) string {
	for _, role := range roles {
		switch strings.TrimSpace(role) {
		case "seller_manager", "seller_catalog_editor", "seller_order_manager", "seller", "superadmin":
			return strings.TrimSpace(role)
		}
	}
	return ""
}

func cloneSellerSessionRoles(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			out = append(out, value)
		}
	}
	return out
}
