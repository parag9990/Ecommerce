package usecase

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
)

func TestPaymentSchemaDefinition(t *testing.T) {
	uc, err := NewPaymentSchemaUsecase(nil, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewPaymentSchemaUsecase returned error: %v", err)
	}

	definition, err := uc.Definition(context.Background())
	if err != nil {
		t.Fatalf("Definition returned error: %v", err)
	}
	if definition.DatabaseName != domain.PaymentDatabaseName {
		t.Fatalf("database = %q, want %q", definition.DatabaseName, domain.PaymentDatabaseName)
	}
	if len(definition.Tables) != 5 {
		t.Fatalf("tables len = %d, want 5", len(definition.Tables))
	}
}

func TestPaymentSchemaHealthUsesVerifier(t *testing.T) {
	verifier := &fakeSchemaVerifier{
		verification: domain.PaymentSchemaVerification{
			DatabaseName: domain.PaymentDatabaseName,
		},
	}
	uc, err := NewPaymentSchemaUsecase(verifier, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewPaymentSchemaUsecase returned error: %v", err)
	}

	out, err := uc.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}
	if !out.Verification.Ready {
		t.Fatalf("schema ready = false, want true: %+v", out.Verification)
	}
	if verifier.seen.DatabaseName != domain.PaymentDatabaseName {
		t.Fatalf("verifier saw database = %q, want %q", verifier.seen.DatabaseName, domain.PaymentDatabaseName)
	}
}

func TestPaymentSchemaHealthPropagatesVerifierError(t *testing.T) {
	wantErr := errors.New("db down")
	uc, err := NewPaymentSchemaUsecase(&fakeSchemaVerifier{err: wantErr}, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewPaymentSchemaUsecase returned error: %v", err)
	}

	_, err = uc.Health(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Health error = %v, want %v", err, wantErr)
	}
}

type fakeSchemaVerifier struct {
	seen         domain.PaymentSchemaDefinition
	verification domain.PaymentSchemaVerification
	err          error
}

func (f *fakeSchemaVerifier) VerifyPaymentSchema(ctx context.Context, definition domain.PaymentSchemaDefinition) (domain.PaymentSchemaVerification, error) {
	if err := ctx.Err(); err != nil {
		return domain.PaymentSchemaVerification{}, err
	}
	f.seen = definition
	if f.err != nil {
		return domain.PaymentSchemaVerification{}, f.err
	}
	return f.verification, nil
}
