package events

import (
	"fmt"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

func BuildTrigger(envelope Envelope) (domain.EventNotificationTrigger, bool, error) {
	var trigger domain.EventNotificationTrigger
	switch envelope.EventType {
	case "OrderCreated", "OrderPaid", "OrderCancelled", "OrderDelivered":
		if err := requireProducer(envelope, "order-service"); err != nil {
			return trigger, false, err
		}
		var payload OrderPayload
		if err := decodePayload(envelope.Payload, &payload); err != nil {
			return trigger, false, err
		}
		status := map[string]string{
			"OrderCreated":   "Created",
			"OrderPaid":      "Paid",
			"OrderCancelled": "Cancelled",
			"OrderDelivered": "Delivered",
		}[envelope.EventType]
		trigger = newEmailTrigger(envelope, payload.UserID, payload.Email, domain.TemplateOrderStatusUpdate,
			map[string]string{"name": payload.Name, "order_id": payload.OrderID, "status": status})

	case "PaymentSucceeded", "PaymentFailed":
		if err := requireProducer(envelope, "payment-service"); err != nil {
			return trigger, false, err
		}
		var payload PaymentPayload
		if err := decodePayload(envelope.Payload, &payload); err != nil {
			return trigger, false, err
		}
		status := "Failed"
		if envelope.EventType == "PaymentSucceeded" {
			status = "Successful"
		}
		trigger = newEmailTrigger(envelope, payload.UserID, payload.Email, domain.TemplatePaymentStatusUpdate,
			map[string]string{
				"name": payload.Name, "order_id": payload.OrderID, "payment_status": status, "amount": payload.Amount,
			})

	case "UserCreated", "SellerApproved", "AddressUpdated":
		if err := requireProducer(envelope, "user-service"); err != nil {
			return trigger, false, err
		}
		var payload UserPayload
		if err := decodePayload(envelope.Payload, &payload); err != nil {
			return trigger, false, err
		}
		templateKey := map[string]domain.TemplateKey{
			"UserCreated":    domain.TemplateWelcomeUser,
			"SellerApproved": domain.TemplateSellerApproved,
			"AddressUpdated": domain.TemplateAddressUpdated,
		}[envelope.EventType]
		trigger = newEmailTrigger(envelope, payload.UserID, payload.Email, templateKey,
			map[string]string{"name": payload.Name})

	default:
		return domain.EventNotificationTrigger{}, false, nil
	}
	if err := trigger.Validate(); err != nil {
		return domain.EventNotificationTrigger{}, false, fmt.Errorf("%w: event notification fields are invalid: %w",
			ErrInvalidEvent, err)
	}
	return trigger, true, nil
}

func requireProducer(envelope Envelope, expected string) error {
	if envelope.Producer != expected {
		return fmt.Errorf("%w: event type %q must be produced by %q", ErrInvalidEvent, envelope.EventType, expected)
	}
	return nil
}

func newEmailTrigger(
	envelope Envelope,
	userID string,
	email string,
	templateKey domain.TemplateKey,
	variables map[string]string,
) domain.EventNotificationTrigger {
	channel := domain.ChannelEmail
	return domain.EventNotificationTrigger{
		IdempotencyKey: envelope.EventID + ":" + string(templateKey) + ":" + string(channel),
		SourceEventID:  envelope.EventID,
		SourceType:     envelope.EventType,
		TraceID:        envelope.TraceID,
		UserID:         userID,
		Channel:        channel,
		Recipient:      email,
		TemplateKey:    templateKey,
		Variables:      variables,
	}
}
