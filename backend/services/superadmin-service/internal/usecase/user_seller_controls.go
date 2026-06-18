package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

type ControlAuthorizer interface {
	RequirePermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission) error
	RequireHighRiskPermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission, reason string) error
}

type UserServiceClient interface {
	ListUsersForAdmin(ctx context.Context, req domain.AdminUserListRequest, actor domain.AdminActor) (domain.AdminUserListResponse, error)
	GetUserForAdmin(ctx context.Context, userID string, actor domain.AdminActor) (domain.UserProfile, error)
	UpdateUserStatus(ctx context.Context, userID string, status domain.UserStatus, mutation domain.AdminMutationContext) error
	ListSellersForAdmin(ctx context.Context, req domain.AdminSellerListRequest, actor domain.AdminActor) (domain.AdminSellerListResponse, error)
	GetSellerForAdmin(ctx context.Context, sellerID string, actor domain.AdminActor) (domain.SellerProfile, error)
	UpdateSellerStatus(ctx context.Context, sellerID string, status domain.SellerStatus, mutation domain.AdminMutationContext) error
}

type ReviewTaskRepository interface {
	CloseForResource(ctx context.Context, req domain.CloseReviewTaskRequest) error
}

type AuditRecorder interface {
	RecordAdminMutation(ctx context.Context, record domain.AuditRecord) error
}

type ControlService struct {
	authorizer  ControlAuthorizer
	userClient  UserServiceClient
	reviewTasks ReviewTaskRepository
	audit       AuditRecorder
	logger      logging.Logger
}

func NewControlService(authorizer ControlAuthorizer, userClient UserServiceClient, reviewTasks ReviewTaskRepository, audit AuditRecorder, logger logging.Logger) (*ControlService, error) {
	if authorizer == nil {
		return nil, errors.New("control service requires authorizer")
	}
	if userClient == nil {
		return nil, errors.New("control service requires user service client")
	}
	if logger == nil {
		logger = logging.NewNop()
	}
	if audit == nil {
		audit = NewLoggingAuditRecorder(logger)
	}
	return &ControlService{
		authorizer:  authorizer,
		userClient:  userClient,
		reviewTasks: reviewTasks,
		audit:       audit,
		logger:      logger,
	}, nil
}

func (s *ControlService) ListUsersForAdmin(ctx context.Context, req domain.AdminUserListRequest) (domain.AdminUserListResponse, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.AdminUserListResponse{}, err
	}
	if err := s.authorizer.RequirePermission(ctx, actor, domain.PermissionUsersRead); err != nil {
		return domain.AdminUserListResponse{}, err
	}

	cleanReq, err := req.Normalize()
	if err != nil {
		return domain.AdminUserListResponse{}, err
	}
	return s.userClient.ListUsersForAdmin(ctx, cleanReq, actor)
}

func (s *ControlService) UpdateUserStatus(ctx context.Context, userID string, req domain.StatusUpdateRequest) (domain.SuccessResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return domain.SuccessResponse{}, domain.NewValidationError("user_id is required")
	}

	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.SuccessResponse{}, err
	}
	reason, err := domain.NormalizeMutationReason(req.Reason)
	if err != nil {
		return domain.SuccessResponse{}, err
	}
	if err := s.authorizer.RequireHighRiskPermission(ctx, actor, domain.PermissionUsersStatusUpdate, reason); err != nil {
		return domain.SuccessResponse{}, err
	}

	nextStatus, err := domain.ParseUserStatus(req.Status)
	if err != nil {
		return domain.SuccessResponse{}, err
	}

	user, err := s.userClient.GetUserForAdmin(ctx, userID, actor)
	if err != nil {
		return domain.SuccessResponse{}, err
	}
	if !user.Status.Valid() {
		return domain.SuccessResponse{}, domain.NewInternal(fmt.Sprintf("user service returned unknown user status %q", user.Status), nil)
	}
	if !domain.CanUpdateUserStatus(user.Status, nextStatus) {
		return domain.SuccessResponse{}, domain.NewInvalidStatusTransition("user", string(user.Status), string(nextStatus))
	}

	mutation := domain.AdminMutationContext{Actor: actor, Reason: reason}
	if err := s.userClient.UpdateUserStatus(ctx, userID, nextStatus, mutation); err != nil {
		return domain.SuccessResponse{}, err
	}

	if err := s.recordAudit(ctx, domain.AuditRecord{
		ActorAdminID: actor.AdminID,
		Action:       domain.UserStatusAuditAction(nextStatus),
		ResourceType: "user",
		ResourceID:   userID,
		RequestID:    actor.RequestID,
		SessionID:    actor.SessionID,
		IPHash:       actor.IPHash,
		Reason:       reason,
		Before:       map[string]string{"status": string(user.Status)},
		After:        map[string]string{"status": string(nextStatus)},
	}); err != nil {
		return domain.SuccessResponse{}, err
	}

	return domain.SuccessResponse{Success: true}, nil
}

