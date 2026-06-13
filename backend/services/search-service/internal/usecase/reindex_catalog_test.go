package usecase

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/repository"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/schema"
)

func TestVersionedCollectionNameUsesUTC(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 30, 45, 0, time.FixedZone("IST", 5*60*60+30*60))
	if got := VersionedCollectionName("products", now); got != "products_20260524_070045" {
		t.Fatalf("collection name = %q", got)
	}
}

func TestReindexCatalogUsecaseAliasSuccessSwapsAfterValidation(t *testing.T) {
	uc, products, index, lock := newReindexTestUsecase(t)
	products.pages = []domain.ProductExportPage{
		{Items: []domain.ProductIndexPayload{validProductPayload("prod_1")}, NextCursor: "cursor_2", HasMore: true, Total: 2},
		{Items: []domain.ProductIndexPayload{validProductPayload("prod_2")}, HasMore: false, Total: 2},
	}
	index.previousAlias = "products_20260523_120000"
	index.count = 2

	result, err := uc.Execute(context.Background(), domain.ReindexRequest{Mode: domain.ReindexModeAlias, BatchSize: 1, Reason: "schema_change"})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if !lock.acquired || !lock.released {
		t.Fatalf("lock state acquired=%v released=%v", lock.acquired, lock.released)
	}
	if len(index.ensureCollections) != 1 || index.ensureCollections[0] != "products_20260524_120000" {
		t.Fatalf("ensured collections = %#v", index.ensureCollections)
	}
	if len(index.imported) != 2 {
		t.Fatalf("imported docs = %#v", index.imported)
	}
	if index.swappedAlias != "products" || index.swappedCollection != "products_20260524_120000" {
		t.Fatalf("swap = %s -> %s", index.swappedAlias, index.swappedCollection)
	}
	if !result.AliasSwapped || result.PreviousCollection != "products_20260523_120000" || result.ProductsIndexed != 2 {
		t.Fatalf("result = %#v", result)
	}
}

func TestReindexCatalogUsecaseDryRunDoesNotWriteOrSwap(t *testing.T) {
	uc, products, index, _ := newReindexTestUsecase(t)
	products.pages = []domain.ProductExportPage{
		{Items: []domain.ProductIndexPayload{validProductPayload("prod_1")}, HasMore: false, Total: 1},
	}

	result, err := uc.Execute(context.Background(), domain.ReindexRequest{Mode: domain.ReindexModeAlias, BatchSize: 10, Reason: "verify", DryRun: true})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(index.ensureCollections) != 0 || len(index.imported) != 0 || index.swappedAlias != "" {
		t.Fatalf("dry run wrote to index: %#v", index)
	}
	if result.ProductsIndexed != 1 || result.AliasSwapped {
		t.Fatalf("result = %#v", result)
	}
}

func TestReindexCatalogUsecaseLockAlreadyRunning(t *testing.T) {
	uc, _, index, lock := newReindexTestUsecase(t)
	lock.allowAcquire = false

	_, err := uc.Execute(context.Background(), domain.ReindexRequest{Mode: domain.ReindexModeAlias, BatchSize: 10, Reason: "manual"})
	if !errors.Is(err, domain.ErrReindexAlreadyRunning) {
		t.Fatalf("expected already running, got %v", err)
	}
	if len(index.imported) != 0 || index.swappedAlias != "" {
		t.Fatalf("index should not be touched: %#v", index)
	}
}

func TestReindexCatalogUsecaseValidationFailureDoesNotSwapAlias(t *testing.T) {
	uc, products, index, _ := newReindexTestUsecase(t)
	products.pages = []domain.ProductExportPage{
		{Items: []domain.ProductIndexPayload{validProductPayload("prod_1")}, HasMore: false, Total: 1},
	}
	index.count = 0

	_, err := uc.Execute(context.Background(), domain.ReindexRequest{Mode: domain.ReindexModeAlias, BatchSize: 10, Reason: "manual"})
	if !errors.Is(err, domain.ErrReindexValidationFailed) {
		t.Fatalf("expected validation failure, got %v", err)
	}
	if index.swappedAlias != "" {
		t.Fatalf("alias should not be swapped: %s", index.swappedAlias)
	}
}

func TestReindexCatalogUsecaseRejectsNonAdvancingProductCursor(t *testing.T) {
	uc, products, _, _ := newReindexTestUsecase(t)
	products.pages = []domain.ProductExportPage{
		{Items: []domain.ProductIndexPayload{validProductPayload("prod_1")}, NextCursor: "same", HasMore: true, Total: 2},
		{Items: []domain.ProductIndexPayload{validProductPayload("prod_2")}, NextCursor: "same", HasMore: true, Total: 2},
	}

	_, err := uc.Execute(context.Background(), domain.ReindexRequest{Mode: domain.ReindexModeAlias, BatchSize: 1, Reason: "manual"})
	if !errors.Is(err, domain.ErrProductCatalogExportUnavailable) {
		t.Fatalf("expected product export error, got %v", err)
	}
}

