package usecase

import (
	"context"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
)

type PaymentSchemaVerifier interface {
	VerifyPaymentSchema(ctx context.Context, definition domain.PaymentSchemaDefinition) (domain.PaymentSchemaVerification, error)
}

type PaymentSchemaUsecase struct {
	verifier PaymentSchemaVerifier
	logger   *slog.Logger
}

type PaymentSchemaHealthOutput struct {
	Definition   domain.PaymentSchemaDefinition
	Verification domain.PaymentSchemaVerification
}

func NewPaymentSchemaUsecase(verifier PaymentSchemaVerifier, logger *slog.Logger) (*PaymentSchemaUsecase, error) {
	if logger == nil {
		logger = slog.Default()
	}
	return &PaymentSchemaUsecase{
		verifier: verifier,
		logger:   logger,
	}, nil
}

func (u *PaymentSchemaUsecase) Definition(ctx context.Context) (domain.PaymentSchemaDefinition, error) {
	if err := ctx.Err(); err != nil {
		return domain.PaymentSchemaDefinition{}, err
	}
	return domain.PaymentSchema(), nil
}

func (u *PaymentSchemaUsecase) Health(ctx context.Context) (PaymentSchemaHealthOutput, error) {
	if err := ctx.Err(); err != nil {
		return PaymentSchemaHealthOutput{}, err
	}

	definition := domain.PaymentSchema()
	if u.verifier == nil {
		verification := domain.PaymentSchemaVerification{
			DatabaseName:  definition.DatabaseName,
			MissingTables: tableNames(definition.Tables),
		}.WithReadyFlag()
		return PaymentSchemaHealthOutput{Definition: definition, Verification: verification}, nil
	}

	verification, err := u.verifier.VerifyPaymentSchema(ctx, definition)
	if err != nil {
		u.logger.ErrorContext(ctx, "payment.schema.verify_failed", slog.String("error", err.Error()))
		return PaymentSchemaHealthOutput{}, err
	}
	return PaymentSchemaHealthOutput{
		Definition:   definition,
		Verification: verification.WithReadyFlag(),
	}, nil
}

func tableNames(tables []domain.PaymentTableDefinition) []string {
	names := make([]string, 0, len(tables))
	for _, table := range tables {
		names = append(names, table.Name)
	}
	return names
}
