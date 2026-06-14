package provider_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

func TestHTTPSMSClientDeliversThroughConfiguredGateway(t *testing.T) {
	t.Parallel()

	var authorization string
	gateway := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		authorization = request.Header.Get("Authorization")
		var input struct {
			To   string `json:"to"`
			Text string `json:"text"`
		}
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			t.Errorf("Decode() error = %v", err)
		}
		if input.To != "+14155550100" || input.Text != "Code 482991" {
			t.Errorf("SMS request = %+v", input)
		}
		_, _ = writer.Write([]byte(`{"message_id":"sms_123"}`))
	}))
	defer gateway.Close()

	client, err := provider.NewHTTPSMSClient(provider.HTTPSMSClientConfig{
		Endpoint: gateway.URL, BearerToken: "from-secret-manager", Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewHTTPSMSClient() error = %v", err)
	}
	messageID, err := client.Deliver(context.Background(), "+14155550100", "Code 482991")
	if err != nil || messageID != "sms_123" || authorization != "Bearer from-secret-manager" {
		t.Fatalf("Deliver() = %q, %v; Authorization = %q", messageID, err, authorization)
	}
}

func TestHTTPSMSClientReturnsOnlyClassifiedGatewayFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		want   error
	}{
		{name: "rejection", status: http.StatusBadRequest, want: provider.ErrProviderRejected},
		{name: "unavailable", status: http.StatusBadGateway, want: provider.ErrProviderUnavailable},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			gateway := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(test.status)
				_, _ = writer.Write([]byte("OTP 482991 rejected for +14155550100"))
			}))
			defer gateway.Close()
			client, err := provider.NewHTTPSMSClient(provider.HTTPSMSClientConfig{Endpoint: gateway.URL, Timeout: time.Second})
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.Deliver(context.Background(), "+14155550100", "OTP 482991")
			if !errors.Is(err, test.want) || strings.Contains(err.Error(), "482991") {
				t.Fatalf("Deliver() error = %v, want sanitized %v", err, test.want)
			}
		})
	}
}

func TestHTTPSMSClientRejectsInsecureRemoteEndpoint(t *testing.T) {
	t.Parallel()

	if _, err := provider.NewHTTPSMSClient(provider.HTTPSMSClientConfig{
		Endpoint: "http://sms.example.com/send", Timeout: time.Second,
	}); err == nil {
		t.Fatal("NewHTTPSMSClient() permitted an insecure remote endpoint")
	}
}

func TestSMTPClientValidatesConfigurationAndCancellation(t *testing.T) {
	t.Parallel()

	if _, err := provider.NewSMTPClient(provider.SMTPClientConfig{
		Host: "smtp.example.com", Port: 25, From: "sender@example.com",
		Username: "user", Password: "password", TLSMode: "none", Timeout: time.Second,
	}); err == nil {
		t.Fatal("NewSMTPClient() permitted credentials without TLS")
	}
	client, err := provider.NewSMTPClient(provider.SMTPClientConfig{
		Host: "localhost", Port: 1025, From: "sender@example.com", TLSMode: "none", Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewSMTPClient() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.Deliver(ctx, "person@example.com", "Code", "Code 482991", ""); !errors.Is(err, context.Canceled) {
		t.Fatalf("Deliver() error = %v, want context.Canceled", err)
	}
}
