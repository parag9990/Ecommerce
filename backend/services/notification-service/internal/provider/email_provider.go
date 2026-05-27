package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type EmailClient interface {
	Deliver(ctx context.Context, to, subject, textBody, htmlBody string) (string, error)
}

type EmailProvider struct {
	name   string
	client EmailClient
}

func NewEmailProvider(name string, client EmailClient) (*EmailProvider, error) {
	name, err := validatedName(name)
	if err != nil {
		return nil, err
	}
	if missingDependency(client) {
		return nil, fmt.Errorf("%w: email client is required", ErrInvalidProvider)
	}
	return &EmailProvider{name: name, client: client}, nil
}

func (p *EmailProvider) Name() string {
	if p == nil {
		return ""
	}
	return p.name
}

func (p *EmailProvider) Channel() domain.Channel {
	return domain.ChannelEmail
}

func (p *EmailProvider) Send(ctx context.Context, message domain.Message) (Result, error) {
	if p == nil || missingDependency(p.client) {
		return Result{}, ErrProviderUnavailable
	}
	if err := validateSend(ctx, p.Channel(), message); err != nil {
		return Result{Channel: p.Channel(), ProviderName: p.name, Status: StatusRejected}, err
	}

	messageID, err := p.client.Deliver(
		ctx,
		strings.TrimSpace(message.Recipient.Email),
		message.Content.Subject,
		message.Content.TextBody,
		message.Content.HTMLBody,
	)
	return deliveryResult(p.Channel(), p.name, messageID, err)
}
