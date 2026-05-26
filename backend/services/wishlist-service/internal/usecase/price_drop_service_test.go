package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

func TestPriceDropServicePublishesNotificationAndUpdatesSnapshot(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 31, 0, 0, time.UTC)
	eventTime := time.Date(2026, 5, 26, 10, 30, 0, 0, time.UTC)
	repository := &fakePriceDropRepository{
		candidates: []domain.PriceDropCandidate{
			{
				UserID:    " user_123 ",
				ProductID: "prod_123",
				VariantID: "var_1",
				PreviousPrice: domain.Money{
					Amount:   349900,
					Currency: "INR",
				},
			},
		},
		updateResult: domain.PriceUpdateResult{MatchedCount: 1, ModifiedCount: 1},
	}
	publisher := &fakeNotificationPublisher{}
	service := mustPriceDropService(t, repository, publisher, now)

	result, err := service.HandleProductPriceChanged(context.Background(), PriceChangeInput{
		EventID:   " evt_price_123 ",
		EventType: ProductPriceChangedEventTypeForTest,
		ProductID: " prod_123 ",
		VariantID: " var_1 ",
		NewPrice: domain.Money{
			Amount:   299900,
			Currency: "inr",
		},
		Title:      "Running Shoes",
		ImageURL:   " https://cdn.example.com/prod_123/main.jpg ",
		ProductURL: " /products/prod_123 ",
		OccurredAt: eventTime,
		TraceID:    " trace_123 ",
	})
	if err != nil {
		t.Fatalf("HandleProductPriceChanged returned error: %v", err)
	}
	if result.NotificationCount != 1 || result.CandidateCount != 1 {
		t.Fatalf("result counts = candidates %d notifications %d, want 1/1", result.CandidateCount, result.NotificationCount)
	}
	if len(publisher.commands) != 1 {
		t.Fatalf("published commands = %d, want 1", len(publisher.commands))
	}
	command := publisher.commands[0]
	if command.UserID != "user_123" {
		t.Fatalf("command.UserID = %q, want user_123", command.UserID)
	}
	if command.TemplateKey != DefaultPriceDropTemplateKey || command.Channel != DefaultNotificationChannel {
		t.Fatalf("command template/channel = %q/%q, want defaults", command.TemplateKey, command.Channel)
	}
	if command.IdempotencyKey != "wishlist_price_drop:user_123:prod_123:var_1:INR:299900" {
		t.Fatalf("idempotency key = %q, want stable key", command.IdempotencyKey)
	}
	if command.Variables["old_price_amount"] != int64(349900) || command.Variables["new_price_amount"] != int64(299900) {
		t.Fatalf("price variables = %#v", command.Variables)
	}
	if command.Variables["savings_amount"] != int64(50000) {
		t.Fatalf("savings_amount = %v, want 50000", command.Variables["savings_amount"])
	}
	if repository.updateProductID != "prod_123" || repository.updateVariantID != "var_1" {
		t.Fatalf("update target = %q/%q, want prod_123/var_1", repository.updateProductID, repository.updateVariantID)
	}
	if repository.updatePrice.Currency != "INR" || repository.updatePrice.Amount != 299900 {
		t.Fatalf("update price = %#v, want INR 299900", repository.updatePrice)
	}
	if !repository.updatedAt.Equal(eventTime) {
		t.Fatalf("updatedAt = %v, want event time %v", repository.updatedAt, eventTime)
	}
}

func TestPriceDropServiceSkipsNonDropsButUpdatesSnapshot(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 31, 0, 0, time.UTC)
	repository := &fakePriceDropRepository{
		candidates: []domain.PriceDropCandidate{
			{UserID: "user_1", ProductID: "prod_123", PreviousPrice: domain.Money{Amount: 299900, Currency: "INR"}},
			{UserID: "user_2", ProductID: "prod_123", PreviousPrice: domain.Money{Amount: 349900, Currency: "USD"}},
		},
		updateResult: domain.PriceUpdateResult{MatchedCount: 2, ModifiedCount: 2},
	}
	publisher := &fakeNotificationPublisher{}
	service := mustPriceDropService(t, repository, publisher, now)

	result, err := service.HandleProductPriceChanged(context.Background(), PriceChangeInput{
		EventID:   "evt_price_123",
		ProductID: "prod_123",
		NewPrice:  domain.Money{Amount: 299900, Currency: "INR"},
	})
	if err != nil {
		t.Fatalf("HandleProductPriceChanged returned error: %v", err)
	}
	if result.NotificationCount != 0 || len(publisher.commands) != 0 {
		t.Fatalf("notifications = result %d published %d, want 0", result.NotificationCount, len(publisher.commands))
	}
	if repository.updateCalls != 1 {
		t.Fatalf("update calls = %d, want 1", repository.updateCalls)
	}
	if !repository.updatedAt.Equal(now) {
		t.Fatalf("updatedAt = %v, want processing time %v", repository.updatedAt, now)
	}
}

