package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

const (
	DefaultPriceDropTemplateKey = "wishlist_price_drop"
	DefaultNotificationChannel  = "auto"
	DefaultPriceDropMinDelta    = int64(1)
)

type PriceChangeInput struct {
	EventID    string
	EventType  string
	ProductID  string
	VariantID  string
	NewPrice   domain.Money
	Title      string
	ImageURL   string
	ProductURL string
	OccurredAt time.Time
	TraceID    string
}

type PriceDropResult struct {
	ProductID              string
	VariantID              string
	NewPrice               domain.Money
	CandidateCount         int64
	NotificationCount      int64
	SnapshotMatchedCount   int64
	SnapshotModifiedCount  int64
	PriceDropTemplateKey   string
	NotificationChannel    string
	SnapshotUpdateOccurred time.Time
}

type NotificationCommand struct {
	EventID        string
	UserID         string
	TemplateKey    string
	Channel        string
	IdempotencyKey string
	Variables      map[string]any
	TraceID        string
	OccurredAt     time.Time
}

type NotificationPublisher interface {
	Publish(ctx context.Context, command NotificationCommand) error
}

type PriceDropRepository interface {
	StreamPriceDropCandidates(
		ctx context.Context,
		productID string,
		variantID string,
		newPrice domain.Money,
		handle func(domain.PriceDropCandidate) error,
	) error
	UpdateLastKnownPriceForProduct(
		ctx context.Context,
		productID string,
		variantID string,
		newPrice domain.Money,
		updatedAt time.Time,
	) (domain.PriceUpdateResult, error)
}

type PriceDropConfig struct {
	TemplateKey    string
	Channel        string
	MinDeltaAmount int64
}

type PriceDropService struct {
	repository PriceDropRepository
	publisher  NotificationPublisher
	config     PriceDropConfig
	clock      func() time.Time
	logger     *slog.Logger
}

type PriceDropServiceOption func(*PriceDropService)

func WithPriceDropClock(clock func() time.Time) PriceDropServiceOption {
	return func(s *PriceDropService) {
		if clock != nil {
			s.clock = clock
		}
	}
}

func WithPriceDropConfig(config PriceDropConfig) PriceDropServiceOption {
	return func(s *PriceDropService) {
		s.config = normalizePriceDropConfig(config)
	}
}

