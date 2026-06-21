package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/usecase"
)

type otpRenderer struct {
	result  domain.RenderedMessage
	err     error
	request domain.RenderRequest
	calls   int
}

func (r *otpRenderer) Render(_ context.Context, request domain.RenderRequest) (domain.RenderedMessage, error) {
	r.calls++
	r.request = request
	return r.result, r.err
}

type otpSender struct {
	channel domain.Channel
	result  provider.Result
	err     error
	message domain.Message
	calls   int
}

func (p *otpSender) Name() string            { return "configured_provider" }
func (p *otpSender) Channel() domain.Channel { return p.channel }
func (p *otpSender) Send(_ context.Context, message domain.Message) (provider.Result, error) {
	p.calls++
	p.message = message
	return p.result, p.err
}

type otpRegistry struct {
	sender provider.Provider
	calls  int
}

func (r *otpRegistry) Resolve(domain.Channel) (provider.Provider, error) {
	r.calls++
	return r.sender, nil
}

type otpDeliveries struct {
	record domain.Delivery
	err    error
	calls  int
}

type otpAnalytics struct {
	sentDelivery   string
	failedDelivery string
	failureCode    string
	err            error
}

func (a *otpAnalytics) RecordSent(_ context.Context, deliveryID, _, _ string, _ time.Time) (domain.DeliveryEventApplyResult, error) {
	a.sentDelivery = deliveryID
	return domain.DeliveryEventApplyResult{MilestoneChanged: true}, a.err
}

func (a *otpAnalytics) RecordFailed(_ context.Context, deliveryID, _ string, code string, _ time.Time) (domain.DeliveryEventApplyResult, error) {
	a.failedDelivery, a.failureCode = deliveryID, code
	return domain.DeliveryEventApplyResult{MilestoneChanged: true}, a.err
}

func (r *otpDeliveries) InsertDelivery(_ context.Context, record domain.Delivery) error {
	r.calls++
	r.record = record
	return r.err
}

func TestSendOTPSendsEmailAndPersistsNoSecret(t *testing.T) {
	t.Parallel()

	renderer := &otpRenderer{result: domain.RenderedMessage{
		Subject: "Your verification code",
		Body:    "Your verification code is 482991.",
	}}
	sender := &otpSender{channel: domain.ChannelEmail, result: provider.Result{
		Channel:           domain.ChannelEmail,
		ProviderName:      "mailpit",
		ProviderMessageID: "msg_123",
		Status:            provider.StatusAccepted,
	}}
	deliveries := &otpDeliveries{}
	service := newOTPService(t, renderer, sender, deliveries)

	got, err := service.Send(context.Background(), emailOTPRequest())
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if got.DeliveryID != "delivery_otp_test" || got.Status != "accepted" {
		t.Fatalf("Send() result = %+v", got)
	}
	if sender.calls != 1 || sender.message.Recipient.Email != "person@example.com" ||
		!strings.Contains(sender.message.Content.TextBody, "482991") {
		t.Fatalf("provider message = %+v", sender.message)
	}
	if renderer.request.TemplateKey != domain.TemplateOTPVerification ||
		renderer.request.Variables["otp"] != "482991" {
		t.Fatalf("renderer request = %+v", renderer.request)
	}

	persisted, err := json.Marshal(deliveries.record)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"482991", "person@example.com", "Your verification code is"} {
		if strings.Contains(string(persisted), forbidden) {
			t.Fatalf("persisted delivery contains sensitive OTP content: %s", persisted)
		}
	}
	if deliveries.record.TemplateKey != string(domain.TemplateOTPVerification) ||
		deliveries.record.UserID != "" || deliveries.record.Status != domain.DeliveryStatusAccepted {
		t.Fatalf("persisted delivery = %+v", deliveries.record)
	}
}