func TestPriceDropServiceReturnsErrorAndDoesNotUpdateWhenPublishFails(t *testing.T) {
	publishErr := errors.New("kafka unavailable")
	repository := &fakePriceDropRepository{
		candidates: []domain.PriceDropCandidate{
			{UserID: "user_123", ProductID: "prod_123", PreviousPrice: domain.Money{Amount: 349900, Currency: "INR"}},
		},
	}
	service := mustPriceDropService(t, repository, &fakeNotificationPublisher{err: publishErr}, time.Now())

	_, err := service.HandleProductPriceChanged(context.Background(), PriceChangeInput{
		EventID:   "evt_price_123",
		ProductID: "prod_123",
		NewPrice:  domain.Money{Amount: 299900, Currency: "INR"},
	})
	if !errors.Is(err, publishErr) {
		t.Fatalf("HandleProductPriceChanged error = %v, want publish error", err)
	}
	if repository.updateCalls != 0 {
		t.Fatalf("update calls = %d, want 0 after publish failure", repository.updateCalls)
	}
}

func TestPriceDropServiceReturnsSnapshotUpdateErrorAfterPublish(t *testing.T) {
	updateErr := errors.New("mongo timeout")
	repository := &fakePriceDropRepository{
		candidates: []domain.PriceDropCandidate{
			{UserID: "user_123", ProductID: "prod_123", PreviousPrice: domain.Money{Amount: 349900, Currency: "INR"}},
		},
		updateErr: updateErr,
	}
	publisher := &fakeNotificationPublisher{}
	service := mustPriceDropService(t, repository, publisher, time.Now())

	_, err := service.HandleProductPriceChanged(context.Background(), PriceChangeInput{
		EventID:   "evt_price_123",
		ProductID: "prod_123",
		NewPrice:  domain.Money{Amount: 299900, Currency: "INR"},
	})
	if !errors.Is(err, updateErr) {
		t.Fatalf("HandleProductPriceChanged error = %v, want update error", err)
	}
	if len(publisher.commands) != 1 {
		t.Fatalf("published commands = %d, want 1 before update failure", len(publisher.commands))
	}
}

func TestPriceDropServiceValidatesInput(t *testing.T) {
	service := mustPriceDropService(t, &fakePriceDropRepository{}, &fakeNotificationPublisher{}, time.Now())

	_, err := service.HandleProductPriceChanged(context.Background(), PriceChangeInput{
		EventID:  "evt_price_123",
		NewPrice: domain.Money{Amount: 100, Currency: "INR"},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("missing product error = %v, want ErrInvalidInput", err)
	}

	_, err = service.HandleProductPriceChanged(context.Background(), PriceChangeInput{
		EventID:   "evt_price_123",
		ProductID: "prod_123",
		NewPrice:  domain.Money{Amount: 100, Currency: "US"},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("invalid currency error = %v, want ErrInvalidInput", err)
	}
}

func mustPriceDropService(t *testing.T, repository PriceDropRepository, publisher NotificationPublisher, now time.Time) *PriceDropService {
	t.Helper()
	service, err := NewPriceDropService(repository, publisher, nil, WithPriceDropClock(func() time.Time {
		return now
	}))
	if err != nil {
		t.Fatalf("NewPriceDropService returned error: %v", err)
	}
	return service
}

const ProductPriceChangedEventTypeForTest = "ProductPriceChanged"

type fakePriceDropRepository struct {
	candidates      []domain.PriceDropCandidate
	streamErr       error
	updateErr       error
	updateResult    domain.PriceUpdateResult
	updateCalls     int
	updateProductID string
	updateVariantID string
	updatePrice     domain.Money
	updatedAt       time.Time
}

func (f *fakePriceDropRepository) StreamPriceDropCandidates(ctx context.Context, productID string, variantID string, newPrice domain.Money, handle func(domain.PriceDropCandidate) error) error {
	if f.streamErr != nil {
		return f.streamErr
	}
	for _, candidate := range f.candidates {
		if err := handle(candidate); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakePriceDropRepository) UpdateLastKnownPriceForProduct(ctx context.Context, productID string, variantID string, newPrice domain.Money, updatedAt time.Time) (domain.PriceUpdateResult, error) {
	f.updateCalls++
	f.updateProductID = productID
	f.updateVariantID = variantID
	f.updatePrice = newPrice
	f.updatedAt = updatedAt
	if f.updateErr != nil {
		return domain.PriceUpdateResult{}, f.updateErr
	}
	return f.updateResult, nil
}

type fakeNotificationPublisher struct {
	commands []NotificationCommand
	err      error
}

func (f *fakeNotificationPublisher) Publish(ctx context.Context, command NotificationCommand) error {
	if f.err != nil {
		return f.err
	}
	f.commands = append(f.commands, command)
	return nil
}
