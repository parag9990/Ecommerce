package domain

import (
	"errors"
	"testing"
	"time"
)

func TestTemplateValidate(t *testing.T) {
	t.Parallel()

	if err := validTemplate().Validate(); err != nil {
		t.Fatalf("Validate() returned error for valid template: %v", err)
	}
}

func TestTemplateValidateRejectsInvalidDocuments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		change func(*Template)
	}{
		{name: "missing id", change: func(template *Template) { template.ID = "" }},
		{name: "unknown channel", change: func(template *Template) { template.Channel = "fax" }},
		{name: "email missing subject", change: func(template *Template) { template.Subject = "" }},
		{name: "invalid status", change: func(template *Template) { template.Status = "deleted" }},
		{name: "invalid version", change: func(template *Template) { template.Version = 0 }},
		{name: "timestamps reversed", change: func(template *Template) { template.UpdatedAt = template.CreatedAt.Add(-time.Second) }},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			template := validTemplate()
			test.change(&template)
			if err := template.Validate(); !errors.Is(err, ErrInvalidTemplate) {
				t.Fatalf("Validate() error = %v, want %v", err, ErrInvalidTemplate)
			}
		})
	}
}

func validTemplate() Template {
	now := time.Date(2026, time.May, 18, 0, 0, 0, 0, time.UTC)
	return Template{
		ID:          "tpl_order_paid_email",
		TemplateKey: "order_paid",
		Channel:     ChannelEmail,
		Subject:     "Order confirmed",
		Body:        "Your order is confirmed.",
		Status:      TemplateStatusActive,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
