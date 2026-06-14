package provider

import (
	"context"
	"fmt"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type PushClient interface {
	Deliver(
		ctx context.Context,
		deviceToken string,
		title string,
		textBody string,
		data map[string]string,
	) (string, error)
}

type PushProvider struct {
	name   string
	client PushClient
}

func NewPushProvider(name string, client PushClient) (*PushProvider, error) {
	name, err := validatedName(name)
	if err != nil {
		return nil, err
	}
	if missingDependency(client) {
		return nil, fmt.Errorf("%w: push client is required", ErrInvalidProvider)
	}
	return &PushProvider{name: name, client: client}, nil
}

func (p *PushProvider) Name() string {
	if p == nil {
		return ""
	}
	return p.name
}

func (p *PushProvider) Channel() domain.Channel {
	return domain.ChannelPush
}

func (p *PushProvider) Send(ctx context.Context, message domain.Message) (Result, error) {
	if p == nil || missingDependency(p.client) {
		return Result{}, ErrProviderUnavailable
	}
	if err := validateSend(ctx, p.Channel(), message); err != nil {
		return Result{Channel: p.Channel(), ProviderName: p.name, Status: StatusRejected}, err
	}

	messageID, err := p.client.Deliver(
		ctx,
		message.Recipient.DeviceToken,
		message.Content.Title,
		message.Content.TextBody,
		cloneData(message.Content.Data),
	)
	return deliveryResult(p.Channel(), p.name, messageID, err)
}

func cloneData(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
