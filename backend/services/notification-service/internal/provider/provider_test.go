package provider_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

type emailClient struct {
	messageID string
	err       error
}

func (c emailClient) Deliver(context.Context, string, string, string, string) (string, error) {
	return c.messageID, c.err
}

type smsClient struct{}

func (smsClient) Deliver(context.Context, string, string) (string, error) {
	return "sms-1", nil
}

type pushClient struct{}

func (pushClient) Deliver(context.Context, string, string, string, map[string]string) (string, error) {
	return "push-1", nil
}

type whatsappClient struct{}

func (whatsappClient) Deliver(context.Context, string, string) (string, error) {
	return "wa-1", nil
}

func TestChannelProvidersSendValidMessage(t *testing.T) {
	t.Parallel()

	email, err := provider.NewEmailProvider("email", emailClient{messageID: "email-1"})
	if err != nil {
		t.Fatal(err)
	}
	sms, err := provider.NewSMSProvider("sms", smsClient{})
	if err != nil {
		t.Fatal(err)
	}
	push, err := provider.NewPushProvider("push", pushClient{})
	if err != nil {
		t.Fatal(err)
	}
	whatsapp, err := provider.NewWhatsAppLikeProvider("wa", whatsappClient{})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		sender   provider.Provider
		message  domain.Message
		expected string
	}{
		{
			name:     "email",
			sender:   email,
			message:  domain.Message{Channel: domain.ChannelEmail, Recipient: domain.Recipient{Email: "buyer@example.com"}, Content: domain.Content{Subject: "Update", TextBody: "Packed"}},
			expected: "email-1",
		},
		{
			name:     "sms",
			sender:   sms,
			message:  domain.Message{Channel: domain.ChannelSMS, Recipient: domain.Recipient{PhoneE164: "+14155550100"}, Content: domain.Content{TextBody: "Packed"}},
			expected: "sms-1",
		},
		{
			name:     "push",
			sender:   push,
			message:  domain.Message{Channel: domain.ChannelPush, Recipient: domain.Recipient{DeviceToken: "token"}, Content: domain.Content{Title: "Update", TextBody: "Packed"}},
			expected: "push-1",
		},
		{
			name:     "whatsapp_like",
			sender:   whatsapp,
			message:  domain.Message{Channel: domain.ChannelWhatsAppLike, Recipient: domain.Recipient{PhoneE164: "+14155550100"}, Content: domain.Content{TextBody: "Packed"}},
			expected: "wa-1",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := test.sender.Send(context.Background(), test.message)
			if err != nil {
				t.Fatalf("Send() returned error: %v", err)
			}
			if got.Status != provider.StatusAccepted || got.ProviderMessageID != test.expected {
				t.Fatalf("Send() = %+v, want accepted result with id %q", got, test.expected)
			}
		})
	}
}

func TestProviderRejectsMessageForWrongChannel(t *testing.T) {
	t.Parallel()

	email, err := provider.NewEmailProvider("email", emailClient{messageID: "email-1"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = email.Send(context.Background(), domain.Message{Channel: domain.ChannelSMS})
	if !errors.Is(err, provider.ErrProviderRejected) {
		t.Fatalf("Send() error = %v, want ErrProviderRejected", err)
	}
}

func TestProviderConvertsClientFailureWithoutLeakingDetails(t *testing.T) {
	t.Parallel()

	email, err := provider.NewEmailProvider("email", emailClient{err: errors.New("recipient buyer@example.com refused")})
	if err != nil {
		t.Fatal(err)
	}
	_, err = email.Send(context.Background(), domain.Message{
		Channel:   domain.ChannelEmail,
		Recipient: domain.Recipient{Email: "buyer@example.com"},
		Content:   domain.Content{Subject: "Update", TextBody: "Packed"},
	})
	if !errors.Is(err, provider.ErrProviderUnavailable) {
		t.Fatalf("Send() error = %v, want ErrProviderUnavailable", err)
	}
	if err.Error() != provider.ErrProviderUnavailable.Error() {
		t.Fatalf("Send() error exposed downstream detail: %v", err)
	}
}

func TestProviderSanitizesClassifiedRejection(t *testing.T) {
	t.Parallel()

	email, err := provider.NewEmailProvider("email", emailClient{err: errors.Join(provider.ErrProviderRejected, errors.New("recipient buyer@example.com refused"))})
	if err != nil {
		t.Fatal(err)
	}
	result, err := email.Send(context.Background(), domain.Message{
		Channel:   domain.ChannelEmail,
		Recipient: domain.Recipient{Email: "buyer@example.com"},
		Content:   domain.Content{Subject: "Update", TextBody: "Packed"},
	})
	if !errors.Is(err, provider.ErrProviderRejected) || result.Status != provider.StatusRejected {
		t.Fatalf("Send() = %+v, %v; want rejected result", result, err)
	}
	if err.Error() != provider.ErrProviderRejected.Error() {
		t.Fatalf("Send() error exposed downstream detail: %v", err)
	}
}

func TestProviderPropagatesContextCancellation(t *testing.T) {
	t.Parallel()

	email, err := provider.NewEmailProvider("email", emailClient{messageID: "email-1"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = email.Send(ctx, domain.Message{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Send() error = %v, want context.Canceled", err)
	}
}
