package usecase

import (
	"fmt"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type templateSpec struct {
	requiredVariables []string
	allowedChannels   map[domain.Channel]struct{}
}

var templateSpecs = map[domain.TemplateKey]templateSpec{
	domain.TemplateOTPVerification: {
		requiredVariables: []string{"otp", "expires_in_minutes"},
		allowedChannels: channels(
			domain.ChannelEmail,
			domain.ChannelSMS,
		),
	},
	domain.TemplateOrderStatusUpdate: {
		requiredVariables: []string{"name", "order_id", "status"},
		allowedChannels: channels(
			domain.ChannelEmail,
			domain.ChannelSMS,
			domain.ChannelPush,
		),
	},
	domain.TemplatePaymentStatusUpdate: {
		requiredVariables: []string{"name", "order_id", "payment_status", "amount"},
		allowedChannels: channels(
			domain.ChannelEmail,
			domain.ChannelSMS,
		),
	},
	domain.TemplateWelcomeUser: {
		requiredVariables: []string{"name"},
		allowedChannels:   channels(domain.ChannelEmail),
	},
	domain.TemplateSellerApproved: {
		requiredVariables: []string{"name"},
		allowedChannels:   channels(domain.ChannelEmail),
	},
	domain.TemplateAddressUpdated: {
		requiredVariables: []string{"name"},
		allowedChannels:   channels(domain.ChannelEmail),
	},
	domain.TemplatePromotionalOffer: {
		requiredVariables: []string{"name", "offer_title", "coupon_code", "valid_until"},
		allowedChannels: channels(
			domain.ChannelEmail,
			domain.ChannelPush,
		),
	},
}

func approvedVariables(req domain.RenderRequest) (map[string]string, error) {
	spec, exists := templateSpecs[req.TemplateKey]
	if !exists {
		return nil, fmt.Errorf("%w: %w: %q",
			domain.ErrInvalidRenderRequest, domain.ErrUnsupportedTemplateKey, req.TemplateKey)
	}
	if !req.Channel.IsSupported() {
		return nil, fmt.Errorf("%w: %w: %q",
			domain.ErrInvalidRenderRequest, domain.ErrUnsupportedChannel, req.Channel)
	}
	if _, allowed := spec.allowedChannels[req.Channel]; !allowed {
		return nil, fmt.Errorf("%w: channel %q is not allowed for template %q",
			domain.ErrInvalidRenderRequest, req.Channel, req.TemplateKey)
	}

	approved := make(map[string]string, len(spec.requiredVariables))
	for _, name := range spec.requiredVariables {
		value, supplied := req.Variables[name]
		if !supplied || strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("%w: required variable %q is missing",
				domain.ErrInvalidRenderRequest, name)
		}
		approved[name] = value
	}
	return approved, nil
}

func channels(values ...domain.Channel) map[domain.Channel]struct{} {
	set := make(map[domain.Channel]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}
