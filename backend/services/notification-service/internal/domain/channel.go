package domain

import (
	"fmt"
	"strings"
)

// Channel identifies the delivery medium without exposing a vendor choice.
type Channel string

const (
	ChannelEmail        Channel = "email"
	ChannelSMS          Channel = "sms"
	ChannelPush         Channel = "push"
	ChannelWhatsAppLike Channel = "whatsapp_like"
)

type RecipientKind string

const (
	RecipientEmailAddress RecipientKind = "email_address"
	RecipientPhoneE164    RecipientKind = "phone_e164"
	RecipientDeviceToken  RecipientKind = "device_token"
)

type ChannelCapabilities struct {
	RequiresSubject bool
	SupportsHTML    bool
	SupportsData    bool
	RecipientKind   RecipientKind
}

func SupportedChannels() []Channel {
	return []Channel{ChannelEmail, ChannelSMS, ChannelPush, ChannelWhatsAppLike}
}

func ParseChannel(value string) (Channel, error) {
	channel := Channel(strings.TrimSpace(value))
	if !channel.IsSupported() {
		return "", fmt.Errorf("%w: %q", ErrUnsupportedChannel, value)
	}
	return channel, nil
}

func (c Channel) IsSupported() bool {
	switch c {
	case ChannelEmail, ChannelSMS, ChannelPush, ChannelWhatsAppLike:
		return true
	default:
		return false
	}
}

func (c Channel) Capabilities() (ChannelCapabilities, error) {
	switch c {
	case ChannelEmail:
		return ChannelCapabilities{
			RequiresSubject: true,
			SupportsHTML:    true,
			RecipientKind:   RecipientEmailAddress,
		}, nil
	case ChannelSMS:
		return ChannelCapabilities{RecipientKind: RecipientPhoneE164}, nil
	case ChannelPush:
		return ChannelCapabilities{
			SupportsData:  true,
			RecipientKind: RecipientDeviceToken,
		}, nil
	case ChannelWhatsAppLike:
		return ChannelCapabilities{RecipientKind: RecipientPhoneE164}, nil
	default:
		return ChannelCapabilities{}, fmt.Errorf("%w: %q", ErrUnsupportedChannel, c)
	}
}