func (s *ControlService) ListSellersForAdmin(ctx context.Context, req domain.AdminSellerListRequest) (domain.AdminSellerListResponse, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.AdminSellerListResponse{}, err
	}
	if err := s.authorizer.RequirePermission(ctx, actor, domain.PermissionSellersRead); err != nil {
		return domain.AdminSellerListResponse{}, err
	}

	cleanReq, err := req.Normalize()
	if err != nil {
		return domain.AdminSellerListResponse{}, err
	}
	return s.userClient.ListSellersForAdmin(ctx, cleanReq, actor)
}

func (s *ControlService) UpdateSellerStatus(ctx context.Context, sellerID string, req domain.StatusUpdateRequest) (domain.SuccessResponse, error) {
	sellerID = strings.TrimSpace(sellerID)
	if sellerID == "" {
		return domain.SuccessResponse{}, domain.NewValidationError("seller_id is required")
	}

	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.SuccessResponse{}, err
	}
	reason, err := domain.NormalizeMutationReason(req.Reason)
	if err != nil {
		return domain.SuccessResponse{}, err
	}
	if err := s.authorizer.RequireHighRiskPermission(ctx, actor, domain.PermissionSellersStatusUpdate, reason); err != nil {
		return domain.SuccessResponse{}, err
	}

	nextStatus, err := domain.ParseSellerStatus(req.Status)
	if err != nil {
		return domain.SuccessResponse{}, err
	}

	seller, err := s.userClient.GetSellerForAdmin(ctx, sellerID, actor)
	if err != nil {
		return domain.SuccessResponse{}, err
	}
	if !seller.Status.Valid() {
		return domain.SuccessResponse{}, domain.NewInternal(fmt.Sprintf("user service returned unknown seller status %q", seller.Status), nil)
	}
	if !domain.CanUpdateSellerStatus(seller.Status, nextStatus) {
		return domain.SuccessResponse{}, domain.NewInvalidStatusTransition("seller", string(seller.Status), string(nextStatus))
	}

	mutation := domain.AdminMutationContext{Actor: actor, Reason: reason}
	if err := s.userClient.UpdateSellerStatus(ctx, sellerID, nextStatus, mutation); err != nil {
		return domain.SuccessResponse{}, err
	}

	s.closeSellerKYCReviewTask(ctx, seller, nextStatus, actor, reason)
	if err := s.recordAudit(ctx, domain.AuditRecord{
		ActorAdminID: actor.AdminID,
		Action:       domain.SellerStatusAuditAction(nextStatus),
		ResourceType: "seller",
		ResourceID:   sellerID,
		RequestID:    actor.RequestID,
		SessionID:    actor.SessionID,
		IPHash:       actor.IPHash,
		Reason:       reason,
		Before:       map[string]string{"status": string(seller.Status)},
		After:        map[string]string{"status": string(nextStatus)},
	}); err != nil {
		return domain.SuccessResponse{}, err
	}

	return domain.SuccessResponse{Success: true}, nil
}

func (s *ControlService) closeSellerKYCReviewTask(ctx context.Context, seller domain.SellerProfile, nextStatus domain.SellerStatus, actor domain.AdminActor, reason string) {
	if s.reviewTasks == nil || seller.Status != domain.SellerStatusPendingReview {
		return
	}

	reviewStatus, ok := domain.SellerKYCReviewStatusForSellerStatus(nextStatus)
	if !ok {
		return
	}

	if err := s.reviewTasks.CloseForResource(ctx, domain.CloseReviewTaskRequest{
		TaskType:     "seller_kyc",
		ResourceType: "seller",
		ResourceID:   seller.SellerID,
		Status:       reviewStatus,
		ReviewedBy:   actor.AdminID,
		Reason:       reason,
	}); err != nil {
		s.logger.Warn(ctx, "seller kyc review task close failed",
			"seller_id", seller.SellerID,
			"request_id", actor.RequestID,
			"error", err,
		)
	}
}

func (s *ControlService) recordAudit(ctx context.Context, record domain.AuditRecord) error {
	if err := s.audit.RecordAdminMutation(ctx, record); err != nil {
		s.logger.Warn(ctx, "admin mutation audit record failed",
			"action", record.Action,
			"resource_type", record.ResourceType,
			"resource_id", record.ResourceID,
			"request_id", record.RequestID,
			"error", err,
		)
		return auditRecordFailure(err)
	}
	return nil
}

func actorFromContext(ctx context.Context) (domain.AdminActor, error) {
	actor, ok := domain.ActorFromContext(ctx)
	if !ok {
		return domain.AdminActor{}, domain.NewAdminContextMissing("admin actor is missing from request context")
	}
	return actor, nil
}
