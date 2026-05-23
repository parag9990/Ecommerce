package auth

import "testing"

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{name: "empty", header: "", want: ""},
		{name: "valid", header: "Bearer token_123", want: "token_123"},
		{name: "case insensitive scheme", header: "bearer token_123", want: "token_123"},
		{name: "missing scheme", header: "token_123", want: ""},
		{name: "missing token", header: "Bearer", want: ""},
		{name: "wrong scheme", header: "Basic token_123", want: ""},
		{name: "too many fields", header: "Bearer token_123 extra", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BearerToken(tt.header); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
