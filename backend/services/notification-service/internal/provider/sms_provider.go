package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type SMSClient interface {
	Deliver(ctx context.Context, phoneE164, textBody string) (string, error)
}

type SMSProvider struct {
	name   string
	client SMSClient
}

func NewSMSProvider(name string, client SMSClient) (*SMSProvider, error) {
	name, err := validatedName(name)
	if err != nil {
		return nil, err
	}
	if missingDependency(client) {
		return nil, fmt.Errorf("%w: SMS client is required", ErrInvalidProvider)
	}
	return &SMSProvider{name: name, client: client}, nil
}

func (p *SMSProvider) Name() string {
	if p == nil {
		return ""
	}
	return p.name
}

func (p *SMSProvider) Channel() domain.Channel {
	return domain.ChannelSMS
}

func (p *SMSProvider) Send(ctx context.Context, message domain.Message) (Result, error) {
	if p == nil || missingDependency(p.client) {
		return Result{}, ErrProviderUnavailable
	}
	if err := validateSend(ctx, p.Channel(), message); err != nil {
		return Result{Channel: p.Channel(), ProviderName: p.name, Status: StatusRejected}, err
	}

	messageID, err := p.client.Deliver(
		ctx,
		strings.TrimSpace(message.Recipient.PhoneE164),
		message.Content.TextBody,
	)
	return deliveryResult(p.Channel(), p.name, messageID, err)
}
