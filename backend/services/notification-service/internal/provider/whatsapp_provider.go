package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type WhatsAppLikeClient interface {
	Deliver(ctx context.Context, phoneE164, textBody string) (string, error)
}

type WhatsAppLikeProvider struct {
	name   string
	client WhatsAppLikeClient
}

func NewWhatsAppLikeProvider(name string, client WhatsAppLikeClient) (*WhatsAppLikeProvider, error) {
	name, err := validatedName(name)
	if err != nil {
		return nil, err
	}
	if missingDependency(client) {
		return nil, fmt.Errorf("%w: WhatsApp-like client is required", ErrInvalidProvider)
	}
	return &WhatsAppLikeProvider{name: name, client: client}, nil
}

func (p *WhatsAppLikeProvider) Name() string {
	if p == nil {
		return ""
	}
	return p.name
}

func (p *WhatsAppLikeProvider) Channel() domain.Channel {
	return domain.ChannelWhatsAppLike
}

func (p *WhatsAppLikeProvider) Send(ctx context.Context, message domain.Message) (Result, error) {
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
