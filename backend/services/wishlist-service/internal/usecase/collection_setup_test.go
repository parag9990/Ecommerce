package usecase

import (
	"context"
	"errors"
	"testing"
)

func TestCollectionSetupServiceEnsuresReadiness(t *testing.T) {
	repository := &fakeCollectionRepository{}
	service, err := NewCollectionSetupService(repository, nil)
	if err != nil {
		t.Fatalf("NewCollectionSetupService returned error: %v", err)
	}

	if err := service.EnsureReady(context.Background()); err != nil {
		t.Fatalf("EnsureReady returned error: %v", err)
	}
	if !repository.pinged {
		t.Fatal("repository Ping was not called")
	}
	if !repository.ensured {
		t.Fatal("repository EnsureCollection was not called")
	}
}

func TestCollectionSetupServiceStopsOnPingFailure(t *testing.T) {
	wantErr := errors.New("mongo down")
	repository := &fakeCollectionRepository{pingErr: wantErr}
	service, err := NewCollectionSetupService(repository, nil)
	if err != nil {
		t.Fatalf("NewCollectionSetupService returned error: %v", err)
	}

	err = service.EnsureReady(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("EnsureReady error = %v, want %v", err, wantErr)
	}
	if repository.ensured {
		t.Fatal("EnsureCollection was called after Ping failed")
	}
}

type fakeCollectionRepository struct {
	pinged    bool
	ensured   bool
	pingErr   error
	ensureErr error
}

func (f *fakeCollectionRepository) Ping(ctx context.Context) error {
	f.pinged = true
	return f.pingErr
}

func (f *fakeCollectionRepository) EnsureCollection(ctx context.Context) error {
	f.ensured = true
	return f.ensureErr
}

func (f *fakeCollectionRepository) CollectionName() string {
	return "wishlists"
}
