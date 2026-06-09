package validation

import (
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/nyaruka/phonenumbers"
)

const (
	MaxIDLength          = 64
	MaxEmailLength       = 254
	MaxE164PhoneLength   = 16
	MaxAvatarURLLength   = 1024
	MaxAddressLineLength = 255
	MaxCityStateLength   = 128
	MaxPostalCodeLength  = 32
	MaxCountryLength     = 64
	MaxShortNameLength   = 120
	MinPersonNameLength  = 2
	MinAddressLineLength = 5
	MinSellerNameLength  = 3
)

var (
	idPattern          = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*$`)
	indiaPINPattern    = regexp.MustCompile(`^[1-9][0-9]{5}$`)
	usZIPPattern       = regexp.MustCompile(`^[0-9]{5}(-[0-9]{4})?$`)
	genericPostalCode  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 -]{1,31}$`)
	gstinPattern       = regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z][1-9A-Z]Z[0-9A-Z]$`)
	gstStateCodeLookup = map[string]struct{}{
		"01": {}, "02": {}, "03": {}, "04": {}, "05": {},
		"06": {}, "07": {}, "08": {}, "09": {}, "10": {},
		"11": {}, "12": {}, "13": {}, "14": {}, "15": {},
		"16": {}, "17": {}, "18": {}, "19": {}, "20": {},
		"21": {}, "22": {}, "23": {}, "24": {}, "25": {},
		"26": {}, "27": {}, "28": {}, "29": {}, "30": {},
		"31": {}, "32": {}, "33": {}, "34": {}, "35": {},
		"36": {}, "37": {}, "38": {}, "97": {},
	}
)

func IsValidID(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && utf8.RuneCountInString(value) <= MaxIDLength && idPattern.MatchString(value)
}

func ValidateID(errs *Error, field string, value string) {
	value = strings.TrimSpace(value)
	switch {
	case value == "":
		errs.AddRequired(field)
	case utf8.RuneCountInString(value) > MaxIDLength:
		errs.Add(field, "max", field+" must be at most 64 characters")
	case !idPattern.MatchString(value):
		errs.Add(field, "id", field+" contains unsupported characters")
	}
}

func IsValidEmail(email string) bool {
	email = NormalizeEmail(email)
	if email == "" || utf8.RuneCountInString(email) > MaxEmailLength {
		return false
	}
	if strings.ContainsAny(email, " \t\r\n") {
		return false
	}
	parsed, err := mail.ParseAddress(email)
	return err == nil && parsed.Address == email
}

func ValidateEmail(errs *Error, field string, email string) {
	email = NormalizeEmail(email)
	switch {
	case email == "":
		errs.AddRequired(field)
	case utf8.RuneCountInString(email) > MaxEmailLength:
		errs.Add(field, "max", field+" must be at most 254 characters")
	case !IsValidEmail(email):
		errs.Add(field, "email", field+" must be a valid email address")
	}
}

func ValidateOptionalEmail(errs *Error, field string, email string) {
	if strings.TrimSpace(email) == "" {
		return
	}
	ValidateEmail(errs, field, email)
}

func NormalizePhone(raw string, defaultRegion string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", false
	}

	region := strings.ToUpper(strings.TrimSpace(defaultRegion))
	if region == "" {
		region = DefaultPhoneRegion
	}

	number, err := phonenumbers.Parse(value, region)
	if err != nil || !phonenumbers.IsValidNumber(number) {
		return "", false
	}
	return phonenumbers.Format(number, phonenumbers.E164), true
}

func ValidateOptionalPhone(errs *Error, field string, phone string, defaultRegion string) {
	if strings.TrimSpace(phone) == "" {
		return
	}
	if _, ok := NormalizePhone(phone, defaultRegion); !ok {
		errs.Add(field, "phone", field+" must be a valid phone number")
	}
}

func ValidateRequiredSafeText(errs *Error, field string, value string, minLength int, maxLength int) {
	value = strings.TrimSpace(value)
	if value == "" {
		errs.AddRequired(field)
		return
	}
	ValidateOptionalSafeText(errs, field, value, minLength, maxLength)
}

func ValidateOptionalSafeText(errs *Error, field string, value string, minLength int, maxLength int) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	length := utf8.RuneCountInString(value)
	if minLength > 0 && length < minLength {
		errs.Add(field, "min", field+" is too short")
	}
	if maxLength > 0 && length > maxLength {
		errs.Add(field, "max", field+" is too long")
	}
	if !IsSafeText(value) {
		errs.Add(field, "safe_text", field+" contains unsupported characters")
	}
}

func IsSafeText(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) || r == '<' || r == '>' {
			return false
		}
	}
	return true
}

func ValidateOptionalHTTPSURL(errs *Error, field string, value string, maxLength int) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) > maxLength {
		errs.Add(field, "max", field+" is too long")
		return
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		errs.Add(field, "https_url", field+" must be a valid HTTPS URL")
	}
}

func IsValidPostalCode(country string, postalCode string) bool {
	country = strings.ToUpper(strings.TrimSpace(country))
	postalCode = strings.TrimSpace(postalCode)
	if postalCode == "" || utf8.RuneCountInString(postalCode) > MaxPostalCodeLength {
		return false
	}

	switch country {
	case "IN", "INDIA":
		return indiaPINPattern.MatchString(postalCode)
	case "US", "USA", "UNITED STATES":
		return usZIPPattern.MatchString(postalCode)
	default:
		return genericPostalCode.MatchString(postalCode)
	}
}

func ValidatePostalCode(errs *Error, field string, postalCode string, country string) {
	postalCode = strings.TrimSpace(postalCode)
	if postalCode == "" {
		errs.AddRequired(field)
		return
	}
	if !IsValidPostalCode(country, postalCode) {
		errs.Add(field, "postal_code", field+" must be valid for country")
	}
}

func IsValidGSTIN(raw string) bool {
	value := NormalizeUpper(raw)
	if len(value) != 15 || !gstinPattern.MatchString(value) {
		return false
	}
	_, ok := gstStateCodeLookup[value[:2]]
	return ok
}

func ValidateOptionalGSTIN(errs *Error, field string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	if !IsValidGSTIN(value) {
		errs.Add(field, "gstin", field+" must be a valid GSTIN")
	}
}
