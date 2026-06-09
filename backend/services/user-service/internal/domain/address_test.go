package domain

import (
	"errors"
	"testing"
	"time"
)

func TestAddressSoftDeleteClearsDefault(t *testing.T) {
	address, err := NewAddress(NewAddressParams{
		AddressID:  "addr_123",
		UserID:     "user_123",
		Name:       "Aarav Sharma",
		Phone:      strptr("+919999999999"),
		Line1:      "221B Baker Street",
		City:       "Mumbai",
		State:      "Maharashtra",
		PostalCode: "400001",
		Country:    "India",
		IsDefault:  true,
		CreatedBy:  "user_123",
		CreatedAt:  fixedTime(),
	})
	if err != nil {
		t.Fatalf("NewAddress returned error: %v", err)
	}

	if err := address.MarkDeleted("user_123", fixedTime().Add(time.Minute)); err != nil {
		t.Fatalf("MarkDeleted returned error: %v", err)
	}

	if !address.IsDeleted() {
		t.Fatal("expected address to be deleted")
	}
	if address.IsDefault {
		t.Fatal("expected deleted address to stop being default")
	}
}

func TestDeletedAddressCannotBePatched(t *testing.T) {
	address, err := NewAddress(NewAddressParams{
		AddressID:  "addr_123",
		UserID:     "user_123",
		Name:       "Aarav Sharma",
		Line1:      "221B Baker Street",
		City:       "Mumbai",
		State:      "Maharashtra",
		PostalCode: "400001",
		Country:    "India",
		CreatedBy:  "user_123",
		CreatedAt:  fixedTime(),
	})
	if err != nil {
		t.Fatalf("NewAddress returned error: %v", err)
	}

	if err := address.MarkDeleted("user_123", fixedTime().Add(time.Minute)); err != nil {
		t.Fatalf("MarkDeleted returned error: %v", err)
	}

	err = address.ApplyPatch(AddressPatch{
		City:      strptr("Pune"),
		UpdatedBy: "user_123",
		UpdatedAt: fixedTime().Add(2 * time.Minute),
	})
	if !errors.Is(err, ErrDeletedResource) {
		t.Fatalf("expected deleted resource error, got %v", err)
	}
}
