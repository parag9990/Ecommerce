package usecase_test

import (
	"context"
	"testing"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/usecase"
)

func TestUpdateUserStatusBlocksActiveUser(t *testing.T) {
	userClient := &controlUserClient{
		users: map[string]domain.UserProfile{
			"user_1": {UserID: "user_1", Status: domain.UserStatusActive},
		},
	}
	audit := &recordingAudit{}
	service := newControlService(t, userClient, nil, audit)

	resp, err := service.UpdateUserStatus(actorContext(), "user_1", domain.StatusUpdateRequest{
		Status: string(domain.UserStatusBlocked),
		Reason: "confirmed account takeover risk",
	})
	if err != nil {
		t.Fatalf("UpdateUserStatus returned error: %v", err)
	}
	if !resp.Success {
		t.Fatal("success = false, want true")
	}
	if userClient.updatedUserStatus != domain.UserStatusBlocked {
		t.Fatalf("updated status = %s, want blocked", userClient.updatedUserStatus)
	}
	if len(audit.records) != 1 || audit.records[0].Action != "user.status.blocked" {
		t.Fatalf("audit records = %+v", audit.records)
	}
}

func TestUpdateUserStatusRejectsInvalidTransition(t *testing.T) {
	userClient := &controlUserClient{
		users: map[string]domain.UserProfile{
			"user_1": {UserID: "user_1", Status: domain.UserStatusDeleted},
		},
	}
	service := newControlService(t, userClient, nil, &recordingAudit{})

	_, err := service.UpdateUserStatus(actorContext(), "user_1", domain.StatusUpdateRequest{
		Status: string(domain.UserStatusActive),
		Reason: "restore requested by support team",
	})
	if appErr, ok := domain.AsAppError(err); !ok || appErr.Code != domain.CodeInvalidStatusTransition {
		t.Fatalf("error = %v, want INVALID_STATUS_TRANSITION", err)
	}
	if userClient.updatedUserStatus != "" {
		t.Fatalf("unexpected downstream update: %s", userClient.updatedUserStatus)
	}
}

