package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
	"github.com/parag/ecommerce/backend/services/user-service/internal/usecase"
	userv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestCreateUserMissingAuthAccountID(t *testing.T) {
	server := NewServer(&fakeUserUsecase{})

	_, err := server.CreateUser(context.Background(), &userv1.CreateUserRequest{
		Email:    "buyer@example.com",
		FullName: "Aarav Sharma",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("status.Code() = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
}

func TestGetUserMapsNotFound(t *testing.T) {
	server := NewServer(&fakeUserUsecase{getUserErr: domain.ErrUserNotFound})

	_, err := server.GetUser(context.Background(), &userv1.GetUserRequest{UserId: "user_missing"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("status.Code() = %v, want %v", status.Code(err), codes.NotFound)
	}
}

func TestUpdateUserProfileMapsMaskAndCallerMetadata(t *testing.T) {
	fake := &fakeUserUsecase{user: grpcValidUser()}
	server := NewServer(fake)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-user-id", "user_123",
		"x-service-name", "api-gateway",
		"x-roles", "buyer,seller",
	))

	got, err := server.UpdateUserProfile(ctx, &userv1.UpdateUserProfileRequest{
		UserId:   "user_123",
		FullName: "Aarav S.",
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{
			"full_name",
		}},
	})
	if err != nil {
		t.Fatalf("UpdateUserProfile returned error: %v", err)
	}
	if got.GetFullName() != "Aarav Sharma" {
		t.Fatalf("FullName = %q, want mapped domain user", got.GetFullName())
	}
	if fake.updateInput.Caller.UserID != "user_123" {
		t.Fatalf("caller user id = %q", fake.updateInput.Caller.UserID)
	}
	if len(fake.updateInput.Mask) != 1 || fake.updateInput.Mask[0] != "full_name" {
		t.Fatalf("mask = %#v", fake.updateInput.Mask)
	}
}

func TestGetSellerProfileRequiresID(t *testing.T) {
	server := NewServer(&fakeUserUsecase{})

	_, err := server.GetSellerProfile(context.Background(), &userv1.GetSellerProfileRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("status.Code() = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
}

func TestMapValidationError(t *testing.T) {
	err := domain.ValidationError{Fields: []domain.FieldError{{Field: "user_id", Message: "is required"}}}

	if status.Code(mapError(err)) != codes.InvalidArgument {
		t.Fatalf("status.Code() = %v, want %v", status.Code(mapError(err)), codes.InvalidArgument)
	}
}

type fakeUserUsecase struct {
	user        domain.User
	seller      domain.SellerProfile
	getUserErr  error
	updateInput usecase.UpdateUserProfileInput
}

func (u *fakeUserUsecase) CreateUser(ctx context.Context, input usecase.CreateUserInput) (domain.User, error) {
	if u.user.UserID == "" {
		return grpcValidUser(), nil
	}
	return u.user, nil
}

func (u *fakeUserUsecase) GetUser(ctx context.Context, input usecase.GetUserInput) (domain.User, error) {
	if u.getUserErr != nil {
		return domain.User{}, u.getUserErr
	}
	return grpcValidUser(), nil
}

func (u *fakeUserUsecase) UpdateUserProfile(ctx context.Context, input usecase.UpdateUserProfileInput) (domain.User, error) {
	u.updateInput = input
	if u.user.UserID == "" {
		return grpcValidUser(), nil
	}
	return u.user, nil
}

func (u *fakeUserUsecase) GetSellerProfile(ctx context.Context, input usecase.GetSellerProfileInput) (domain.SellerProfile, error) {
	if u.seller.SellerID == "" {
		return grpcValidSeller(), nil
	}
	return u.seller, nil
}

func grpcValidUser() domain.User {
	return domain.User{
		UserID:        "user_123",
		AuthAccountID: "auth_123",
		Email:         "buyer@example.com",
		FullName:      "Aarav Sharma",
		Status:        domain.UserStatusActive,
		CreatedAt:     fixedGRPCTime(),
		UpdatedAt:     fixedGRPCTime(),
	}
}

func grpcValidSeller() domain.SellerProfile {
	return domain.SellerProfile{
		SellerID:  "seller_123",
		UserID:    "user_123",
		StoreName: "Aarav Store",
		Status:    domain.SellerStatusDraft,
		CreatedAt: fixedGRPCTime(),
		UpdatedAt: fixedGRPCTime(),
	}
}

func fixedGRPCTime() time.Time {
	return time.Date(2026, 5, 21, 10, 30, 0, 0, time.UTC)
}
