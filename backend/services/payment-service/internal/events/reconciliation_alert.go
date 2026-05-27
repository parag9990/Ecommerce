package events

import (
	"context"
	"errors"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
)

type reconciliationAlertEnvelope struct {
	Topic string                     `json:"topic"`
	Event domain.ReconciliationAlert `json:"event"`
}

func (p *HTTPPublisher) PublishMismatch(ctx context.Context, topic string, alert domain.ReconciliationAlert) error {
	if p == nil {
		return errors.New("payment reconciliation alert publisher is not initialized")
	}
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return errors.New("payment reconciliation alert topic is required")
	}
	alert = alert.Normalized()
	if err := alert.Validate(); err != nil {
		return err
	}
	return p.publishJSON(ctx, alert.EventID, reconciliationAlertEnvelope{Topic: topic, Event: alert})
}
