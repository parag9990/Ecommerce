package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

func TestGetSellerSettingsReturnsDefaultWhenMissing(t *testing.T) {
	repo := &memorySellerSettingsRepository{settings: map[string]domain.SellerSettings{}}
	uc, err := NewSellerSettingsUsecase(
		newTestAuthorizer(t, nil, nil),
		repo,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewSellerSettingsUsecase: %v", err)
	}

	result, err := uc.GetSellerSettings(context.Background(), GetSellerSettingsInput{
		Actor: activeActor(domain.RoleSeller),
	})
	if err != nil {
		t.Fatalf("GetSellerSettings: %v", err)
	}
	if result.SellerID != "seller_1" || result.ReturnPolicy != "" || result.ShippingPolicy != "" {
		t.Fatalf("unexpected default settings: %+v", result)
	}
	if result.Settings == nil {
		t.Fatalf("expected default settings map")
	}
}

func TestGetSellerSettingsAllowsInternalCaller(t *testing.T) {
	updatedAt := time.Unix(100, 0).UTC()
	repo := &memorySellerSettingsRepository{settings: map[string]domain.SellerSettings{
		"seller_2": {
			SellerID:       "seller_2",
			ReturnPolicy:   "7 day returns",
			ShippingPolicy: "Ships in two business days",
			UpdatedAt:      updatedAt,
		},
	}}
	uc, err := NewSellerSettingsUsecase(
		newTestAuthorizer(t, nil, nil),
		repo,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewSellerSettingsUsecase: %v", err)
	}

	result, err := uc.GetSellerSettings(context.Background(), GetSellerSettingsInput{
		SellerID:       "seller_2",
		InternalCaller: "api-gateway",
	})
	if err != nil {
		t.Fatalf("GetSellerSettings: %v", err)
	}
	if result.SellerID != "seller_2" || result.ReturnPolicy != "7 day returns" || !result.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("unexpected settings: %+v", result)
	}
}

type memorySellerSettingsRepository struct {
	settings map[string]domain.SellerSettings
}

func (r *memorySellerSettingsRepository) GetSellerSettings(ctx context.Context, sellerID string) (domain.SellerSettings, error) {
	settings, ok := r.settings[sellerID]
	if !ok {
		return domain.SellerSettings{}, domain.ErrSellerSettingsNotFound
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return domain.SellerSettings{}, ctx.Err()
	}
	return settings, nil
}
