package domain

import (
	"errors"
	"testing"
)

func TestSendOTPRequestValidate(t *testing.T) {
	t.Parallel()

	request := validOTPRequest()
	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	request.Channel = ChannelSMS
	request.Target = "+14155550100"
	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() SMS error = %v", err)
	}
}

func TestSendOTPRequestRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		change func(*SendOTPRequest)
	}{
		{name: "missing challenge", change: func(r *SendOTPRequest) { r.ChallengeID = "" }},
		{name: "push channel", change: func(r *SendOTPRequest) { r.Channel = ChannelPush }},
		{name: "missing target", change: func(r *SendOTPRequest) { r.Target = "" }},
		{name: "short OTP", change: func(r *SendOTPRequest) { r.OTP = "12345" }},
		{name: "nonnumeric OTP", change: func(r *SendOTPRequest) { r.OTP = "12345x" }},
		{name: "missing expiry", change: func(r *SendOTPRequest) { r.ExpiresInMinutes = 0 }},
		{name: "missing purpose", change: func(r *SendOTPRequest) { r.Purpose = "" }},
		{name: "unsupported purpose", change: func(r *SendOTPRequest) { r.Purpose = "export_secret" }},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := validOTPRequest()
			test.change(&request)
			if err := request.Validate(); !errors.Is(err, ErrInvalidOTPRequest) {
				t.Fatalf("Validate() error = %v, want %v", err, ErrInvalidOTPRequest)
			}
		})
	}
}

func validOTPRequest() SendOTPRequest {
	return SendOTPRequest{
		ChallengeID:      "challenge_123",
		Channel:          ChannelEmail,
		Target:           "person@example.com",
		OTP:              "482991",
		ExpiresInMinutes: 5,
		Purpose:          "login",
	}
}