func TestSendOTPSendsSMS(t *testing.T) {
	t.Parallel()

	sender := &otpSender{channel: domain.ChannelSMS, result: provider.Result{
		Status: provider.StatusAccepted,
	}}
	service := newOTPService(t, &otpRenderer{result: domain.RenderedMessage{Body: "Code 482991"}}, sender, &otpDeliveries{})
	request := emailOTPRequest()
	request.Channel = domain.ChannelSMS
	request.Target = "+14155550100"

	if _, err := service.Send(context.Background(), request); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if sender.message.Recipient.PhoneE164 != "+14155550100" || sender.message.Content.Subject != "" {
		t.Fatalf("SMS message = %+v", sender.message)
	}
}

func TestSendOTPRejectsInvalidRequestBeforeRendering(t *testing.T) {
	t.Parallel()

	renderer := &otpRenderer{}
	sender := &otpSender{}
	service := newOTPService(t, renderer, sender, &otpDeliveries{})
	request := emailOTPRequest()
	request.Channel = domain.ChannelPush

	_, err := service.Send(context.Background(), request)
	if !errors.Is(err, domain.ErrInvalidOTPRequest) {
		t.Fatalf("Send() error = %v, want ErrInvalidOTPRequest", err)
	}
	if renderer.calls != 0 || sender.calls != 0 {
		t.Fatalf("invalid request rendered %d or sent %d times", renderer.calls, sender.calls)
	}
}

func TestSendOTPRejectsInvalidRecipientBeforeProvider(t *testing.T) {
	t.Parallel()

	renderer := &otpRenderer{result: domain.RenderedMessage{Subject: "Code", Body: "Code 482991"}}
	sender := &otpSender{}
	service := newOTPService(t, renderer, sender, &otpDeliveries{})
	request := emailOTPRequest()
	request.Target = "not an email"

	_, err := service.Send(context.Background(), request)
	if !errors.Is(err, domain.ErrInvalidRecipient) {
		t.Fatalf("Send() error = %v, want ErrInvalidRecipient", err)
	}
	if sender.calls != 0 {
		t.Fatalf("provider calls = %d, want 0", sender.calls)
	}
}

func TestSendOTPStopsWhenOTPTemplateCannotRender(t *testing.T) {
	t.Parallel()

	sender := &otpSender{}
	deliveries := &otpDeliveries{}
	service := newOTPService(t, &otpRenderer{err: domain.ErrTemplateNotFound}, sender, deliveries)

	_, err := service.Send(context.Background(), emailOTPRequest())
	if !errors.Is(err, domain.ErrTemplateNotFound) {
		t.Fatalf("Send() error = %v, want ErrTemplateNotFound", err)
	}
	if sender.calls != 0 || deliveries.calls != 0 {
		t.Fatalf("template failure sent %d or persisted %d times", sender.calls, deliveries.calls)
	}
}

func TestSendOTPRecordsSanitizedProviderFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		providerErr error
		wantError   error
		wantStatus  domain.DeliveryStatus
	}{
		{name: "rejected", providerErr: provider.ErrProviderRejected, wantError: provider.ErrProviderRejected, wantStatus: domain.DeliveryStatusRejected},
		{name: "raw unavailable failure", providerErr: errors.New("failed for person@example.com with OTP 482991"), wantError: provider.ErrProviderUnavailable, wantStatus: domain.DeliveryStatusFailed},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			deliveries := &otpDeliveries{}
			sender := &otpSender{channel: domain.ChannelEmail, err: test.providerErr}
			service := newOTPService(t,
				&otpRenderer{result: domain.RenderedMessage{Subject: "Code", Body: "Code 482991"}},
				sender,
				deliveries,
			)
			_, err := service.Send(context.Background(), emailOTPRequest())
			if !errors.Is(err, test.wantError) || strings.Contains(err.Error(), "person@example.com") ||
				strings.Contains(err.Error(), "482991") {
				t.Fatalf("Send() error = %v", err)
			}
			if deliveries.record.Status != test.wantStatus || deliveries.record.Provider != sender.Name() {
				t.Fatalf("delivery record = %+v", deliveries.record)
			}
		})
	}
}

