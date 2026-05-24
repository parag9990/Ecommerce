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

func TestSubmitProductForReviewCreatesReviewAndUpdatesProductStatus(t *testing.T) {
	productClient := &moderationProductClient{
		product: validModerationProduct(domain.ProductStatusDraft),
	}
	reviews := newMemoryReviewRepository()
	auditEvents := make([]domain.AuditEvent, 0)
	uc := newTestProductModerationUsecase(t, reviews, productClient, auditRecorderFunc(func(_ context.Context, event domain.AuditEvent) error {
		auditEvents = append(auditEvents, event)
		return nil
	}))

	result, err := uc.SubmitProductForReview(context.Background(), SubmitProductReviewInput{
		Actor:     activeActor(domain.RoleSellerCatalogEditor),
		ProductID: "prod_1",
	})
	if err != nil {
		t.Fatalf("SubmitProductForReview: %v", err)
	}
	if result.Status != domain.ProductStatusSubmitted {
		t.Fatalf("expected submitted status, got %+v", result)
	}
	if result.ReviewID == "" {
		t.Fatal("expected review id")
	}
	if productClient.lastUpdate.ToStatus != domain.ProductStatusSubmitted {
		t.Fatalf("expected product status update, got %+v", productClient.lastUpdate)
	}
	if len(auditEvents) == 0 || auditEvents[len(auditEvents)-1].Action != domain.AuditActionProductSubmittedReview {
		t.Fatalf("expected submitted review audit event, got %+v", auditEvents)
	}
}

func TestSubmitProductForReviewDeniesWrongSeller(t *testing.T) {
	product := validModerationProduct(domain.ProductStatusDraft)
	product.SellerID = "seller_2"
	uc := newTestProductModerationUsecase(t, newMemoryReviewRepository(), &moderationProductClient{product: product}, nil)

	_, err := uc.SubmitProductForReview(context.Background(), SubmitProductReviewInput{
		Actor:     activeActor(domain.RoleSeller),
		ProductID: "prod_1",
	})
	if !errors.Is(err, domain.ErrProductOwnershipMismatch) {
		t.Fatalf("expected ownership error, got %v", err)
	}
}

func TestDecideProductReviewRequiresRejectReason(t *testing.T) {
	reviews := newMemoryReviewRepository()
	reviews.review = domain.ProductModerationReview{
		ReviewID:    "review_1",
		ProductID:   "prod_1",
		SellerID:    "seller_1",
		Status:      domain.ProductReviewStatusSubmitted,
		SubmittedBy: "user_1",
		SubmittedAt: time.Now().UTC(),
	}
	uc := newTestProductModerationUsecase(t, reviews, &moderationProductClient{product: validModerationProduct(domain.ProductStatusSubmitted)}, nil)

	_, err := uc.DecideProductReview(context.Background(), DecideProductReviewInput{
		Actor:    adminActor(domain.RoleCatalogAdmin),
		ReviewID: "review_1",
		Decision: domain.ModerationDecisionReject,
	})
	if !errors.Is(err, domain.ErrRejectionReasonRequired) {
		t.Fatalf("expected rejection reason error, got %v", err)
	}
}

func TestDecideProductReviewApprovesSubmittedProduct(t *testing.T) {
	reviews := newMemoryReviewRepository()
	reviews.review = domain.ProductModerationReview{
		ReviewID:    "review_1",
		ProductID:   "prod_1",
		SellerID:    "seller_1",
		Status:      domain.ProductReviewStatusSubmitted,
		SubmittedBy: "user_1",
		SubmittedAt: time.Now().UTC(),
	}
	productClient := &moderationProductClient{product: validModerationProduct(domain.ProductStatusSubmitted)}
	uc := newTestProductModerationUsecase(t, reviews, productClient, nil)

	result, err := uc.DecideProductReview(context.Background(), DecideProductReviewInput{
		Actor:    adminActor(domain.RoleCatalogAdmin),
		ReviewID: "review_1",
		Decision: domain.ModerationDecisionApprove,
		Reason:   "Looks good.",
	})
	if err != nil {
		t.Fatalf("DecideProductReview: %v", err)
	}
	if result.Status != domain.ProductStatusApproved {
		t.Fatalf("expected approved status, got %+v", result)
	}
	if productClient.lastUpdate.ToStatus != domain.ProductStatusApproved {
		t.Fatalf("expected product approval update, got %+v", productClient.lastUpdate)
	}
	if reviews.review.Status != domain.ProductReviewStatusApproved {
		t.Fatalf("expected review approved, got %+v", reviews.review)
	}
}

func TestPublishApprovedProductRejectsDraft(t *testing.T) {
	uc := newTestProductModerationUsecase(t, newMemoryReviewRepository(), &moderationProductClient{product: validModerationProduct(domain.ProductStatusDraft)}, nil)

	_, err := uc.PublishApprovedProduct(context.Background(), PublishProductInput{
		Actor:     activeActor(domain.RoleSeller),
		ProductID: "prod_1",
	})
	if !errors.Is(err, domain.ErrInvalidProductTransition) {
		t.Fatalf("expected invalid transition, got %v", err)
	}
}

