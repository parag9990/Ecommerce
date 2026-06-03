package validation

import sharedvalidation "github.com/parag/ecommerce/backend/shared/validation"

type Config struct {
	PhoneRegion string
}

type UserValidator struct {
	phoneRegion string
}

type CreateUserInput struct {
	AuthAccountID string
	Email         string
	Phone         string
	FullName      string
}

type UpdateProfileInput struct {
	FullName  string
	Phone     string
	AvatarURL string
}

type ProfileUpdateFields struct {
	FullName  bool
	Phone     bool
	AvatarURL bool
}

type AddressInput struct {
	Name       string
	Phone      string
	Line1      string
	Line2      string
	City       string
	State      string
	PostalCode string
	Country    string
}

type UpdateSellerProfileInput struct {
	StoreName    string
	DisplayName  string
	GSTNumber    string
	SupportEmail string
}

type SellerUpdateFields struct {
	StoreName    bool
	DisplayName  bool
	GSTNumber    bool
	SupportEmail bool
}

func NewUserValidator(config Config) UserValidator {
	region := config.PhoneRegion
	if region == "" {
		region = sharedvalidation.DefaultPhoneRegion
	}
	return UserValidator{phoneRegion: region}
}

func (v UserValidator) NormalizeAndValidateCreateUser(input CreateUserInput) (CreateUserInput, sharedvalidation.Error) {
	var errs sharedvalidation.Error
	input.AuthAccountID = sharedvalidation.NormalizeTrimmed(input.AuthAccountID)
	input.Email = sharedvalidation.NormalizeEmail(input.Email)
	input.FullName = sharedvalidation.NormalizeText(input.FullName)
	input.Phone = v.normalizeOptionalPhone(&errs, "phone", input.Phone)

	sharedvalidation.ValidateID(&errs, "auth_account_id", input.AuthAccountID)
	sharedvalidation.ValidateEmail(&errs, "email", input.Email)
	sharedvalidation.ValidateRequiredSafeText(&errs, "full_name", input.FullName, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxShortNameLength)

	return input, errs
}

func (v UserValidator) NormalizeAndValidateUpdateProfile(input UpdateProfileInput, fields ProfileUpdateFields) (UpdateProfileInput, sharedvalidation.Error) {
	var errs sharedvalidation.Error
	if fields.FullName {
		input.FullName = sharedvalidation.NormalizeText(input.FullName)
		sharedvalidation.ValidateRequiredSafeText(&errs, "full_name", input.FullName, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxShortNameLength)
	}
	if fields.Phone {
		input.Phone = v.normalizeOptionalPhone(&errs, "phone", input.Phone)
	}
	if fields.AvatarURL {
		input.AvatarURL = sharedvalidation.NormalizeTrimmed(input.AvatarURL)
		sharedvalidation.ValidateOptionalHTTPSURL(&errs, "avatar_url", input.AvatarURL, sharedvalidation.MaxAvatarURLLength)
	}

	return input, errs
}

func (v UserValidator) NormalizeAndValidateAddress(input AddressInput) (AddressInput, sharedvalidation.Error) {
	var errs sharedvalidation.Error
	input.Name = sharedvalidation.NormalizeText(input.Name)
	input.Phone = v.normalizeOptionalPhone(&errs, "phone", input.Phone)
	input.Line1 = sharedvalidation.NormalizeText(input.Line1)
	input.Line2 = sharedvalidation.NormalizeText(input.Line2)
	input.City = sharedvalidation.NormalizeText(input.City)
	input.State = sharedvalidation.NormalizeText(input.State)
	input.PostalCode = sharedvalidation.NormalizeTrimmed(input.PostalCode)
	input.Country = sharedvalidation.NormalizeCountry(input.Country)

	sharedvalidation.ValidateRequiredSafeText(&errs, "name", input.Name, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxShortNameLength)
	sharedvalidation.ValidateRequiredSafeText(&errs, "line1", input.Line1, sharedvalidation.MinAddressLineLength, sharedvalidation.MaxAddressLineLength)
	sharedvalidation.ValidateRequiredSafeText(&errs, "city", input.City, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxCityStateLength)
	sharedvalidation.ValidateRequiredSafeText(&errs, "state", input.State, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxCityStateLength)
	sharedvalidation.ValidateOptionalSafeText(&errs, "line2", input.Line2, 0, sharedvalidation.MaxAddressLineLength)
	sharedvalidation.ValidateRequiredSafeText(&errs, "country", input.Country, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxCountryLength)
	sharedvalidation.ValidatePostalCode(&errs, "postal_code", input.PostalCode, input.Country)

	return input, errs
}

func (v UserValidator) NormalizeAndValidateUpdateSellerProfile(input UpdateSellerProfileInput, fields SellerUpdateFields) (UpdateSellerProfileInput, sharedvalidation.Error) {
	var errs sharedvalidation.Error
	if fields.StoreName {
		input.StoreName = sharedvalidation.NormalizeText(input.StoreName)
		sharedvalidation.ValidateRequiredSafeText(&errs, "store_name", input.StoreName, sharedvalidation.MinSellerNameLength, sharedvalidation.MaxShortNameLength)
	}
	if fields.DisplayName {
		input.DisplayName = sharedvalidation.NormalizeText(input.DisplayName)
		sharedvalidation.ValidateOptionalSafeText(&errs, "display_name", input.DisplayName, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxShortNameLength)
	}
	if fields.GSTNumber {
		input.GSTNumber = sharedvalidation.NormalizeUpper(input.GSTNumber)
		sharedvalidation.ValidateOptionalGSTIN(&errs, "gst_number", input.GSTNumber)
	}
	if fields.SupportEmail {
		input.SupportEmail = sharedvalidation.NormalizeEmail(input.SupportEmail)
		sharedvalidation.ValidateOptionalEmail(&errs, "support_email", input.SupportEmail)
	}

	return input, errs
}

func (v UserValidator) normalizeOptionalPhone(errs *sharedvalidation.Error, field string, value string) string {
	value = sharedvalidation.NormalizeTrimmed(value)
	if value == "" {
		return ""
	}
	normalized, ok := sharedvalidation.NormalizePhone(value, v.phoneRegion)
	if !ok {
		errs.Add(field, "phone", field+" must be a valid phone number")
		return value
	}
	return normalized
}
