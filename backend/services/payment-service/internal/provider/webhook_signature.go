package provider

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func validHMACSHA256(secret string, signedPayload []byte, signatureHex string) bool {
	expected, err := hex.DecodeString(strings.TrimSpace(signatureHex))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(signedPayload)
	return hmac.Equal(mac.Sum(nil), expected)
}

func invalidWebhookSignature(providerName string) error {
	return NewError(
		providerName,
		"verify_webhook",
		ErrorCodeWebhookSignature,
		"payment webhook signature is invalid",
		WithStatusCode(http.StatusBadRequest),
		WithCause(ErrInvalidWebhookSignature),
	)
}

func invalidWebhookPayload(providerName string, message string, cause error) error {
	if cause != nil {
		cause = errors.Join(ErrInvalidProviderRequest, cause)
	} else {
		cause = ErrInvalidProviderRequest
	}
	return NewError(
		providerName,
		"verify_webhook",
		ErrorCodeValidation,
		message,
		WithStatusCode(http.StatusBadRequest),
		WithCause(cause),
	)
}

func stripeSignatureParts(raw string) (int64, []string, error) {
	var timestamp int64
	var signatures []string
	for _, part := range strings.Split(raw, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			parsed, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return 0, nil, err
			}
			timestamp = parsed
		case "v1":
			signatures = append(signatures, value)
		}
	}
	if timestamp <= 0 || len(signatures) == 0 {
		return 0, nil, fmt.Errorf("signature header lacks timestamp or v1 signature")
	}
	return timestamp, signatures, nil
}

func validWebhookTimestamp(timestamp int64, tolerance time.Duration) bool {
	eventTime := time.Unix(timestamp, 0)
	delta := time.Since(eventTime)
	if delta < 0 {
		delta = -delta
	}
	return delta <= tolerance
}