func newTestProductModerationUsecase(t *testing.T, reviews ProductModerationReviewRepository, productClient ProductClient, auditRecorder AuditRecorder) *ProductModerationUsecase {
	t.Helper()
	if auditRecorder == nil {
		auditRecorder = auditRecorderFunc(func(context.Context, domain.AuditEvent) error { return nil })
	}
	authorizer := newTestAuthorizer(t, nil, nil)
	uc, err := NewProductModerationUsecase(authorizer, reviews, productClient, auditRecorder, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewProductModerationUsecase: %v", err)
	}
	uc.now = func() time.Time { return time.Unix(1700000000, 0).UTC() }
	uc.newID = func(prefix string) string { return prefix + "_test" }
	return uc
}

type moderationProductClient struct {
	product    domain.Product
	lastUpdate domain.ProductStatusUpdate
}

func (c *moderationProductClient) GetProduct(context.Context, string) (domain.Product, error) {
	if c.product.ID == "" {
		return domain.Product{}, domain.ErrProductNotFound
	}
	return c.product, nil
}

func (c *moderationProductClient) UpdateProductStatus(_ context.Context, update domain.ProductStatusUpdate) (domain.Product, error) {
	c.lastUpdate = update
	c.product.Status = update.ToStatus
	return c.product, nil
}

func (c *moderationProductClient) PublishProduct(_ context.Context, update domain.ProductStatusUpdate) (domain.Product, error) {
	c.lastUpdate = update
	c.product.Status = domain.ProductStatusPublished
	return c.product, nil
}

func (c *moderationProductClient) UnpublishProduct(_ context.Context, update domain.ProductStatusUpdate) (domain.Product, error) {
	c.lastUpdate = update
	c.product.Status = domain.ProductStatusUnpublished
	return c.product, nil
}

type memoryReviewRepository struct {
	review domain.ProductModerationReview
}

func newMemoryReviewRepository() *memoryReviewRepository {
	return &memoryReviewRepository{}
}

func (r *memoryReviewRepository) CreateSubmittedReview(_ context.Context, review domain.ProductModerationReview) (domain.ProductModerationReview, error) {
	if r.review.ReviewID != "" && r.review.ProductID == review.ProductID && r.review.Status == domain.ProductReviewStatusSubmitted {
		return r.review, nil
	}
	r.review = review
	return r.review, nil
}

func (r *memoryReviewRepository) GetReview(_ context.Context, reviewID string) (domain.ProductModerationReview, error) {
	if r.review.ReviewID != reviewID {
		return domain.ProductModerationReview{}, domain.ErrReviewNotFound
	}
	return r.review, nil
}

func (r *memoryReviewRepository) MarkDecision(_ context.Context, reviewID string, decision domain.ModerationDecision, reviewedBy string, reason string, reviewedAt time.Time) (domain.ProductModerationReview, error) {
	if r.review.ReviewID != reviewID {
		return domain.ProductModerationReview{}, domain.ErrReviewNotFound
	}
	if r.review.Status != domain.ProductReviewStatusSubmitted {
		return domain.ProductModerationReview{}, domain.ErrReviewAlreadyDecided
	}
	r.review.Status = decision.ReviewStatus()
	r.review.ReviewedBy = reviewedBy
	r.review.RejectionReason = reason
	r.review.ReviewedAt = reviewedAt
	return r.review, nil
}

func (r *memoryReviewRepository) CancelSubmittedReview(_ context.Context, reviewID string, _ string, _ time.Time) error {
	if r.review.ReviewID == reviewID && r.review.Status == domain.ProductReviewStatusSubmitted {
		r.review.Status = domain.ProductReviewStatusCancelled
	}
	return nil
}

func (r *memoryReviewRepository) ListReviews(context.Context, ProductReviewFilter) ([]domain.ProductModerationReview, error) {
	if r.review.ReviewID == "" {
		return nil, nil
	}
	return []domain.ProductModerationReview{r.review}, nil
}

func validModerationProduct(status domain.ProductStatus) domain.Product {
	return domain.Product{
		ID:          "prod_1",
		SellerID:    "seller_1",
		Title:       "Running Shoes",
		Description: "Comfortable running shoes for daily training.",
		CategoryID:  "cat_1",
		Brand:       "Acme",
		Status:      status,
		Images: []domain.ProductImage{
			{URL: "https://cdn.example.com/prod_1/main.jpg"},
		},
		Variants: []domain.ProductVariant{
			{SKU: "SKU-1", Price: domain.Money{Amount: 120000, Currency: "INR"}},
		},
	}
}

func adminActor(role domain.Role) domain.ActorContext {
	return domain.ActorContext{
		UserID:      "admin_1",
		Roles:       []domain.Role{role},
		StaffStatus: domain.StaffStatusActive,
		RequestID:   "req_1",
	}
}
