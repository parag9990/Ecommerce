package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
)

type noopEventRecorder struct{}

func (noopEventRecorder) RecordUserCreated(context.Context, domain.UserCreatedPayload) error {
	return nil
}

func (noopEventRecorder) RecordSellerApproved(context.Context, domain.SellerApprovedPayload) error {
	return nil
}

func (noopEventRecorder) RecordAddressUpdated(context.Context, domain.AddressUpdatedPayload) error {
	return nil
}

func (s *Service) withWriteRepositories(ctx context.Context, fn func(context.Context, TransactionRepositories) error) error {
	if s.unitOfWork != nil {
		return s.unitOfWork.WithinTx(ctx, func(ctx context.Context, repositories TransactionRepositories) error {
			if repositories.Events == nil {
				repositories.Events = s.events
			}
			return fn(ctx, repositories)
		})
	}
	return fn(ctx, TransactionRepositories{
		Users:     s.users,
		Addresses: s.addresses,
		Sellers:   s.sellers,
		Events:    s.events,
	})
}

func buildUserCreatedPayload(user domain.User) domain.UserCreatedPayload {
	return domain.UserCreatedPayload{
		UserID:        user.UserID,
		AuthAccountID: user.AuthAccountID,
		Status:        string(user.Status),
		EmailHash:     hashForEvent(user.Email),
		PhoneHash:     hashOptionalForEvent(user.Phone),
		CreatedBy:     user.CreatedBy,
		CreatedAt:     user.CreatedAt,
	}
}

func buildSellerApprovedPayload(previous domain.SellerProfile, current domain.SellerProfile, reason string) domain.SellerApprovedPayload {
	approvedBy := ""
	if current.ApprovedBy != nil {
		approvedBy = *current.ApprovedBy
	}

	approvedAt := current.UpdatedAt
	if current.ApprovedAt != nil {
		approvedAt = *current.ApprovedAt
	}

	return domain.SellerApprovedPayload{
		SellerID:       current.SellerID,
		UserID:         current.UserID,
		StoreName:      current.StoreName,
		PreviousStatus: string(previous.Status),
		CurrentStatus:  string(current.Status),
		ApprovedBy:     approvedBy,
		ApprovedAt:     approvedAt,
		StatusReason:   strings.TrimSpace(reason),
	}
}

func buildAddressUpdatedPayload(address domain.Address, changeType domain.AddressChangeType, actorID string, changedAt time.Time) domain.AddressUpdatedPayload {
	return domain.AddressUpdatedPayload{
		AddressID:  address.AddressID,
		UserID:     address.UserID,
		ChangeType: string(changeType),
		City:       address.City,
		State:      address.State,
		Country:    address.Country,
		IsDefault:  address.IsDefault,
		UpdatedBy:  actorID,
		UpdatedAt:  changedAt,
	}
}

func sellerApprovalTriggersEvent(previous domain.SellerStatus, current domain.SellerStatus) bool {
	return current == domain.SellerStatusActive &&
		(previous == domain.SellerStatusPendingReview || previous == domain.SellerStatusDraft)
}

func hashOptionalForEvent(value *string) string {
	if value == nil {
		return ""
	}
	return hashForEvent(*value)
}

func hashForEvent(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}
