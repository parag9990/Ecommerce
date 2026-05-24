package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

type ProductClient interface {
	GetProduct(ctx context.Context, productID string) (domain.Product, error)
	UpdateProductStatus(ctx context.Context, update domain.ProductStatusUpdate) (domain.Product, error)
	PublishProduct(ctx context.Context, update domain.ProductStatusUpdate) (domain.Product, error)
	UnpublishProduct(ctx context.Context, update domain.ProductStatusUpdate) (domain.Product, error)
}

type ProductModerationReviewRepository interface {
	CreateSubmittedReview(ctx context.Context, review domain.ProductModerationReview) (domain.ProductModerationReview, error)
	GetReview(ctx context.Context, reviewID string) (domain.ProductModerationReview, error)
	MarkDecision(ctx context.Context, reviewID string, decision domain.ModerationDecision, reviewedBy string, reason string, reviewedAt time.Time) (domain.ProductModerationReview, error)
	CancelSubmittedReview(ctx context.Context, reviewID string, reason string, cancelledAt time.Time) error
	ListReviews(ctx context.Context, filter ProductReviewFilter) ([]domain.ProductModerationReview, error)
}

type ProductModerationUsecase struct {
	authorizer    *Authorizer
	reviews       ProductModerationReviewRepository
	productClient ProductClient
	auditRecorder AuditRecorder
	logger        *slog.Logger
	now           func() time.Time
	newID         func(prefix string) string
}

type ProductReviewFilter struct {
	Status   domain.ProductReviewStatus
	SellerID string
	Limit    int
	Offset   int
}

type SubmitProductReviewInput struct {
	Actor     domain.ActorContext
	ProductID string
	RequestID string
}

type DecideProductReviewInput struct {
	Actor     domain.ActorContext
	ReviewID  string
	Decision  domain.ModerationDecision
	Reason    string
	RequestID string
}

type PublishProductInput struct {
	Actor     domain.ActorContext
	ProductID string
	RequestID string
}

type UnpublishProductInput struct {
	Actor     domain.ActorContext
	ProductID string
	Reason    string
	Force     bool
	RequestID string
}

type ListProductReviewsInput struct {
	Actor     domain.ActorContext
	Status    domain.ProductReviewStatus
	SellerID  string
	Limit     int
	Offset    int
	RequestID string
}

