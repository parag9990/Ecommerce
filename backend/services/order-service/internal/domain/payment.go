package domain

import "fmt"

type PaymentResult string

const (
	PaymentResultCaptured PaymentResult = "captured"
	PaymentResultFailed   PaymentResult = "failed"
)

type PaymentIntentStatus string

const (
	PaymentIntentStatusInitiated      PaymentIntentStatus = "initiated"
	PaymentIntentStatusRequiresAction PaymentIntentStatus = "requires_action"
)

func ParsePaymentResult(value string) (PaymentResult, error) {
	result := PaymentResult(value)
	switch result {
	case PaymentResultCaptured, PaymentResultFailed:
		return result, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedPaymentResult, value)
	}
}

func IsPayableIntentStatus(value string) bool {
	switch PaymentIntentStatus(value) {
	case PaymentIntentStatusInitiated, PaymentIntentStatusRequiresAction:
		return true
	default:
		return false
	}
}
