package repository

import (
	"testing"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

func TestCartIndexModelsCoverDomainSpec(t *testing.T) {
	models := cartIndexModels()
	spec := domain.CartCollectionSpec()
	if len(models) != len(spec.Indexes) {
		t.Fatalf("index model count = %d, want %d", len(models), len(spec.Indexes))
	}

	got := make(map[string]struct{}, len(models))
	for _, model := range models {
		if model.Options == nil || model.Options.Name == nil {
			t.Fatalf("index model missing name: %#v", model)
		}
		got[*model.Options.Name] = struct{}{}
	}
	for _, index := range spec.Indexes {
		if _, ok := got[index.Name]; !ok {
			t.Fatalf("domain index %q missing from mongo models", index.Name)
		}
	}
}

func TestCartValidatorIsStrictJSONSchema(t *testing.T) {
	validator := cartValidator()
	if len(validator) != 1 {
		t.Fatalf("validator length = %d, want 1", len(validator))
	}
	if validator[0].Key != "$jsonSchema" {
		t.Fatalf("validator key = %q, want $jsonSchema", validator[0].Key)
	}
}

func TestNormalizeCartIDsTrimsAndDeduplicates(t *testing.T) {
	ids := normalizeCartIDs([]string{" cart_1 ", "", "cart_2", "cart_1"})
	if len(ids) != 2 || ids[0] != "cart_1" || ids[1] != "cart_2" {
		t.Fatalf("normalizeCartIDs() = %#v, want cart_1/cart_2", ids)
	}
}
