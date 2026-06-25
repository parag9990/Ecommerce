package usecase

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

func TestCreateSynonymUsecaseNormalizesAndUpserts(t *testing.T) {
	uc, repo := newCreateSynonymTestUsecase(t)

	got, err := uc.Execute(context.Background(), domain.SearchSynonymInput{
		Root:     " Mobile ",
		Synonyms: []string{" Phone ", "SMARTPHONE"},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if repo.upserted.ID != "syn_mobile" {
		t.Fatalf("upserted id = %q", repo.upserted.ID)
	}
	if got.Root != "mobile" || len(got.Synonyms) != 2 || got.Synonyms[0] != "phone" {
		t.Fatalf("synonym = %#v", got)
	}
}

func TestCreateSynonymUsecaseRejectsInvalidInput(t *testing.T) {
	uc, repo := newCreateSynonymTestUsecase(t)

	_, err := uc.Execute(context.Background(), domain.SearchSynonymInput{
		Root:     "m",
		Synonyms: []string{"phone"},
	})
	if !errors.Is(err, domain.ErrInvalidSynonym) {
		t.Fatalf("expected invalid synonym, got %v", err)
	}
	if repo.upsertCalls != 0 {
		t.Fatalf("repository should not be called")
	}
}

func TestCreateSynonymUsecaseReturnsRepositoryError(t *testing.T) {
	uc, repo := newCreateSynonymTestUsecase(t)
	repo.upsertErr = domain.ErrSearchBackendUnavailable

	_, err := uc.Execute(context.Background(), domain.SearchSynonymInput{
		Root:     "mobile",
		Synonyms: []string{"phone"},
	})
	if !errors.Is(err, domain.ErrSearchBackendUnavailable) {
		t.Fatalf("expected backend error, got %v", err)
	}
}

func TestListSynonymsUsecaseNormalizesPagination(t *testing.T) {
	repo := &fakeSynonymRepository{
		listed: []domain.SearchSynonym{{ID: "syn_mobile", Root: "mobile", Synonyms: []string{"phone"}}},
	}
	uc, err := NewListSynonymsUsecase(repo, ListSynonymsOptions{Timeout: time.Second}, slog.Default())
	if err != nil {
		t.Fatalf("usecase: %v", err)
	}

	got, err := uc.Execute(context.Background(), domain.SearchSynonymPageRequest{Page: 0, PageSize: 1000})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if repo.page.Page != domain.DefaultSynonymPage {
		t.Fatalf("page = %d", repo.page.Page)
	}
	if repo.page.PageSize != domain.MaxSynonymPageSize {
		t.Fatalf("page_size = %d", repo.page.PageSize)
	}
	if len(got) != 1 || got[0].ID != "syn_mobile" {
		t.Fatalf("synonyms = %#v", got)
	}
}

func TestUpdateSynonymUsecaseUpsertsAndDeletesRenamedSynonym(t *testing.T) {
	repo := &fakeSynonymRepository{
		before: domain.SearchSynonym{ID: "syn_mobile", Root: "mobile", Synonyms: []string{"phone"}},
		found:  true,
	}
	uc, err := NewUpdateSynonymUsecase(repo, UpdateSynonymOptions{Timeout: time.Second}, slog.Default())
	if err != nil {
		t.Fatalf("usecase: %v", err)
	}

	got, err := uc.Execute(context.Background(), "syn_mobile", domain.SearchSynonymInput{
		Root:     "cell phone",
		Synonyms: []string{"smartphone"},
		Reason:   "rename root",
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if got.ID != "syn_cell_phone" {
		t.Fatalf("updated id = %q", got.ID)
	}
	if repo.deletedID != "syn_mobile" {
		t.Fatalf("deleted id = %q", repo.deletedID)
	}
}

func TestUpdateSynonymUsecaseReturnsNotFound(t *testing.T) {
	repo := &fakeSynonymRepository{}
	uc, err := NewUpdateSynonymUsecase(repo, UpdateSynonymOptions{Timeout: time.Second}, slog.Default())
	if err != nil {
		t.Fatalf("usecase: %v", err)
	}

	_, err = uc.Execute(context.Background(), "syn_missing", domain.SearchSynonymInput{
		Root:     "mobile",
		Synonyms: []string{"phone"},
	})
	if !errors.Is(err, domain.ErrSearchSynonymNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestDeleteSynonymUsecaseDeletesExistingSynonym(t *testing.T) {
	repo := &fakeSynonymRepository{
		before: domain.SearchSynonym{ID: "syn_mobile", Root: "mobile", Synonyms: []string{"phone"}},
		found:  true,
	}
	uc, err := NewDeleteSynonymUsecase(repo, DeleteSynonymOptions{Timeout: time.Second}, slog.Default())
	if err != nil {
		t.Fatalf("usecase: %v", err)
	}

	got, err := uc.Execute(context.Background(), "syn_mobile", "cleanup")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if got.ID != "syn_mobile" || repo.deletedID != "syn_mobile" {
		t.Fatalf("deleted = %#v, id=%q", got, repo.deletedID)
	}
}

func newCreateSynonymTestUsecase(t *testing.T) (*CreateSynonymUsecase, *fakeSynonymRepository) {
	t.Helper()
	repo := &fakeSynonymRepository{}
	uc, err := NewCreateSynonymUsecase(repo, CreateSynonymOptions{Timeout: time.Second}, slog.Default())
	if err != nil {
		t.Fatalf("usecase: %v", err)
	}
	return uc, repo
}

type fakeSynonymRepository struct {
	before      domain.SearchSynonym
	found       bool
	getErr      error
	upsertCalls int
	upserted    domain.SearchSynonym
	upsertErr   error
	deleteErr   error
	deletedID   string
	listCalls   int
	page        domain.SearchSynonymPageRequest
	listed      []domain.SearchSynonym
	listErr     error
}

func (r *fakeSynonymRepository) GetSynonym(_ context.Context, _ string) (domain.SearchSynonym, bool, error) {
	return r.before, r.found, r.getErr
}

func (r *fakeSynonymRepository) UpsertSynonym(_ context.Context, synonym domain.SearchSynonym) (domain.SearchSynonym, error) {
	r.upsertCalls++
	r.upserted = synonym
	if r.upsertErr != nil {
		return domain.SearchSynonym{}, r.upsertErr
	}
	return synonym, nil
}

func (r *fakeSynonymRepository) DeleteSynonym(_ context.Context, id string) (domain.SearchSynonym, bool, error) {
	r.deletedID = id
	if r.deleteErr != nil {
		return domain.SearchSynonym{}, false, r.deleteErr
	}
	if r.found && id == r.before.ID {
		return r.before, true, nil
	}
	return domain.SearchSynonym{}, false, nil
}

func (r *fakeSynonymRepository) ListSynonyms(_ context.Context, page domain.SearchSynonymPageRequest) ([]domain.SearchSynonym, error) {
	r.listCalls++
	r.page = page
	if r.listErr != nil {
		return nil, r.listErr
	}
	return append([]domain.SearchSynonym(nil), r.listed...), nil
}
