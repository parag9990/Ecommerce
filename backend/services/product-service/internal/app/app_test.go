package app

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"product-service/internal/config"
	"product-service/internal/usecase"
)

type expiryUseCaseStub struct {
	called chan usecase.ExpireInventoryReservationsRequest
}

func (s *expiryUseCaseStub) ReserveInventory(context.Context, usecase.ReserveInventoryRequest) (*usecase.InventoryReservationResult, error) {
	return nil, nil
}

func (s *expiryUseCaseStub) ReleaseInventory(context.Context, usecase.InventoryReservationActionRequest) error {
	return nil
}

func (s *expiryUseCaseStub) CommitInventory(context.Context, usecase.InventoryReservationActionRequest) error {
	return nil
}

func (s *expiryUseCaseStub) ExpireReservations(_ context.Context, request usecase.ExpireInventoryReservationsRequest) (usecase.ExpireInventoryReservationsResult, error) {
	s.called <- request
	return usecase.ExpireInventoryReservationsResult{}, nil
}

func TestStartBackgroundWorkersRunsInventoryExpiryImmediately(t *testing.T) {
	cfg := config.Default()
	cfg.Inventory.ExpiryBatchLimit = 17
	cfg.Inventory.ExpiryIntervalSeconds = 60
	stub := &expiryUseCaseStub{called: make(chan usecase.ExpireInventoryReservationsRequest, 1)}
	application := &App{
		Config:           cfg,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
		InventoryUseCase: stub,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	application.StartBackgroundWorkers(ctx)

	select {
	case request := <-stub.called:
		if request.Limit != 17 {
			t.Fatalf("expiry limit = %d, want 17", request.Limit)
		}
	case <-time.After(time.Second):
		t.Fatal("inventory expiry worker did not run immediately")
	}
}
