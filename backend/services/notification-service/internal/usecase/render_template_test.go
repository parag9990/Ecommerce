package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/usecase"
)

type templateRepository struct {
	template domain.Template
	err      error
	calls    int
	key      string
	channel  domain.Channel
}

func (r *templateRepository) FindLatestActive(
	_ context.Context,
	templateKey string,
	channel domain.Channel,
) (domain.Template, error) {
	r.calls++
	r.key = templateKey
	r.channel = channel
	return r.template, r.err
}

func (*templateRepository) InsertTemplate(context.Context, domain.Template) error {
	return nil
}

func TestTemplateRendererRendersRequiredTemplateFamilies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		request     domain.RenderRequest
		subject     string
		body        string
		wantSubject string
		wantBody    string
	}{
		{
			name: "OTP SMS",
			request: domain.RenderRequest{
				TemplateKey: domain.TemplateOTPVerification,
				Channel:     domain.ChannelSMS,
				Variables:   map[string]string{"otp": "482991", "expires_in_minutes": "5"},
			},
			body:     "Your code is {{.otp}}. Valid for {{.expires_in_minutes}} minutes.",
			wantBody: "Your code is 482991. Valid for 5 minutes.",
		},
		{
			name: "order email",
			request: domain.RenderRequest{
				TemplateKey: domain.TemplateOrderStatusUpdate,
				Channel:     domain.ChannelEmail,
				Variables:   map[string]string{"name": "Riya", "order_id": "ORD-1042", "status": "Shipped"},
			},
			subject:     "Order {{.order_id}} is {{.status}}",
			body:        "Hi {{.name}}, your order {{.order_id}} is now {{.status}}.",
			wantSubject: "Order ORD-1042 is Shipped",
			wantBody:    "Hi Riya, your order ORD-1042 is now Shipped.",
		},
		{
			name: "payment email",
			request: domain.RenderRequest{
				TemplateKey: domain.TemplatePaymentStatusUpdate,
				Channel:     domain.ChannelEmail,
				Variables: map[string]string{
					"name": "Riya", "order_id": "ORD-1042", "payment_status": "successful", "amount": "INR 1299",
				},
			},
			subject:     "Payment {{.payment_status}} for order {{.order_id}}",
			body:        "Hi {{.name}}, payment of {{.amount}} for order {{.order_id}} is {{.payment_status}}.",
			wantSubject: "Payment successful for order ORD-1042",
			wantBody:    "Hi Riya, payment of INR 1299 for order ORD-1042 is successful.",
		},
		{
			name: "welcome user email",
			request: domain.RenderRequest{
				TemplateKey: domain.TemplateWelcomeUser,
				Channel:     domain.ChannelEmail,
				Variables:   map[string]string{"name": "Riya"},
			},
			subject:     "Welcome to Ecommerce",
			body:        "Hi {{.name}}, welcome to Ecommerce. Your account is ready.",
			wantSubject: "Welcome to Ecommerce",
			wantBody:    "Hi Riya, welcome to Ecommerce. Your account is ready.",
		},
		{
			name: "promotional push",
			request: domain.RenderRequest{
				TemplateKey: domain.TemplatePromotionalOffer,
				Channel:     domain.ChannelPush,
				Variables: map[string]string{
					"name": "Riya", "offer_title": "Weekend sale", "coupon_code": "SAVE20", "valid_until": "31 May",
				},
			},
			body:     "Hi {{.name}}, {{.offer_title}}! Use {{.coupon_code}} before {{.valid_until}}.",
			wantBody: "Hi Riya, Weekend sale! Use SAVE20 before 31 May.",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := &templateRepository{template: storedTemplate(test.request, test.subject, test.body)}
			renderer := newRenderer(t, repo)
			got, err := renderer.Render(context.Background(), test.request)
			if err != nil {
				t.Fatalf("Render() returned error: %v", err)
			}
			if got.Subject != test.wantSubject || got.Body != test.wantBody {
				t.Fatalf("Render() = subject %q body %q, want subject %q body %q",
					got.Subject, got.Body, test.wantSubject, test.wantBody)
			}
			if got.TemplateID != repo.template.ID || got.TemplateVersion != 1 ||
				got.TemplateKey != test.request.TemplateKey || got.Channel != test.request.Channel {
				t.Fatalf("Render() metadata = %+v", got)
			}
			if repo.calls != 1 || repo.key != string(test.request.TemplateKey) || repo.channel != test.request.Channel {
				t.Fatalf("FindLatestActive() called %d times with key %q channel %q",
					repo.calls, repo.key, repo.channel)
			}
		})
	}
}

func TestTemplateRendererRejectsInvalidRequestsBeforeLookup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		request domain.RenderRequest
		want    error
	}{
		{
			name: "missing OTP",
			request: domain.RenderRequest{
				TemplateKey: domain.TemplateOTPVerification,
				Channel:     domain.ChannelSMS,
				Variables:   map[string]string{"expires_in_minutes": "5"},
			},
			want: domain.ErrInvalidRenderRequest,
		},
		{
			name: "disallowed channel",
			request: domain.RenderRequest{
				TemplateKey: domain.TemplatePromotionalOffer,
				Channel:     domain.ChannelSMS,
				Variables: map[string]string{
					"name": "Riya", "offer_title": "Sale", "coupon_code": "SAVE20", "valid_until": "31 May",
				},
			},
			want: domain.ErrInvalidRenderRequest,
		},
		{
			name: "unknown key",
			request: domain.RenderRequest{
				TemplateKey: "unknown",
				Channel:     domain.ChannelEmail,
			},
			want: domain.ErrUnsupportedTemplateKey,
		},
		{
			name: "unknown channel",
			request: domain.RenderRequest{
				TemplateKey: domain.TemplateOTPVerification,
				Channel:     "fax",
				Variables:   map[string]string{"otp": "482991", "expires_in_minutes": "5"},
			},
			want: domain.ErrUnsupportedChannel,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := &templateRepository{}
			renderer := newRenderer(t, repo)
			_, err := renderer.Render(context.Background(), test.request)
			if !errors.Is(err, test.want) {
				t.Fatalf("Render() error = %v, want %v", err, test.want)
			}
			if repo.calls != 0 {
				t.Fatalf("FindLatestActive() calls = %d, want 0", repo.calls)
			}
		})
	}
}

