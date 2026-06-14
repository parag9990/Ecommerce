package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type PreferenceReader interface {
	Get(ctx context.Context, userID string) (domain.Preference, error)
}

type ConsentGate struct {
	preferences PreferenceReader
}

func NewConsentGate(preferences PreferenceReader) (*ConsentGate, error) {
	if nilServiceDependency(preferences) {
		return nil, errors.New("notification preference reader is required")
	}
	return &ConsentGate{preferences: preferences}, nil
}

func (g *ConsentGate) Evaluate(
	ctx context.Context,
	userID string,
	channel domain.Channel,
	templateKey domain.TemplateKey,
	trustedSecurityFlow bool,
) (domain.ConsentDecision, error) {
	purpose, err := purposeForTemplate(templateKey)
	if err != nil {
		return domain.ConsentDecision{}, err
	}
	if purpose == domain.PurposeSecurity {
		if !trustedSecurityFlow {
			return domain.ConsentDecision{}, fmt.Errorf("%w: security template requires trusted send path",
				domain.ErrInvalidPreference)
		}
		return (domain.Preference{}).Decide(purpose, channel)
	}
	preference, err := g.preferences.Get(ctx, userID)
	if err != nil {
		return domain.ConsentDecision{}, fmt.Errorf("load notification preference for send: %w", err)
	}
	return preference.Decide(purpose, channel)
}

func purposeForTemplate(templateKey domain.TemplateKey) (domain.NotificationPurpose, error) {
	switch templateKey {
	case domain.TemplateOTPVerification:
		return domain.PurposeSecurity, nil
	case domain.TemplateOrderStatusUpdate, domain.TemplatePaymentStatusUpdate,
		domain.TemplateWelcomeUser, domain.TemplateSellerApproved, domain.TemplateAddressUpdated:
		return domain.PurposeTransactional, nil
	case domain.TemplatePromotionalOffer, domain.TemplateKey("price_drop"):
		return domain.PurposeMarketing, nil
	default:
		return "", fmt.Errorf("%w: template has no consent policy: %q",
			domain.ErrUnsupportedTemplateKey, templateKey)
	}
}
