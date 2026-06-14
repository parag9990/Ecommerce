package domain

import (
	"errors"
	"testing"
)

func TestMessageValidateAcceptsChannelContracts(t *testing.T) {
	t.Parallel()

	messages := []Message{
		{
			Channel:   ChannelEmail,
			Recipient: Recipient{Email: "buyer@example.com"},
			Content:   Content{Subject: "Order update", TextBody: "Packed", HTMLBody: "<p>Packed</p>"},
		},
		{
			Channel:   ChannelSMS,
			Recipient: Recipient{PhoneE164: "+919876543210"},
			Content:   Content{TextBody: "Your order is packed."},
		},
		{
			Channel:   ChannelPush,
			Recipient: Recipient{DeviceToken: "registration-token"},
			Content:   Content{Title: "Order update", TextBody: "Packed", Data: map[string]string{"order_id": "1"}},
		},
		{
			Channel:   ChannelWhatsAppLike,
			Recipient: Recipient{PhoneE164: "+14155550100"},
			Content:   Content{TextBody: "Your order is packed."},
		},
	}

	for _, message := range messages {
		message := message
		t.Run(string(message.Channel), func(t *testing.T) {
			t.Parallel()
			if err := message.Validate(); err != nil {
				t.Fatalf("Validate() returned error: %v", err)
			}
		})
	}
}

func TestMessageValidateRejectsInvalidContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message Message
		want    error
	}{
		{
			name:    "unsupported channel",
			message: Message{Channel: "fax"},
			want:    ErrUnsupportedChannel,
		},
		{
			name:    "invalid email recipient",
			message: Message{Channel: ChannelEmail, Recipient: Recipient{Email: "not-an-email"}, Content: Content{Subject: "s", TextBody: "b"}},
			want:    ErrInvalidRecipient,
		},
		{
			name:    "email requires subject",
			message: Message{Channel: ChannelEmail, Recipient: Recipient{Email: "buyer@example.com"}, Content: Content{TextBody: "b"}},
			want:    ErrInvalidContent,
		},
		{
			name:    "SMS requires e164 phone",
			message: Message{Channel: ChannelSMS, Recipient: Recipient{PhoneE164: "9876543210"}, Content: Content{TextBody: "b"}},
			want:    ErrInvalidRecipient,
		},
		{
			name:    "SMS rejects rich content",
			message: Message{Channel: ChannelSMS, Recipient: Recipient{PhoneE164: "+14155550100"}, Content: Content{TextBody: "b", HTMLBody: "<p>b</p>"}},
			want:    ErrInvalidContent,
		},
		{
			name:    "push requires device token",
			message: Message{Channel: ChannelPush, Content: Content{Title: "t", TextBody: "b"}},
			want:    ErrInvalidRecipient,
		},
		{
			name:    "whatsapp requires text",
			message: Message{Channel: ChannelWhatsAppLike, Recipient: Recipient{PhoneE164: "+14155550100"}},
			want:    ErrInvalidContent,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.message.Validate(); !errors.Is(err, test.want) {
				t.Fatalf("Validate() error = %v, want %v", err, test.want)
			}
		})
	}
}