func newReindexTestUsecase(t *testing.T) (*ReindexCatalogUsecase, *fakeProductCatalogReader, *fakeReindexRepository, *fakeReindexLock) {
	t.Helper()
	collection, policy, synonyms := schema.MustProductSchemaContract()
	schemaRepo, err := repository.NewStaticSchemaRepository(collection, policy, synonyms)
	if err != nil {
		t.Fatalf("schema repo: %v", err)
	}
	products := &fakeProductCatalogReader{}
	index := &fakeReindexRepository{count: 1}
	lock := &fakeReindexLock{allowAcquire: true}
	uc, err := NewReindexCatalogUsecase(schemaRepo, products, index, lock, ReindexCatalogOptions{
		ProductsAlias:          "products",
		DefaultMode:            domain.ReindexModeAlias,
		DefaultBatchSize:       10,
		MaxBatchSize:           100,
		CollectionPrefix:       "products",
		LockTTL:                time.Hour,
		JobTimeout:             time.Minute,
		ProductPageTimeout:     time.Second,
		ImportTimeout:          time.Second,
		OldCollectionRetention: time.Hour,
		Clock: func() time.Time {
			return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
		},
	}, slog.Default())
	if err != nil {
		t.Fatalf("usecase: %v", err)
	}
	return uc, products, index, lock
}

func validProductPayload(productID string) domain.ProductIndexPayload {
	price := 1999.0
	rating := 4.5
	popularity := int32(100)
	inStock := true
	createdAt := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	return domain.ProductIndexPayload{
		ProductID:       productID,
		Title:           "Running Shoes",
		Description:     "Lightweight running shoes",
		Brand:           "Nike",
		CategoryIDs:     []string{"cat_shoes"},
		SellerID:        "seller_1",
		Price:           &price,
		Rating:          &rating,
		PopularityScore: &popularity,
		InStock:         &inStock,
		Status:          domain.ProductStatusPublished,
		CreatedAt:       &createdAt,
		UpdatedAt:       createdAt.Add(time.Hour),
	}
}

type fakeProductCatalogReader struct {
	pages []domain.ProductExportPage
	reqs  []domain.ProductExportRequest
	err   error
}

func (r *fakeProductCatalogReader) ListSearchableProducts(_ context.Context, req domain.ProductExportRequest) (domain.ProductExportPage, error) {
	r.reqs = append(r.reqs, req)
	if r.err != nil {
		return domain.ProductExportPage{}, r.err
	}
	if len(r.pages) == 0 {
		return domain.ProductExportPage{}, nil
	}
	page := r.pages[0]
	r.pages = r.pages[1:]
	return page, nil
}

type fakeReindexRepository struct {
	previousAlias     string
	ensureCollections []string
	imported          []domain.ProductDocument
	count             int
	countErr          error
	smokeErr          error
	swappedAlias      string
	swappedCollection string
	cleanupCalled     bool
}

func (r *fakeReindexRepository) EnsureCollection(_ context.Context, collection domain.CollectionSchema) error {
	r.ensureCollections = append(r.ensureCollections, collection.Name)
	return nil
}

func (r *fakeReindexRepository) ImportProducts(_ context.Context, _ string, docs []domain.ProductDocument) error {
	r.imported = append(r.imported, docs...)
	return nil
}

func (r *fakeReindexRepository) CountDocuments(context.Context, string) (int, error) {
	return r.count, r.countErr
}

func (r *fakeReindexRepository) SmokeSearch(context.Context, string) error {
	return r.smokeErr
}

func (r *fakeReindexRepository) SwapAlias(_ context.Context, alias string, collection string) error {
	r.swappedAlias = alias
	r.swappedCollection = collection
	return nil
}

func (r *fakeReindexRepository) ResolveAlias(context.Context, string) (string, error) {
	return r.previousAlias, nil
}

func (r *fakeReindexRepository) CleanupOldCollections(context.Context, string, string, string, time.Duration, time.Time) ([]string, error) {
	r.cleanupCalled = true
	return []string{}, nil
}

type fakeReindexLock struct {
	allowAcquire bool
	acquired     bool
	released     bool
}

func (l *fakeReindexLock) Acquire(context.Context, string, time.Duration) (bool, error) {
	l.acquired = true
	return l.allowAcquire, nil
}

func (l *fakeReindexLock) Refresh(context.Context, string, time.Duration) (bool, error) {
	return true, nil
}

func (l *fakeReindexLock) Release(context.Context, string) error {
	l.released = true
	return nil
}
