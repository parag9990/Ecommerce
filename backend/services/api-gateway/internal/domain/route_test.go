package domain

import "testing"

func TestRouteDefinitionValidate(t *testing.T) {
	route := RouteDefinition{
		ID:             "product.detail",
		Method:         MethodGet,
		Path:           "/api/v1/products/{product_id}",
		Service:        "product-service",
		GRPCMethod:     "ProductService.GetProduct",
		AuthLevel:      AuthPublic,
		RequestSchema:  "IdPathRequest",
		ResponseSchema: "Product",
	}

	if err := route.Validate("/api/v1"); err != nil {
		t.Fatalf("expected route to validate: %v", err)
	}
}

func TestRouteDefinitionValidateRejectsInvalidPathParameter(t *testing.T) {
	route := RouteDefinition{
		ID:             "product.detail",
		Method:         MethodGet,
		Path:           "/api/v1/products/{product-id}",
		Service:        "product-service",
		GRPCMethod:     "ProductService.GetProduct",
		AuthLevel:      AuthPublic,
		RequestSchema:  "IdPathRequest",
		ResponseSchema: "Product",
	}

	if err := route.Validate("/api/v1"); err == nil {
		t.Fatal("expected invalid path parameter to fail validation")
	}
}

func TestAllowedRolesReturnsCopy(t *testing.T) {
	route := RouteDefinition{AuthLevel: AuthBuyer}

	roles := route.AllowedRoles()
	roles[0] = "mutated"

	fresh := route.AllowedRoles()
	if fresh[0] != "buyer" {
		t.Fatalf("expected roles to be immutable copy, got %q", fresh[0])
	}
}
