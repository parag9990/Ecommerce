package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

type OTPProviderRegistry interface {
	Resolve(channel domain.Channel) (provider.Provider, error)
}

type OTPDeliveryRepository interface {
	InsertDelivery(ctx context.Context, delivery domain.Delivery) error
}

type DeliveryAnalyticsRecorder interface {
	RecordSent(ctx context.Context, deliveryID, providerName, providerMessageID string, occurredAt time.Time) (domain.DeliveryEventApplyResult, error)
	RecordFailed(ctx context.Context, deliveryID, providerName, failureCode string, occurredAt time.Time) (domain.DeliveryEventApplyResult, error)
}

type SendOTPService struct {
	renderer   Renderer
	providers  OTPProviderRegistry
	deliveries OTPDeliveryRepository
	analytics  DeliveryAnalyticsRecorder
	logger     *slog.Logger
	now        func() time.Time
	newID      func() (string, error)
}

func NewSendOTPService(
	renderer Renderer,
	providers OTPProviderRegistry,
	deliveries OTPDeliveryRepository,
	analytics DeliveryAnalyticsRecorder,
	logger *slog.Logger,
) (*SendOTPService, error) {
	if nilDependency(renderer) {
		return nil, errors.New("OTP template renderer is required")
	}
	if nilDependency(providers) {
		return nil, errors.New("OTP provider registry is required")
	}
	if nilDependency(deliveries) {
		return nil, errors.New("OTP delivery repository is required")
	}
	if nilDependency(analytics) {
		return nil, errors.New("OTP delivery analytics recorder is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SendOTPService{
		renderer:   renderer,
		providers:  providers,
		deliveries: deliveries,
		analytics:  analytics,
		logger:     logger,
		now:        time.Now,
		newID:      newOTPDeliveryID,
	}, nil
}

func (s *SendOTPService) WithClock(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

func (s *SendOTPService) WithIDFactory(newID func() (string, error)) {
	if newID != nil {
		s.newID = newID
	}
}

func (s *SendOTPService) Send(ctx context.Context, req domain.SendOTPRequest) (domain.SendOTPResult, error) {
	if ctx == nil {
		return domain.SendOTPResult{}, fmt.Errorf("%w: context is required", domain.ErrInvalidOTPRequest)
	}
	if err := ctx.Err(); err != nil {
		return domain.SendOTPResult{}, err
	}
	if err := req.Validate(); err != nil {
		return domain.SendOTPResult{}, err
	}

	rendered, err := s.renderer.Render(ctx, domain.RenderRequest{
		TemplateKey: domain.TemplateOTPVerification,
		Channel:     req.Channel,
		Variables: map[string]string{
			"otp":                strings.TrimSpace(req.OTP),
			"expires_in_minutes": strconv.Itoa(req.ExpiresInMinutes),
		},
	})
	if err != nil {
		return domain.SendOTPResult{}, fmt.Errorf("render OTP template: %w", err)
	}

	message, err := otpMessage(req, rendered)
	if err != nil {
		return domain.SendOTPResult{}, fmt.Errorf("build OTP message: %w", err)
	}
	sender, err := s.providers.Resolve(req.Channel)
	if err != nil {
		return domain.SendOTPResult{}, fmt.Errorf("resolve OTP provider: %w", err)
	}
	deliveryID, err := s.newID()
	if err != nil {
		return domain.SendOTPResult{}, fmt.Errorf("generate OTP delivery id: %w", err)
	}

	result, sendErr := sender.Send(ctx, message)
	if errors.Is(sendErr, context.Canceled) || errors.Is(sendErr, context.DeadlineExceeded) {
		return domain.SendOTPResult{}, sendErr
	}
	status, safeSendErr := otpDeliveryOutcome(result, sendErr)
	record := newOTPDeliveryRecord(deliveryID, req, sender, result, status, s.now().UTC())
	if err := s.deliveries.InsertDelivery(ctx, record); err != nil {
		return domain.SendOTPResult{}, fmt.Errorf("record OTP delivery outcome: %w", err)
	}
	if err := s.recordAnalytics(ctx, record, safeSendErr); err != nil {
		return domain.SendOTPResult{}, fmt.Errorf("record OTP delivery analytics: %w", err)
	}

	s.logger.InfoContext(ctx, "notification.otp.delivery_recorded",
		slog.String("delivery_id", deliveryID),
		slog.String("challenge_id", strings.TrimSpace(req.ChallengeID)),
		slog.String("channel", string(req.Channel)),
		slog.String("status", string(status)),
		slog.String("correlation_id", strings.TrimSpace(req.CorrelationID)),
	)

	if safeSendErr != nil {
		return domain.SendOTPResult{}, fmt.Errorf("send OTP: %w", safeSendErr)
	}
	return domain.SendOTPResult{
		DeliveryID: deliveryID,
		Status:     string(status),
	}, nil
}

func (s *SendOTPService) recordAnalytics(ctx context.Context, record domain.Delivery, sendErr error) error {
	switch record.Status {
	case domain.DeliveryStatusAccepted:
		_, err := s.analytics.RecordSent(ctx, record.ID, record.Provider, record.ProviderMessageID, record.UpdatedAt)
		return err
	case domain.DeliveryStatusRejected:
		_, err := s.analytics.RecordFailed(ctx, record.ID, record.Provider, "provider_rejected_recipient", record.UpdatedAt)
		return err
	case domain.DeliveryStatusFailed:
		code := "provider_unavailable"
		if errors.Is(sendErr, provider.ErrProviderRejected) {
			code = "provider_rejected_recipient"
		}
		_, err := s.analytics.RecordFailed(ctx, record.ID, record.Provider, code, record.UpdatedAt)
		return err
	default:
		return nil
	}
}

func otpMessage(req domain.SendOTPRequest, rendered domain.RenderedMessage) (domain.Message, error) {
	message := domain.Message{
		Channel:       req.Channel,
		CorrelationID: strings.TrimSpace(req.CorrelationID),
		Content: domain.Content{
			Subject:  rendered.Subject,
			TextBody: rendered.Body,
		},
	}
	switch req.Channel {
	case domain.ChannelEmail:
		message.Recipient.Email = strings.TrimSpace(req.Target)
	case domain.ChannelSMS:
		message.Recipient.PhoneE164 = strings.TrimSpace(req.Target)
	default:
		return domain.Message{}, domain.ErrUnsupportedChannel
	}
	if err := message.Validate(); err != nil {
		return domain.Message{}, err
	}
	return message, nil
}

func otpDeliveryOutcome(result provider.Result, sendErr error) (domain.DeliveryStatus, error) {
	if sendErr == nil && result.Status == provider.StatusAccepted {
		return domain.DeliveryStatusAccepted, nil
	}
	if errors.Is(sendErr, provider.ErrProviderRejected) {
		return domain.DeliveryStatusRejected, provider.ErrProviderRejected
	}
	return domain.DeliveryStatusFailed, provider.ErrProviderUnavailable
}

func newOTPDeliveryRecord(
	deliveryID string,
	req domain.SendOTPRequest,
	sender provider.Provider,
	result provider.Result,
	status domain.DeliveryStatus,
	now time.Time,
) domain.Delivery {
	providerName := strings.TrimSpace(result.ProviderName)
	if providerName == "" {
		providerName = sender.Name()
	}
	payload := map[string]any{
		"challenge_id": strings.TrimSpace(req.ChallengeID),
		"purpose":      strings.TrimSpace(req.Purpose),
	}
	if correlationID := strings.TrimSpace(req.CorrelationID); correlationID != "" {
		payload["correlation_id"] = correlationID
	}
	return domain.Delivery{
		ID:                deliveryID,
		UserID:            strings.TrimSpace(req.UserID),
		Channel:           req.Channel,
		TemplateKey:       string(domain.TemplateOTPVerification),
		Status:            status,
		Provider:          providerName,
		ProviderMessageID: strings.TrimSpace(result.ProviderMessageID),
		Attempts:          1,
		Payload:           payload,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func newOTPDeliveryID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return "delivery_otp_" + hex.EncodeToString(bytes[:]), nil
}
