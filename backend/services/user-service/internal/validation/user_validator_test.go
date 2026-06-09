package validation

import "testing"

func TestUserValidatorNormalizesCreateUser(t *testing.T) {
	validator := NewUserValidator(Config{PhoneRegion: "IN"})

	got, validationErr := validator.NormalizeAndValidateCreateUser(CreateUserInput{
		AuthAccountID: " auth_123 ",
		Email:         " BUYER@EXAMPLE.COM ",
		Phone:         "99999 99999",
		FullName:      "  Aarav   Sharma ",
	})
	if validationErr.HasErrors() {
		t.Fatalf("unexpected validation error: %#v", validationErr.Fields)
	}
	if got.Email != "buyer@example.com" {
		t.Fatalf("Email = %q, want buyer@example.com", got.Email)
	}
	if got.Phone != "+919999999999" {
		t.Fatalf("Phone = %q, want +919999999999", got.Phone)
	}
	if got.FullName != "Aarav Sharma" {
		t.Fatalf("FullName = %q, want Aarav Sharma", got.FullName)
	}
}

func TestUserValidatorRejectsAddressPostalCode(t *testing.T) {
	validator := NewUserValidator(Config{PhoneRegion: "IN"})

	_, validationErr := validator.NormalizeAndValidateAddress(AddressInput{
		Name:       "Aarav Sharma",
		Line1:      "221B MG Road",
		City:       "Bengaluru",
		State:      "Karnataka",
		PostalCode: "012345",
		Country:    "IN",
	})
	if !validationErr.HasErrors() {
		t.Fatal("expected validation errors")
	}
}

func TestUserValidatorRejectsSellerGSTIN(t *testing.T) {
	validator := NewUserValidator(Config{PhoneRegion: "IN"})

	_, validationErr := validator.NormalizeAndValidateUpdateSellerProfile(UpdateSellerProfileInput{
		GSTNumber: "99AAAAA0000A1Z5",
	}, SellerUpdateFields{GSTNumber: true})
	if !validationErr.HasErrors() {
		t.Fatal("expected validation errors")
	}
}