func TestTemplateRendererFiltersAdditionalSensitiveVariables(t *testing.T) {
	t.Parallel()

	request := domain.RenderRequest{
		TemplateKey: domain.TemplateOTPVerification,
		Channel:     domain.ChannelSMS,
		Variables: map[string]string{
			"otp": "482991", "expires_in_minutes": "5", "access_token": "do-not-render",
		},
	}
	repo := &templateRepository{template: storedTemplate(
		request, "", "Code {{.otp}} token {{.access_token}}",
	)}
	renderer := newRenderer(t, repo)

	_, err := renderer.Render(context.Background(), request)
	if !errors.Is(err, domain.ErrTemplateRender) {
		t.Fatalf("Render() error = %v, want ErrTemplateRender", err)
	}
	if strings.Contains(err.Error(), "do-not-render") {
		t.Fatalf("Render() exposed filtered variable value: %v", err)
	}
}

func TestTemplateRendererRejectsInvalidStoredTemplateContent(t *testing.T) {
	t.Parallel()

	request := domain.RenderRequest{
		TemplateKey: domain.TemplateOrderStatusUpdate,
		Channel:     domain.ChannelEmail,
		Variables:   map[string]string{"name": "Riya", "order_id": "ORD-1042", "status": "Shipped"},
	}
	tests := []struct {
		name   string
		change func(*domain.Template, *domain.RenderRequest)
	}{
		{
			name: "malformed source",
			change: func(stored *domain.Template, _ *domain.RenderRequest) {
				stored.Body = "Hi {{.name"
			},
		},
		{
			name: "empty email subject",
			change: func(stored *domain.Template, _ *domain.RenderRequest) {
				stored.Subject = ""
			},
		},
		{
			name: "inactive repository result",
			change: func(stored *domain.Template, _ *domain.RenderRequest) {
				stored.Status = domain.TemplateStatusInactive
			},
		},
		{
			name: "unsafe rendered email subject",
			change: func(_ *domain.Template, req *domain.RenderRequest) {
				req.Variables["status"] = "Packed\r\nBcc: hidden@example.com"
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			testRequest := request
			testRequest.Variables = cloneVariables(request.Variables)
			stored := storedTemplate(testRequest,
				"Order {{.order_id}} is {{.status}}",
				"Hi {{.name}}, order {{.order_id}} is {{.status}}.",
			)
			test.change(&stored, &testRequest)
			repo := &templateRepository{template: stored}
			renderer := newRenderer(t, repo)
			if _, err := renderer.Render(context.Background(), testRequest); !errors.Is(err, domain.ErrTemplateRender) {
				t.Fatalf("Render() error = %v, want ErrTemplateRender", err)
			}
		})
	}
}

func TestTemplateRendererPropagatesContextAndRepositoryErrors(t *testing.T) {
	t.Parallel()

	request := domain.RenderRequest{
		TemplateKey: domain.TemplateOTPVerification,
		Channel:     domain.ChannelSMS,
		Variables:   map[string]string{"otp": "482991", "expires_in_minutes": "5"},
	}

	t.Run("canceled context avoids lookup", func(t *testing.T) {
		t.Parallel()

		repo := &templateRepository{}
		renderer := newRenderer(t, repo)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := renderer.Render(ctx, request); !errors.Is(err, context.Canceled) {
			t.Fatalf("Render() error = %v, want context.Canceled", err)
		}
		if repo.calls != 0 {
			t.Fatalf("FindLatestActive() calls = %d, want 0", repo.calls)
		}
	})

	t.Run("repository not found remains classifiable", func(t *testing.T) {
		t.Parallel()

		repo := &templateRepository{err: domain.ErrTemplateNotFound}
		renderer := newRenderer(t, repo)
		if _, err := renderer.Render(context.Background(), request); !errors.Is(err, domain.ErrTemplateNotFound) {
			t.Fatalf("Render() error = %v, want ErrTemplateNotFound", err)
		}
	})
}

func TestNewTemplateRendererRejectsMissingRepository(t *testing.T) {
	t.Parallel()

	if _, err := usecase.NewTemplateRenderer(nil); err == nil {
		t.Fatal("NewTemplateRenderer() returned nil error")
	}
}

func newRenderer(t *testing.T, repository *templateRepository) *usecase.TemplateRenderer {
	t.Helper()
	renderer, err := usecase.NewTemplateRenderer(repository)
	if err != nil {
		t.Fatalf("NewTemplateRenderer() returned error: %v", err)
	}
	return renderer
}

func storedTemplate(req domain.RenderRequest, subject, body string) domain.Template {
	now := time.Date(2026, time.May, 27, 0, 0, 0, 0, time.UTC)
	return domain.Template{
		ID:          "tpl_" + string(req.TemplateKey) + "_" + string(req.Channel) + "_v1",
		TemplateKey: string(req.TemplateKey),
		Channel:     req.Channel,
		Subject:     subject,
		Body:        body,
		Status:      domain.TemplateStatusActive,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func cloneVariables(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
