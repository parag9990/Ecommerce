package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

func TestSchemaUsecaseEnsureCollections(t *testing.T) {
	manager := &fakeCollectionManager{
		report: domain.CollectionReport{
			DatabaseName:      domain.CartDatabaseName,
			CollectionName:    domain.CartCollectionName,
			CollectionCreated: true,
			IndexesEnsured:    []string{"idx_carts_user_status"},
		},
	}
	uc, err := NewSchemaUsecase(manager, nil)
	if err != nil {
		t.Fatalf("NewSchemaUsecase() error = %v", err)
	}

	report, err := uc.EnsureCollections(context.Background())
	if err != nil {
		t.Fatalf("EnsureCollections() error = %v", err)
	}
	if !manager.ensureCalled {
		t.Fatal("EnsureCollections() did not call manager")
	}
	if report.CollectionName != domain.CartCollectionName {
		t.Fatalf("CollectionName = %q, want %q", report.CollectionName, domain.CartCollectionName)
	}
}

func TestSchemaUsecaseReadyPropagatesPingError(t *testing.T) {
	wantErr := errors.New("mongo unavailable")
	manager := &fakeCollectionManager{pingErr: wantErr}
	uc, err := NewSchemaUsecase(manager, nil)
	if err != nil {
		t.Fatalf("NewSchemaUsecase() error = %v", err)
	}

	err = uc.Ready(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Ready() error = %v, want %v", err, wantErr)
	}
}

type fakeCollectionManager struct {
	report       domain.CollectionReport
	ensureErr    error
	pingErr      error
	ensureCalled bool
}

func (f *fakeCollectionManager) EnsureCartCollections(ctx context.Context) (domain.CollectionReport, error) {
	f.ensureCalled = true
	return f.report, f.ensureErr
}

func (f *fakeCollectionManager) Ping(ctx context.Context) error {
	return f.pingErr
}
