package domain

import (
	"errors"
	"testing"
)

func TestParseChannel(t *testing.T) {
	t.Parallel()

	for _, channel := range SupportedChannels() {
		channel := channel
		t.Run(string(channel), func(t *testing.T) {
			t.Parallel()
			got, err := ParseChannel(string(channel))
			if err != nil {
				t.Fatalf("ParseChannel(%q) returned error: %v", channel, err)
			}
			if got != channel {
				t.Fatalf("ParseChannel(%q) = %q, want %q", channel, got, channel)
			}
		})
	}

	if _, err := ParseChannel("fax"); !errors.Is(err, ErrUnsupportedChannel) {
		t.Fatalf("ParseChannel(\"fax\") error = %v, want ErrUnsupportedChannel", err)
	}
}

func TestChannelCapabilities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		channel         Channel
		recipient       RecipientKind
		requiresSubject bool
		supportsHTML    bool
		supportsData    bool
	}{
		{channel: ChannelEmail, recipient: RecipientEmailAddress, requiresSubject: true, supportsHTML: true},
		{channel: ChannelSMS, recipient: RecipientPhoneE164},
		{channel: ChannelPush, recipient: RecipientDeviceToken, supportsData: true},
		{channel: ChannelWhatsAppLike, recipient: RecipientPhoneE164},
	}

	for _, test := range tests {
		test := test
		t.Run(string(test.channel), func(t *testing.T) {
			t.Parallel()
			got, err := test.channel.Capabilities()
			if err != nil {
				t.Fatalf("Capabilities() returned error: %v", err)
			}
			if got.RecipientKind != test.recipient ||
				got.RequiresSubject != test.requiresSubject ||
				got.SupportsHTML != test.supportsHTML ||
				got.SupportsData != test.supportsData {
				t.Fatalf("Capabilities() = %+v, want recipient=%q subject=%v html=%v data=%v",
					got, test.recipient, test.requiresSubject, test.supportsHTML, test.supportsData)
			}
		})
	}
}