func NewProductModerationUsecase(authorizer *Authorizer, reviews ProductModerationReviewRepository, productClient ProductClient, auditRecorder AuditRecorder, logger *slog.Logger) (*ProductModerationUsecase, error) {
	if authorizer == nil {
		return nil, domain.ErrInvalidAuthorizationIn
	}
	if reviews == nil {
		return nil, domain.ErrReviewRepositoryRequired
	}
	if productClient == nil {
		return nil, domain.ErrProductClientRequired
	}
	if auditRecorder == nil {
		return nil, domain.ErrAuditRecorderRequired
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &ProductModerationUsecase{
		authorizer:    authorizer,
		reviews:       reviews,
		productClient: productClient,
		auditRecorder: auditRecorder,
		logger:        logger,
		now:           func() time.Time { return time.Now().UTC() },
		newID:         newModerationID,
	}, nil
}

func (uc *ProductModerationUsecase) SubmitProductForReview(ctx context.Context, input SubmitProductReviewInput) (domain.ProductModerationResult, error) {
	input = input.normalized()
	if input.ProductID == "" {
		return domain.ProductModerationResult{}, domain.NewValidationError("Invalid submit review request.", []domain.FieldViolation{{Field: "product_id", Reason: "Product id is required."}})
	}

	product, err := uc.productClient.GetProduct(ctx, input.ProductID)
	if err != nil {
		return domain.ProductModerationResult{}, err
	}
	product = product.Normalized()

	if err := uc.authorizeSellerAction(ctx, input.Actor, product, domain.PermissionProductsSubmitReview, input.requestID()); err != nil {
		return domain.ProductModerationResult{}, err
	}
	if err := domain.ValidateProductReadyForReview(product); err != nil {
		return domain.ProductModerationResult{}, err
	}

	currentStatus := product.Status.Normalized()
	if currentStatus == domain.ProductStatusSubmitted {
		review, err := uc.ensureSubmittedReview(ctx, product, input.Actor)
		if err != nil {
			return domain.ProductModerationResult{}, err
		}
		return domain.ProductModerationResult{
			ProductID: product.ID,
			SellerID:  product.SellerID,
			ReviewID:  review.ReviewID,
			Status:    currentStatus,
			Message:   "Product is already submitted for review.",
		}, nil
	}
	if err := domain.ValidateProductTransition(currentStatus, domain.ProductStatusSubmitted); err != nil {
		return domain.ProductModerationResult{}, err
	}

	review, err := uc.ensureSubmittedReview(ctx, product, input.Actor)
	if err != nil {
		return domain.ProductModerationResult{}, err
	}

	updated, err := uc.productClient.UpdateProductStatus(ctx, domain.ProductStatusUpdate{
		ProductID:   product.ID,
		SellerID:    product.SellerID,
		ActorUserID: input.Actor.UserID,
		FromStatus:  currentStatus,
		ToStatus:    domain.ProductStatusSubmitted,
		ReviewID:    review.ReviewID,
		RequestID:   input.requestID(),
	})
	if err != nil {
		_ = uc.reviews.CancelSubmittedReview(ctx, review.ReviewID, "product status update failed", uc.now())
		return domain.ProductModerationResult{}, err
	}

	if err := uc.recordAudit(ctx, domain.AuditActionProductSubmittedReview, input.Actor, product.SellerID, "product_review", review.ReviewID, input.requestID(), map[string]any{
		"product_id": product.ID,
		"status":     currentStatus,
	}, map[string]any{
		"product_id": product.ID,
		"status":     updated.Status.Normalized(),
		"review_id":  review.ReviewID,
	}); err != nil {
		return domain.ProductModerationResult{}, err
	}

	uc.logger.InfoContext(ctx, "cms.product_review.submitted",
		slog.String("product_id", product.ID),
		slog.String("seller_id", product.SellerID),
		slog.String("review_id", review.ReviewID),
		slog.String("request_id", input.requestID()),
	)

	return domain.ProductModerationResult{
		ProductID: updated.ID,
		SellerID:  updated.SellerID,
		ReviewID:  review.ReviewID,
		Status:    updated.Status.Normalized(),
		Message:   "Product submitted for review.",
	}, nil
}

func (uc *ProductModerationUsecase) DecideProductReview(ctx context.Context, input DecideProductReviewInput) (domain.ProductModerationResult, error) {
	input = input.normalized()
	if input.ReviewID == "" {
		return domain.ProductModerationResult{}, domain.NewValidationError("Invalid review decision request.", []domain.FieldViolation{{Field: "review_id", Reason: "Review id is required."}})
	}
	if err := uc.requireAdminPermission(input.Actor, domain.PermissionAdminCatalogModerate); err != nil {
		return domain.ProductModerationResult{}, err
	}
	if !input.Decision.Valid() {
		return domain.ProductModerationResult{}, domain.ErrInvalidModerationDecision
	}
	if input.Decision.Normalized() == domain.ModerationDecisionReject && input.Reason == "" {
		return domain.ProductModerationResult{}, domain.ErrRejectionReasonRequired
	}

	review, err := uc.reviews.GetReview(ctx, input.ReviewID)
	if err != nil {
		return domain.ProductModerationResult{}, err
	}
	review.Status = review.Status.Normalized()
	if review.Status != domain.ProductReviewStatusSubmitted {
		return domain.ProductModerationResult{}, domain.ErrReviewAlreadyDecided
	}

	product, err := uc.productClient.GetProduct(ctx, review.ProductID)
	if err != nil {
		return domain.ProductModerationResult{}, err
	}
	product = product.Normalized()
	targetStatus := input.Decision.ProductStatus()
	if err := domain.ValidateProductTransition(product.Status, targetStatus); err != nil {
		return domain.ProductModerationResult{}, err
	}

	updatedProduct, err := uc.productClient.UpdateProductStatus(ctx, domain.ProductStatusUpdate{
		ProductID:   product.ID,
		SellerID:    review.SellerID,
		ActorUserID: input.Actor.UserID,
		FromStatus:  product.Status,
		ToStatus:    targetStatus,
		ReviewID:    review.ReviewID,
		Reason:      input.Reason,
		RequestID:   input.requestID(),
	})
	if err != nil {
		return domain.ProductModerationResult{}, err
	}

	updatedReview, err := uc.reviews.MarkDecision(ctx, review.ReviewID, input.Decision, input.Actor.UserID, input.Reason, uc.now())
	if err != nil {
		return domain.ProductModerationResult{}, err
	}

	auditAction := domain.AuditActionProductApproved
	message := "Product approved."
	if input.Decision.Normalized() == domain.ModerationDecisionReject {
		auditAction = domain.AuditActionProductRejected
		message = "Product rejected. Seller can edit and resubmit."
	}
	if err := uc.recordAudit(ctx, auditAction, input.Actor, review.SellerID, "product_review", review.ReviewID, input.requestID(), map[string]any{
		"product_id": product.ID,
		"status":     product.Status,
		"review_id":  review.ReviewID,
	}, map[string]any{
		"product_id":       product.ID,
		"status":           updatedProduct.Status.Normalized(),
		"review_id":        updatedReview.ReviewID,
		"rejection_reason": updatedReview.RejectionReason,
	}); err != nil {
		return domain.ProductModerationResult{}, err
	}

	uc.logger.InfoContext(ctx, "cms.product_review.decided",
		slog.String("product_id", product.ID),
		slog.String("seller_id", review.SellerID),
		slog.String("review_id", review.ReviewID),
		slog.String("decision", string(input.Decision.Normalized())),
		slog.String("request_id", input.requestID()),
	)

	return domain.ProductModerationResult{
		ProductID: updatedProduct.ID,
		SellerID:  updatedProduct.SellerID,
		ReviewID:  updatedReview.ReviewID,
		Status:    updatedProduct.Status.Normalized(),
		Message:   message,
	}, nil
}

func (uc *ProductModerationUsecase) PublishApprovedProduct(ctx context.Context, input PublishProductInput) (domain.ProductModerationResult, error) {
	input = input.normalized()
	if input.ProductID == "" {
		return domain.ProductModerationResult{}, domain.NewValidationError("Invalid publish request.", []domain.FieldViolation{{Field: "product_id", Reason: "Product id is required."}})
	}

	product, err := uc.productClient.GetProduct(ctx, input.ProductID)
	if err != nil {
		return domain.ProductModerationResult{}, err
	}
	product = product.Normalized()
	if err := uc.authorizeSellerAction(ctx, input.Actor, product, domain.PermissionProductsSubmitReview, input.requestID()); err != nil {
		return domain.ProductModerationResult{}, err
	}

	if product.Status == domain.ProductStatusPublished {
		return domain.ProductModerationResult{
			ProductID: product.ID,
			SellerID:  product.SellerID,
			Status:    product.Status,
			Message:   "Product is already published.",
		}, nil
	}
	if err := domain.ValidateProductTransition(product.Status, domain.ProductStatusPublished); err != nil {
		return domain.ProductModerationResult{}, err
	}

	updated, err := uc.productClient.PublishProduct(ctx, domain.ProductStatusUpdate{
		ProductID:   product.ID,
		SellerID:    product.SellerID,
		ActorUserID: input.Actor.UserID,
		FromStatus:  product.Status,
		ToStatus:    domain.ProductStatusPublished,
		RequestID:   input.requestID(),
	})
	if err != nil {
		return domain.ProductModerationResult{}, err
	}

	if err := uc.recordAudit(ctx, domain.AuditActionProductPublished, input.Actor, product.SellerID, "product", product.ID, input.requestID(), map[string]any{
		"status": product.Status,
	}, map[string]any{
		"status": updated.Status.Normalized(),
	}); err != nil {
		return domain.ProductModerationResult{}, err
	}

	return domain.ProductModerationResult{
		ProductID: updated.ID,
		SellerID:  updated.SellerID,
		Status:    updated.Status.Normalized(),
		Message:   "Product published.",
	}, nil
}

func (uc *ProductModerationUsecase) UnpublishProduct(ctx context.Context, input UnpublishProductInput) (domain.ProductModerationResult, error) {
	input = input.normalized()
	if input.ProductID == "" {
		return domain.ProductModerationResult{}, domain.NewValidationError("Invalid unpublish request.", []domain.FieldViolation{{Field: "product_id", Reason: "Product id is required."}})
	}

	product, err := uc.productClient.GetProduct(ctx, input.ProductID)
	if err != nil {
		return domain.ProductModerationResult{}, err
	}
	product = product.Normalized()

	if input.Force {
		if err := uc.requireAdminPermission(input.Actor, domain.PermissionAdminForceUnpublish); err != nil {
			return domain.ProductModerationResult{}, err
		}
		if input.Reason == "" {
			return domain.ProductModerationResult{}, domain.ErrForceUnpublishReasonNeeded
		}
	} else if err := uc.authorizeSellerAction(ctx, input.Actor, product, domain.PermissionProductsUnpublish, input.requestID()); err != nil {
		return domain.ProductModerationResult{}, err
	}

	if product.Status == domain.ProductStatusUnpublished {
		return domain.ProductModerationResult{
			ProductID: product.ID,
			SellerID:  product.SellerID,
			Status:    product.Status,
			Message:   "Product is already unpublished.",
		}, nil
	}
	if err := domain.ValidateProductTransition(product.Status, domain.ProductStatusUnpublished); err != nil {
		return domain.ProductModerationResult{}, err
	}

	updated, err := uc.productClient.UnpublishProduct(ctx, domain.ProductStatusUpdate{
		ProductID:      product.ID,
		SellerID:       product.SellerID,
		ActorUserID:    input.Actor.UserID,
		FromStatus:     product.Status,
		ToStatus:       domain.ProductStatusUnpublished,
		Reason:         input.Reason,
		RequestID:      input.requestID(),
		ForceUnpublish: input.Force,
	})
	if err != nil {
		return domain.ProductModerationResult{}, err
	}

	if err := uc.recordAudit(ctx, domain.AuditActionProductUnpublished, input.Actor, product.SellerID, "product", product.ID, input.requestID(), map[string]any{
		"status": product.Status,
	}, map[string]any{
		"status":          updated.Status.Normalized(),
		"reason":          input.Reason,
		"force_unpublish": input.Force,
	}); err != nil {
		return domain.ProductModerationResult{}, err
	}

	return domain.ProductModerationResult{
		ProductID: updated.ID,
		SellerID:  updated.SellerID,
		Status:    updated.Status.Normalized(),
		Message:   "Product unpublished.",
	}, nil
}

func (uc *ProductModerationUsecase) ListProductReviews(ctx context.Context, input ListProductReviewsInput) ([]domain.ProductModerationReview, error) {
	input = input.normalized()
	if err := uc.requireAdminPermission(input.Actor, domain.PermissionAdminCatalogModerate); err != nil {
		return nil, err
	}
	return uc.reviews.ListReviews(ctx, ProductReviewFilter{
		Status:   input.Status,
		SellerID: input.SellerID,
		Limit:    input.Limit,
		Offset:   input.Offset,
	})
}

func (uc *ProductModerationUsecase) authorizeSellerAction(ctx context.Context, actor domain.ActorContext, product domain.Product, permission domain.Permission, requestID string) error {
	if product.SellerID == "" || actor.Normalized().SellerID != product.SellerID {
		return domain.ErrProductOwnershipMismatch
	}
	_, err := uc.authorizer.Authorize(ctx, AuthorizeInput{
		Actor:              actor,
		ResourceSellerID:   product.SellerID,
		RequiredPermission: permission,
		ResourceType:       "product",
		ResourceID:         product.ID,
		RequestID:          requestID,
	})
	return err
}

func (uc *ProductModerationUsecase) requireAdminPermission(actor domain.ActorContext, permission domain.Permission) error {
	actor = actor.Normalized()
	if !actor.Authenticated() {
		return domain.ErrUnauthenticated
	}
	if !domain.RolesHavePermission(actor.Roles, permission) {
		return domain.ErrForbidden
	}
	return nil
}

func (uc *ProductModerationUsecase) ensureSubmittedReview(ctx context.Context, product domain.Product, actor domain.ActorContext) (domain.ProductModerationReview, error) {
	now := uc.now()
	return uc.reviews.CreateSubmittedReview(ctx, domain.ProductModerationReview{
		ReviewID:    uc.newID("review"),
		ProductID:   product.ID,
		SellerID:    product.SellerID,
		Status:      domain.ProductReviewStatusSubmitted,
		SubmittedBy: actor.UserID,
		SubmittedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (uc *ProductModerationUsecase) recordAudit(ctx context.Context, action domain.Permission, actor domain.ActorContext, sellerID string, resourceType string, resourceID string, requestID string, before map[string]any, after map[string]any) error {
	actor = actor.Normalized()
	event := domain.AuditEvent{
		AuditID:          uc.newID("audit"),
		ActorUserID:      actor.UserID,
		ActorSellerID:    actor.SellerID,
		ActorRoles:       actor.Roles,
		Action:           action,
		ResourceType:     resourceType,
		ResourceID:       resourceID,
		ResourceSellerID: sellerID,
		RequestID:        requestID,
		Decision:         domain.AuditDecisionAllowed,
		Before:           before,
		After:            after,
		CreatedAt:        uc.now(),
	}
	if event.ActorSellerID == "" {
		event.ActorSellerID = sellerID
	}
	if err := uc.auditRecorder.Record(ctx, event); err != nil {
		uc.logger.ErrorContext(ctx, "cms.product_moderation.audit_failed",
			slog.String("action", action.String()),
			slog.String("resource_type", resourceType),
			slog.String("resource_id", resourceID),
			slog.String("request_id", requestID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%w: %v", domain.ErrAuditWriteFailed, err)
	}
	return nil
}

func (i SubmitProductReviewInput) normalized() SubmitProductReviewInput {
	i.Actor = i.Actor.Normalized()
	i.ProductID = strings.TrimSpace(i.ProductID)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i SubmitProductReviewInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i DecideProductReviewInput) normalized() DecideProductReviewInput {
	i.Actor = i.Actor.Normalized()
	i.ReviewID = strings.TrimSpace(i.ReviewID)
	i.Decision = i.Decision.Normalized()
	i.Reason = domain.NormalizeModerationReason(i.Reason)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i DecideProductReviewInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i PublishProductInput) normalized() PublishProductInput {
	i.Actor = i.Actor.Normalized()
	i.ProductID = strings.TrimSpace(i.ProductID)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i PublishProductInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i UnpublishProductInput) normalized() UnpublishProductInput {
	i.Actor = i.Actor.Normalized()
	i.ProductID = strings.TrimSpace(i.ProductID)
	i.Reason = domain.NormalizeModerationReason(i.Reason)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i UnpublishProductInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i ListProductReviewsInput) normalized() ListProductReviewsInput {
	i.Actor = i.Actor.Normalized()
	i.Status = i.Status.Normalized()
	i.SellerID = strings.TrimSpace(i.SellerID)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func newModerationID(prefix string) string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s_%d", strings.TrimSpace(prefix), time.Now().UTC().UnixNano())
	}
	return strings.TrimSpace(prefix) + "_" + hex.EncodeToString(bytes[:])
}

func MapProductClientStatusError(statusCode int, body string) error {
	body = strings.TrimSpace(body)
	switch statusCode {
	case 404:
		return domain.ErrProductNotFound
	case 409, 412:
		if body == "" {
			return domain.ErrInvalidProductTransition
		}
		return fmt.Errorf("%w: %s", domain.ErrInvalidProductTransition, body)
	case 422:
		if body == "" {
			return domain.ErrValidationFailed
		}
		return fmt.Errorf("%w: %s", domain.ErrValidationFailed, body)
	case 503, 504:
		return domain.ErrProductServiceUnavailable
	default:
		if statusCode >= 500 {
			return domain.ErrProductServiceUnavailable
		}
		if body != "" {
			return errors.New(body)
		}
		return fmt.Errorf("product service returned status %d", statusCode)
	}
}