func TestSendOTPDoesNotAcknowledgeDeliveryWithoutSafeTrace(t *testing.T) {
	t.Parallel()

	deliveries := &otpDeliveries{err: errors.New("insert failed")}
	sender := &otpSender{channel: domain.ChannelEmail, result: provider.Result{Status: provider.StatusAccepted}}
	service := newOTPService(t,
		&otpRenderer{result: domain.RenderedMessage{Subject: "Code", Body: "Code 482991"}},
		sender,
		deliveries,
	)
	result, err := service.Send(context.Background(), emailOTPRequest())
	if err == nil || result.Status != "" {
		t.Fatalf("Send() = %+v, %v; want no acknowledgement without persistence", result, err)
	}
}

func TestSendOTPAcknowledgesDurableDeliveryWhenAnalyticsIsUnavailable(t *testing.T) {
	t.Parallel()

	deliveries := &otpDeliveries{}
	sender := &otpSender{channel: domain.ChannelEmail, result: provider.Result{
		ProviderName: "mailpit", ProviderMessageID: "msg_123", Status: provider.StatusAccepted,
	}}
	analytics := &otpAnalytics{err: errors.New("transactions unavailable")}
	service := newOTPServiceWithAnalytics(t,
		&otpRenderer{result: domain.RenderedMessage{Subject: "Code", Body: "Code 482991"}},
		sender,
		deliveries,
		analytics,
	)

	result, err := service.Send(context.Background(), emailOTPRequest())
	if err != nil || result.Status != string(domain.DeliveryStatusAccepted) {
		t.Fatalf("Send() = %+v, %v; want durable accepted delivery", result, err)
	}
	if deliveries.calls != 1 || analytics.sentDelivery != "delivery_otp_test" {
		t.Fatalf("delivery calls = %d, analytics delivery = %q", deliveries.calls, analytics.sentDelivery)
	}
}

func TestSendOTPPropagatesCancellationWithoutNewPersistenceWork(t *testing.T) {
	t.Parallel()

	deliveries := &otpDeliveries{}
	sender := &otpSender{channel: domain.ChannelEmail, err: context.DeadlineExceeded}
	service := newOTPService(t,
		&otpRenderer{result: domain.RenderedMessage{Subject: "Code", Body: "Code 482991"}},
		sender,
		deliveries,
	)
	_, err := service.Send(context.Background(), emailOTPRequest())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Send() error = %v, want deadline exceeded", err)
	}
	if deliveries.calls != 0 {
		t.Fatalf("InsertDelivery() calls = %d, want 0", deliveries.calls)
	}
}

func newOTPService(t *testing.T, renderer usecase.Renderer, sender provider.Provider, deliveries *otpDeliveries) *usecase.SendOTPService {
	return newOTPServiceWithAnalytics(t, renderer, sender, deliveries, &otpAnalytics{})
}

func newOTPServiceWithAnalytics(t *testing.T, renderer usecase.Renderer, sender provider.Provider, deliveries *otpDeliveries, analytics usecase.DeliveryAnalyticsRecorder) *usecase.SendOTPService {
	t.Helper()
	registry := &otpRegistry{sender: sender}
	service, err := usecase.NewSendOTPService(renderer, registry, deliveries, analytics, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewSendOTPService() error = %v", err)
	}
	service.WithClock(func() time.Time {
		return time.Date(2026, time.May, 27, 0, 0, 0, 0, time.UTC)
	})
	service.WithIDFactory(func() (string, error) { return "delivery_otp_test", nil })
	return service
}

func emailOTPRequest() domain.SendOTPRequest {
	return domain.SendOTPRequest{
		ChallengeID:      "challenge_123",
		Channel:          domain.ChannelEmail,
		Target:           "person@example.com",
		OTP:              "482991",
		ExpiresInMinutes: 5,
		Purpose:          "login",
		CorrelationID:    "request_456",
	}
}
