package domain

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
)

var e164Pattern = regexp.MustCompile(`^\+[1-9][0-9]{1,14}$`)

type Recipient struct {
	Email       string
	PhoneE164   string
	DeviceToken string
}

type Content struct {
	Subject  string
	TextBody string
	HTMLBody string
	Title    string
	Data     map[string]string
}

type Message struct {
	Channel       Channel
	Recipient     Recipient
	Content       Content
	CorrelationID string
}

func (m Message) Validate() error {
	if !m.Channel.IsSupported() {
		return fmt.Errorf("%w: %q", ErrUnsupportedChannel, m.Channel)
	}

	switch m.Channel {
	case ChannelEmail:
		if !validEmail(m.Recipient.Email) {
			return fmt.Errorf("%w: email address is required and must be valid", ErrInvalidRecipient)
		}
		if blank(m.Content.Subject) || blank(m.Content.TextBody) {
			return fmt.Errorf("%w: email subject and text body are required", ErrInvalidContent)
		}
		if !blank(m.Content.Title) || len(m.Content.Data) != 0 {
			return fmt.Errorf("%w: email does not accept push title or data", ErrInvalidContent)
		}
	case ChannelSMS:
		if !validPhoneE164(m.Recipient.PhoneE164) {
			return fmt.Errorf("%w: SMS recipient must be in E.164 format", ErrInvalidRecipient)
		}
		if blank(m.Content.TextBody) {
			return fmt.Errorf("%w: SMS text body is required", ErrInvalidContent)
		}
		if hasRichContent(m.Content) {
			return fmt.Errorf("%w: SMS accepts text body only", ErrInvalidContent)
		}
	case ChannelPush:
		if blank(m.Recipient.DeviceToken) {
			return fmt.Errorf("%w: push device token is required", ErrInvalidRecipient)
		}
		if blank(m.Content.Title) || blank(m.Content.TextBody) {
			return fmt.Errorf("%w: push title and text body are required", ErrInvalidContent)
		}
		if !blank(m.Content.Subject) || !blank(m.Content.HTMLBody) {
			return fmt.Errorf("%w: push does not accept email content", ErrInvalidContent)
		}
	case ChannelWhatsAppLike:
		if !validPhoneE164(m.Recipient.PhoneE164) {
			return fmt.Errorf("%w: WhatsApp-like recipient must be in E.164 format", ErrInvalidRecipient)
		}
		if blank(m.Content.TextBody) {
			return fmt.Errorf("%w: WhatsApp-like text body is required", ErrInvalidContent)
		}
		if hasRichContent(m.Content) {
			return fmt.Errorf("%w: WhatsApp-like channel accepts text body only", ErrInvalidContent)
		}
	}

	return nil
}

func validEmail(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return false
	}
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value && strings.Contains(value, "@")
}

func validPhoneE164(value string) bool {
	return e164Pattern.MatchString(strings.TrimSpace(value))
}

func blank(value string) bool {
	return strings.TrimSpace(value) == ""
}

func hasRichContent(content Content) bool {
	return !blank(content.Subject) ||
		!blank(content.HTMLBody) ||
		!blank(content.Title) ||
		len(content.Data) != 0
}
