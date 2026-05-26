package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type WishlistCollectionRepository interface {
	Ping(ctx context.Context) error
	EnsureCollection(ctx context.Context) error
	CollectionName() string
}

type CollectionSetupService struct {
	repository WishlistCollectionRepository
	logger     *slog.Logger
}

func NewCollectionSetupService(repository WishlistCollectionRepository, logger *slog.Logger) (*CollectionSetupService, error) {
	if repository == nil {
		return nil, errors.New("wishlist collection repository is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &CollectionSetupService{
		repository: repository,
		logger:     logger,
	}, nil
}

func (s *CollectionSetupService) EnsureReady(ctx context.Context) error {
	if s == nil || s.repository == nil {
		return errors.New("wishlist collection setup service is not initialized")
	}
	if err := s.repository.Ping(ctx); err != nil {
		return fmt.Errorf("wishlist mongo readiness check failed: %w", err)
	}
	if err := s.repository.EnsureCollection(ctx); err != nil {
		return fmt.Errorf("wishlist collection setup failed: %w", err)
	}
	s.logger.Info("wishlist collection setup complete", "collection", s.repository.CollectionName())
	return nil
}