func TestUpdateUserStatusRequiresPermission(t *testing.T) {
	service, err := usecase.NewControlService(
		&controlAuthorizer{deny: domain.PermissionUsersStatusUpdate},
		&controlUserClient{},
		nil,
		&recordingAudit{},
		logging.NewNop(),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.UpdateUserStatus(actorContext(), "user_1", domain.StatusUpdateRequest{
		Status: string(domain.UserStatusBlocked),
		Reason: "confirmed account takeover risk",
	})
	if appErr, ok := domain.AsAppError(err); !ok || appErr.Code != domain.CodeForbidden {
		t.Fatalf("error = %v, want FORBIDDEN", err)
	}
}

func TestUpdateUserStatusReturnsAuditFailure(t *testing.T) {
	userClient := &controlUserClient{
		users: map[string]domain.UserProfile{
			"user_1": {UserID: "user_1", Status: domain.UserStatusActive},
		},
	}
	audit := &recordingAudit{err: domain.NewInternal("audit unavailable", nil)}
	service := newControlService(t, userClient, nil, audit)

	_, err := service.UpdateUserStatus(actorContext(), "user_1", domain.StatusUpdateRequest{
		Status: string(domain.UserStatusBlocked),
		Reason: "confirmed account takeover risk",
	})
	if appErr, ok := domain.AsAppError(err); !ok || appErr.Code != domain.CodeInternal {
		t.Fatalf("error = %v, want INTERNAL", err)
	}
}

func TestUpdateSellerStatusApprovesPendingReviewAndClosesTask(t *testing.T) {
	userClient := &controlUserClient{
		sellers: map[string]domain.SellerProfile{
			"seller_1": {SellerID: "seller_1", Status: domain.SellerStatusPendingReview},
		},
	}
	reviewTasks := &recordingReviewTasks{}
	audit := &recordingAudit{}
	service := newControlService(t, userClient, reviewTasks, audit)

	resp, err := service.UpdateSellerStatus(actorContext(), "seller_1", domain.StatusUpdateRequest{
		Status: string(domain.SellerStatusActive),
		Reason: "gst and bank verification passed",
	})
	if err != nil {
		t.Fatalf("UpdateSellerStatus returned error: %v", err)
	}
	if !resp.Success {
		t.Fatal("success = false, want true")
	}
	if userClient.updatedSellerStatus != domain.SellerStatusActive {
		t.Fatalf("updated status = %s, want active", userClient.updatedSellerStatus)
	}
	if len(reviewTasks.requests) != 1 || reviewTasks.requests[0].Status != domain.ReviewTaskStatusApproved {
		t.Fatalf("review close requests = %+v", reviewTasks.requests)
	}
	if len(audit.records) != 1 || audit.records[0].Action != "seller.status.active" {
		t.Fatalf("audit records = %+v", audit.records)
	}
}

func TestUpdateSellerStatusRejectsDraftApproval(t *testing.T) {
	userClient := &controlUserClient{
		sellers: map[string]domain.SellerProfile{
			"seller_1": {SellerID: "seller_1", Status: domain.SellerStatusDraft},
		},
	}
	reviewTasks := &recordingReviewTasks{}
	service := newControlService(t, userClient, reviewTasks, &recordingAudit{})

	_, err := service.UpdateSellerStatus(actorContext(), "seller_1", domain.StatusUpdateRequest{
		Status: string(domain.SellerStatusActive),
		Reason: "documents are not submitted yet",
	})
	if appErr, ok := domain.AsAppError(err); !ok || appErr.Code != domain.CodeInvalidStatusTransition {
		t.Fatalf("error = %v, want INVALID_STATUS_TRANSITION", err)
	}
	if len(reviewTasks.requests) != 0 {
		t.Fatalf("unexpected review task close: %+v", reviewTasks.requests)
	}
}

func newControlService(t *testing.T, userClient *controlUserClient, reviewTasks usecase.ReviewTaskRepository, audit usecase.AuditRecorder) *usecase.ControlService {
	t.Helper()
	service, err := usecase.NewControlService(&controlAuthorizer{}, userClient, reviewTasks, audit, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func actorContext() context.Context {
	actor := domain.AdminActor{
		AdminID:   "admin_1",
		UserID:    "user_admin_1",
		Roles:     []domain.AdminRole{domain.RoleSuperadmin},
		SessionID: "sess_1",
		RequestID: "req_1",
		IPHash:    "hash_1",
	}
	return domain.ContextWithActor(context.Background(), actor)
}

type controlAuthorizer struct {
	deny domain.Permission
}

func (a *controlAuthorizer) RequirePermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission) error {
	if a.deny == permission {
		return domain.NewForbidden(permission)
	}
	return actor.ValidateForAdminRoute()
}

func (a *controlAuthorizer) RequireHighRiskPermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission, reason string) error {
	if err := a.RequirePermission(ctx, actor, permission); err != nil {
		return err
	}
	if reason == "" {
		return domain.NewReasonRequired()
	}
	if actor.RequestID == "" {
		return domain.NewRequestContextMissing()
	}
	return nil
}

type controlUserClient struct {
	users               map[string]domain.UserProfile
	sellers             map[string]domain.SellerProfile
	updatedUserStatus   domain.UserStatus
	updatedSellerStatus domain.SellerStatus
}

func (c *controlUserClient) ListUsersForAdmin(ctx context.Context, req domain.AdminUserListRequest, actor domain.AdminActor) (domain.AdminUserListResponse, error) {
	users := make([]domain.UserProfile, 0, len(c.users))
	for _, user := range c.users {
		users = append(users, user)
	}
	return domain.AdminUserListResponse{Users: users}, nil
}

func (c *controlUserClient) GetUserForAdmin(ctx context.Context, userID string, actor domain.AdminActor) (domain.UserProfile, error) {
	user, ok := c.users[userID]
	if !ok {
		return domain.UserProfile{}, domain.NewUserNotFound(userID)
	}
	return user, nil
}

func (c *controlUserClient) UpdateUserStatus(ctx context.Context, userID string, status domain.UserStatus, mutation domain.AdminMutationContext) error {
	c.updatedUserStatus = status
	return nil
}

func (c *controlUserClient) ListSellersForAdmin(ctx context.Context, req domain.AdminSellerListRequest, actor domain.AdminActor) (domain.AdminSellerListResponse, error) {
	sellers := make([]domain.SellerProfile, 0, len(c.sellers))
	for _, seller := range c.sellers {
		sellers = append(sellers, seller)
	}
	return domain.AdminSellerListResponse{Sellers: sellers}, nil
}

func (c *controlUserClient) GetSellerForAdmin(ctx context.Context, sellerID string, actor domain.AdminActor) (domain.SellerProfile, error) {
	seller, ok := c.sellers[sellerID]
	if !ok {
		return domain.SellerProfile{}, domain.NewSellerNotFound(sellerID)
	}
	return seller, nil
}

func (c *controlUserClient) UpdateSellerStatus(ctx context.Context, sellerID string, status domain.SellerStatus, mutation domain.AdminMutationContext) error {
	c.updatedSellerStatus = status
	return nil
}

type recordingReviewTasks struct {
	requests []domain.CloseReviewTaskRequest
}

func (r *recordingReviewTasks) CloseForResource(ctx context.Context, req domain.CloseReviewTaskRequest) error {
	r.requests = append(r.requests, req)
	return nil
}

type recordingAudit struct {
	records []domain.AuditRecord
	err     error
}

func (r *recordingAudit) RecordAdminMutation(ctx context.Context, record domain.AuditRecord) error {
	r.records = append(r.records, record)
	return r.err
}
