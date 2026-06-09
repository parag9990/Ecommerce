package handlers

import sharedvalidation "github.com/parag/ecommerce/backend/shared/validation"

func (h *UserHandler) normalizeAndValidateUpdateProfile(req *updateUserProfileRequest) sharedvalidation.Error {
	var errs sharedvalidation.Error
	if req.FullName == nil && req.Phone == nil && req.AvatarURL == nil {
		errs.Add("request", "empty_patch", "at least one profile field is required")
		return errs
	}

	if req.FullName != nil {
		value := sharedvalidation.NormalizeText(*req.FullName)
		req.FullName = &value
		sharedvalidation.ValidateRequiredSafeText(&errs, "full_name", value, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxShortNameLength)
	}
	if req.Phone != nil {
		value := sharedvalidation.NormalizeTrimmed(*req.Phone)
		if value != "" {
			normalized, ok := sharedvalidation.NormalizePhone(value, h.validationPhoneRegion)
			if !ok {
				errs.Add("phone", "phone", "phone must be a valid phone number")
			} else {
				value = normalized
			}
		}
		req.Phone = &value
	}
	if req.AvatarURL != nil {
		value := sharedvalidation.NormalizeTrimmed(*req.AvatarURL)
		req.AvatarURL = &value
		sharedvalidation.ValidateOptionalHTTPSURL(&errs, "avatar_url", value, sharedvalidation.MaxAvatarURLLength)
	}

	return errs
}

func (h *UserHandler) normalizeAndValidateAddress(req *addressInputRequest) sharedvalidation.Error {
	var errs sharedvalidation.Error

	req.Name = sharedvalidation.NormalizeText(req.Name)
	req.Line1 = sharedvalidation.NormalizeText(req.Line1)
	req.City = sharedvalidation.NormalizeText(req.City)
	req.State = sharedvalidation.NormalizeText(req.State)
	req.PostalCode = sharedvalidation.NormalizeTrimmed(req.PostalCode)
	req.Country = sharedvalidation.NormalizeCountry(req.Country)
	if req.Phone != nil {
		value := sharedvalidation.NormalizeTrimmed(*req.Phone)
		if value != "" {
			normalized, ok := sharedvalidation.NormalizePhone(value, h.validationPhoneRegion)
			if !ok {
				errs.Add("phone", "phone", "phone must be a valid phone number")
			} else {
				value = normalized
			}
		}
		req.Phone = &value
	}
	if req.Line2 != nil {
		value := sharedvalidation.NormalizeText(*req.Line2)
		req.Line2 = &value
		sharedvalidation.ValidateOptionalSafeText(&errs, "line2", value, 0, sharedvalidation.MaxAddressLineLength)
	}

	sharedvalidation.ValidateRequiredSafeText(&errs, "name", req.Name, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxShortNameLength)
	sharedvalidation.ValidateRequiredSafeText(&errs, "line1", req.Line1, sharedvalidation.MinAddressLineLength, sharedvalidation.MaxAddressLineLength)
	sharedvalidation.ValidateRequiredSafeText(&errs, "city", req.City, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxCityStateLength)
	sharedvalidation.ValidateRequiredSafeText(&errs, "state", req.State, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxCityStateLength)
	sharedvalidation.ValidateRequiredSafeText(&errs, "country", req.Country, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxCountryLength)
	sharedvalidation.ValidatePostalCode(&errs, "postal_code", req.PostalCode, req.Country)

	return errs
}

func (h *UserHandler) normalizeAndValidateSellerProfile(req *updateSellerProfileRequest) sharedvalidation.Error {
	var errs sharedvalidation.Error
	if req.StoreName == nil && req.DisplayName == nil && req.GSTNumber == nil && req.SupportEmail == nil {
		errs.Add("request", "empty_patch", "at least one seller profile field is required")
		return errs
	}

	if req.StoreName != nil {
		value := sharedvalidation.NormalizeText(*req.StoreName)
		req.StoreName = &value
		sharedvalidation.ValidateRequiredSafeText(&errs, "store_name", value, sharedvalidation.MinSellerNameLength, sharedvalidation.MaxShortNameLength)
	}
	if req.DisplayName != nil {
		value := sharedvalidation.NormalizeText(*req.DisplayName)
		req.DisplayName = &value
		sharedvalidation.ValidateOptionalSafeText(&errs, "display_name", value, sharedvalidation.MinPersonNameLength, sharedvalidation.MaxShortNameLength)
	}
	if req.GSTNumber != nil {
		value := sharedvalidation.NormalizeUpper(*req.GSTNumber)
		req.GSTNumber = &value
		sharedvalidation.ValidateOptionalGSTIN(&errs, "gst_number", value)
	}
	if req.SupportEmail != nil {
		value := sharedvalidation.NormalizeEmail(*req.SupportEmail)
		req.SupportEmail = &value
		sharedvalidation.ValidateOptionalEmail(&errs, "support_email", value)
	}

	return errs
}

func normalizeAndValidatePathID(field string, value string) (string, sharedvalidation.Error) {
	normalized := sharedvalidation.NormalizeTrimmed(value)
	var errs sharedvalidation.Error
	sharedvalidation.ValidateID(&errs, field, normalized)
	return normalized, errs
}