func NewPriceDropService(repository PriceDropRepository, publisher NotificationPublisher, logger *slog.Logger, options ...PriceDropServiceOption) (*PriceDropService, error) {
	if repository == nil {
		return nil, errors.New("price drop repository is required")
	}
	if publisher == nil {
		return nil, errors.New("notification publisher is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	service := &PriceDropService{
		repository: repository,
		publisher:  publisher,
		config:     normalizePriceDropConfig(PriceDropConfig{}),
		clock: func() time.Time {
			return time.Now().UTC()
		},
		logger: logger,
	}
	for _, option := range options {
		option(service)
	}
	return service, nil
}

func (s *PriceDropService) HandleProductPriceChanged(ctx context.Context, input PriceChangeInput) (PriceDropResult, error) {
	if s == nil || s.repository == nil || s.publisher == nil {
		return PriceDropResult{}, errors.New("price drop service is not initialized")
	}
	ctx = contextOrBackground(ctx)

	normalized, err := normalizePriceChangeInput(input)
	if err != nil {
		return PriceDropResult{}, err
	}
	snapshotUpdatedAt := normalized.OccurredAt
	if snapshotUpdatedAt.IsZero() {
		snapshotUpdatedAt = s.clock().UTC()
		s.logger.Warn("product price event occurred_at missing; using processing time",
			"event_id", normalized.EventID,
			"event_type", normalized.EventType,
			"product_id", normalized.ProductID,
			"trace_id", normalized.TraceID,
		)
	}

	var candidateCount int64
	var notificationCount int64
	err = s.repository.StreamPriceDropCandidates(ctx, normalized.ProductID, normalized.VariantID, normalized.NewPrice, func(candidate domain.PriceDropCandidate) error {
		candidate = normalizePriceDropCandidate(candidate)
		if !isPriceDrop(candidate.PreviousPrice, normalized.NewPrice, s.config.MinDeltaAmount) {
			return nil
		}
		candidateCount++
		command := s.buildPriceDropNotification(normalized, candidate)
		if err := s.publisher.Publish(ctx, command); err != nil {
			return fmt.Errorf("publish price-drop notification command: %w", err)
		}
		notificationCount++
		return nil
	})
	if err != nil {
		s.logger.Error("wishlist price-drop notification publishing failed",
			"event_id", normalized.EventID,
			"product_id", normalized.ProductID,
			"variant_id", normalized.VariantID,
			"trace_id", normalized.TraceID,
			"error", err,
		)
		return PriceDropResult{}, err
	}

	updateResult, err := s.repository.UpdateLastKnownPriceForProduct(
		ctx,
		normalized.ProductID,
		normalized.VariantID,
		normalized.NewPrice,
		snapshotUpdatedAt,
	)
	if err != nil {
		s.logger.Error("wishlist price snapshot update failed",
			"event_id", normalized.EventID,
			"product_id", normalized.ProductID,
			"variant_id", normalized.VariantID,
			"trace_id", normalized.TraceID,
			"error", err,
		)
		return PriceDropResult{}, err
	}

	result := PriceDropResult{
		ProductID:              normalized.ProductID,
		VariantID:              normalized.VariantID,
		NewPrice:               normalized.NewPrice,
		CandidateCount:         candidateCount,
		NotificationCount:      notificationCount,
		SnapshotMatchedCount:   updateResult.MatchedCount,
		SnapshotModifiedCount:  updateResult.ModifiedCount,
		PriceDropTemplateKey:   s.config.TemplateKey,
		NotificationChannel:    s.config.Channel,
		SnapshotUpdateOccurred: snapshotUpdatedAt,
	}
	s.logger.Info("wishlist price drop processed",
		"event_id", normalized.EventID,
		"product_id", normalized.ProductID,
		"variant_id", normalized.VariantID,
		"currency", normalized.NewPrice.Currency,
		"new_amount", normalized.NewPrice.Amount,
		"candidate_count", candidateCount,
		"notification_count", notificationCount,
		"snapshot_matched_count", updateResult.MatchedCount,
		"snapshot_modified_count", updateResult.ModifiedCount,
		"trace_id", normalized.TraceID,
	)
	return result, nil
}

func (s *PriceDropService) buildPriceDropNotification(input PriceChangeInput, candidate domain.PriceDropCandidate) NotificationCommand {
	savingsAmount := candidate.PreviousPrice.Amount - input.NewPrice.Amount
	return NotificationCommand{
		EventID:        notificationEventID(input.EventID, candidate.UserID, candidate.ProductID, candidate.VariantID),
		UserID:         candidate.UserID,
		TemplateKey:    s.config.TemplateKey,
		Channel:        s.config.Channel,
		IdempotencyKey: priceDropIdempotencyKey(candidate, input.NewPrice),
		TraceID:        input.TraceID,
		OccurredAt:     s.clock().UTC(),
		Variables: map[string]any{
			"product_id":       input.ProductID,
			"variant_id":       input.VariantID,
			"title":            input.Title,
			"old_price_amount": candidate.PreviousPrice.Amount,
			"new_price_amount": input.NewPrice.Amount,
			"currency":         input.NewPrice.Currency,
			"savings_amount":   savingsAmount,
			"product_url":      input.ProductURL,
			"image_url":        input.ImageURL,
		},
	}
}

func normalizePriceDropConfig(config PriceDropConfig) PriceDropConfig {
	config.TemplateKey = normalizeID(config.TemplateKey)
	if config.TemplateKey == "" {
		config.TemplateKey = DefaultPriceDropTemplateKey
	}
	config.Channel = strings.ToLower(normalizeID(config.Channel))
	if config.Channel == "" {
		config.Channel = DefaultNotificationChannel
	}
	if config.MinDeltaAmount <= 0 {
		config.MinDeltaAmount = DefaultPriceDropMinDelta
	}
	return config
}

func normalizePriceChangeInput(input PriceChangeInput) (PriceChangeInput, error) {
	input.EventID = normalizeID(input.EventID)
	input.EventType = normalizeID(input.EventType)
	input.ProductID = normalizeID(input.ProductID)
	input.VariantID = normalizeID(input.VariantID)
	input.Title = strings.TrimSpace(input.Title)
	input.ImageURL = strings.TrimSpace(input.ImageURL)
	input.ProductURL = strings.TrimSpace(input.ProductURL)
	input.TraceID = normalizeID(input.TraceID)
	input.NewPrice.Currency = strings.ToUpper(strings.TrimSpace(input.NewPrice.Currency))
	input.OccurredAt = normalizeTimeOrZero(input.OccurredAt)

	if input.EventID == "" {
		return PriceChangeInput{}, ValidationError{Field: "event_id", Message: "is required"}
	}
	if input.ProductID == "" {
		return PriceChangeInput{}, ValidationError{Field: "product_id", Message: "is required"}
	}
	if input.NewPrice.Amount < 0 {
		return PriceChangeInput{}, ValidationError{Field: "new_price.amount", Message: "must be greater than or equal to zero"}
	}
	if !isCurrencyCode(input.NewPrice.Currency) {
		return PriceChangeInput{}, ValidationError{Field: "new_price.currency", Message: "must be a three-letter uppercase currency code"}
	}
	return input, nil
}

func normalizePriceDropCandidate(candidate domain.PriceDropCandidate) domain.PriceDropCandidate {
	candidate.UserID = normalizeID(candidate.UserID)
	candidate.ProductID = normalizeID(candidate.ProductID)
	candidate.VariantID = normalizeID(candidate.VariantID)
	candidate.PreviousPrice.Currency = strings.ToUpper(strings.TrimSpace(candidate.PreviousPrice.Currency))
	return candidate
}

func isPriceDrop(previous domain.Money, current domain.Money, minDeltaAmount int64) bool {
	if previous.Currency != current.Currency {
		return false
	}
	return previous.Amount-current.Amount >= minDeltaAmount
}

func priceDropIdempotencyKey(candidate domain.PriceDropCandidate, newPrice domain.Money) string {
	return strings.Join([]string{
		"wishlist_price_drop",
		normalizeID(candidate.UserID),
		normalizeID(candidate.ProductID),
		normalizeID(candidate.VariantID),
		strings.ToUpper(strings.TrimSpace(newPrice.Currency)),
		fmt.Sprintf("%d", newPrice.Amount),
	}, ":")
}

func notificationEventID(eventID, userID, productID, variantID string) string {
	parts := []string{"notif", normalizeID(eventID), normalizeID(userID), normalizeID(productID)}
	if variantID = normalizeID(variantID); variantID != "" {
		parts = append(parts, variantID)
	}
	return strings.Join(parts, "_")
}

func normalizeTimeOrZero(value time.Time) time.Time {
	if value.IsZero() {
		return value
	}
	return value.UTC()
}

func isCurrencyCode(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	for _, c := range currency {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}
