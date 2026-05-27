package provider

import "strings"

func NormalizeStripeLikeStatus(status string) (IntentStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "requires_action", "requires_confirmation":
		return IntentStatusRequiresAction, true
	case "requires_capture":
		return IntentStatusAuthorized, true
	case "succeeded":
		return IntentStatusCaptured, true
	case "canceled", "cancelled", "payment_failed":
		return IntentStatusFailed, true
	case "requires_payment_method", "processing":
		return IntentStatusInitiated, true
	default:
		return "", false
	}
}

func NormalizeRazorpayLikeStatus(status string) (IntentStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "created", "attempted":
		return IntentStatusInitiated, true
	case "authorized":
		return IntentStatusAuthorized, true
	case "paid", "captured":
		return IntentStatusCaptured, true
	case "failed":
		return IntentStatusFailed, true
	default:
		return "", false
	}
}

func ProviderEventTypeForIntentStatus(status IntentStatus) (WebhookEventType, bool) {
	switch status.Normalized() {
	case IntentStatusRequiresAction:
		return WebhookEventPaymentRequiresAction, true
	case IntentStatusAuthorized:
		return WebhookEventPaymentAuthorized, true
	case IntentStatusCaptured:
		return WebhookEventPaymentCaptured, true
	case IntentStatusFailed:
		return WebhookEventPaymentFailed, true
	default:
		return "", false
	}
}

func NormalizeStripeLikeRefundStatus(status string) (RefundStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "pending", "requires_action":
		return RefundStatusProcessing, true
	case "succeeded":
		return RefundStatusSucceeded, true
	case "failed", "canceled", "cancelled":
		return RefundStatusFailed, true
	default:
		return "", false
	}
}

func NormalizeRazorpayLikeRefundStatus(status string) (RefundStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "created", "pending", "processed":
		if strings.EqualFold(strings.TrimSpace(status), "processed") {
			return RefundStatusSucceeded, true
		}
		return RefundStatusProcessing, true
	case "failed":
		return RefundStatusFailed, true
	default:
		return "", false
	}
}
