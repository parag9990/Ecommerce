package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

var sensitivePaymentJSONKeys = []string{
	"api_key",
	"api_secret",
	"authorization",
	"card_number",
	"client_secret",
	"cvv",
	"cvc",
	"secret_key",
	"token",
	"webhook_secret",
}

func validateSanitizedJSONPayload(raw []byte, required bool, field string) error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		if required {
			return fmt.Errorf("%w: %s is required", ErrInvalidPayment, field)
		}
		return nil
	}
	if !json.Valid(raw) {
		return fmt.Errorf("%w: %s must be valid JSON", ErrInvalidPayment, field)
	}
	lower := strings.ToLower(string(raw))
	for _, key := range sensitivePaymentJSONKeys {
		if strings.Contains(lower, `"`+key+`"`) {
			return fmt.Errorf("%w: %s contains sensitive key %q", ErrInvalidPayment, field, key)
		}
	}
	return nil
}
