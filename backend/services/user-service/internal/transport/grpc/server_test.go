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

func TestCreateAddressMapsRequestAndCallerMetadata(t *testing.T) {
	fake := &fakeUserUsecase{address: grpcValidAddress()}
	server := NewServer(fake)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-user-id", "user_123",
		"x-service-name", "api-gateway",
		"x-roles", "buyer",
	))

	got, err := server.CreateAddress(ctx, &userv1.CreateAddressRequest{
		UserId: "user_123",
		Address: &userv1.AddressInput{
			Name:       "Aarav Sharma",
			Line1:      "221B MG Road",
			City:       "Bengaluru",
			State:      "Karnataka",
			PostalCode: "560001",
			Country:    "IN",
			IsDefault:  true,
		},
	})
	if err != nil {
		t.Fatalf("CreateAddress returned error: %v", err)
	}
	if got.GetAddressId() != "addr_123" {
		t.Fatalf("AddressId = %q, want addr_123", got.GetAddressId())
	}
	if fake.createAddressInput.Caller.UserID != "user_123" {
		t.Fatalf("caller user id = %q", fake.createAddressInput.Caller.UserID)
	}
	if !fake.createAddressInput.IsDefault {
		t.Fatal("expected is_default to be mapped")
	}
}

func TestDeleteAddressMapsNotFound(t *testing.T) {
	server := NewServer(&fakeUserUsecase{deleteAddressErr: domain.ErrAddressNotFound})

	_, err := server.DeleteAddress(context.Background(), &userv1.DeleteAddressRequest{
		UserId:    "user_123",
		AddressId: "addr_missing",
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("status.Code() = %v, want %v", status.Code(err), codes.NotFound)
	}
}

func TestUpdateSellerProfileMapsMask(t *testing.T) {
	fake := &fakeUserUsecase{seller: grpcValidSeller()}
	server := NewServer(fake)

	_, err := server.UpdateSellerProfile(context.Background(), &userv1.UpdateSellerProfileRequest{
		SellerId:  "seller_123",
		UserId:    "user_123",
		StoreName: "Updated Store",
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{
			"store_name",
		}},
	})
	if err != nil {
		t.Fatalf("UpdateSellerProfile returned error: %v", err)
	}
	if len(fake.updateSellerInput.Mask) != 1 || fake.updateSellerInput.Mask[0] != "store_name" {
		t.Fatalf("mask = %#v", fake.updateSellerInput.Mask)
	}
}

func TestMapValidationError(t *testing.T) {
	err := domain.ValidationError{Fields: []domain.FieldError{{Field: "user_id", Message: "is required"}}}

	if status.Code(mapError(err)) != codes.InvalidArgument {
		t.Fatalf("status.Code() = %v, want %v", status.Code(mapError(err)), codes.InvalidArgument)
	}
}

type fakeUserUsecase struct {
	user                    domain.User
	seller                  domain.SellerProfile
	address                 domain.Address
	getUserErr              error
	deleteAddressErr        error
	updateInput             usecase.UpdateUserProfileInput
	createAddressInput      usecase.CreateAddressInput
	updateSellerInput       usecase.UpdateSellerProfileInput
	updateUserStatusInput   usecase.UpdateUserStatusInput
	updateSellerStatusInput usecase.UpdateSellerStatusInput
	reviewKYCInput          usecase.ReviewKYCDocumentInput
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

func (u *fakeUserUsecase) ListUserAddresses(ctx context.Context, input usecase.ListUserAddressesInput) ([]domain.Address, error) {
	if u.address.AddressID == "" {
		return []domain.Address{grpcValidAddress()}, nil
	}
	return []domain.Address{u.address}, nil
}

func (u *fakeUserUsecase) CreateAddress(ctx context.Context, input usecase.CreateAddressInput) (domain.Address, error) {
	u.createAddressInput = input
	if u.address.AddressID == "" {
		return grpcValidAddress(), nil
	}
	return u.address, nil
}

func (u *fakeUserUsecase) UpdateAddress(ctx context.Context, input usecase.UpdateAddressInput) (domain.Address, error) {
	if u.address.AddressID == "" {
		return grpcValidAddress(), nil
	}
	return u.address, nil
}

func (u *fakeUserUsecase) DeleteAddress(ctx context.Context, input usecase.DeleteAddressInput) error {
	if u.deleteAddressErr != nil {
		return u.deleteAddressErr
	}
	return nil
}

func (u *fakeUserUsecase) GetSellerProfile(ctx context.Context, input usecase.GetSellerProfileInput) (domain.SellerProfile, error) {
	if u.seller.SellerID == "" {
		return grpcValidSeller(), nil
	}
	return u.seller, nil
}

func (u *fakeUserUsecase) UpdateSellerProfile(ctx context.Context, input usecase.UpdateSellerProfileInput) (domain.SellerProfile, error) {
	u.updateSellerInput = input
	if u.seller.SellerID == "" {
		return grpcValidSeller(), nil
	}
	return u.seller, nil
}

func (u *fakeUserUsecase) UpdateUserStatus(ctx context.Context, input usecase.UpdateUserStatusInput) (domain.User, error) {
	u.updateUserStatusInput = input
	if u.user.UserID == "" {
		return grpcValidUser(), nil
	}
	return u.user, nil
}

func (u *fakeUserUsecase) UpdateSellerStatus(ctx context.Context, input usecase.UpdateSellerStatusInput) (domain.SellerProfile, error) {
	u.updateSellerStatusInput = input
	if u.seller.SellerID == "" {
		return grpcValidSeller(), nil
	}
	return u.seller, nil
}

func (u *fakeUserUsecase) ReviewKYCDocument(ctx context.Context, input usecase.ReviewKYCDocumentInput) (domain.KYCDocument, error) {
	u.reviewKYCInput = input
	return grpcValidKYCDocument(), nil
}

func grpcValidUser() domain.User {
	return domain.User{
		UserID:        "user_123",
		AuthAccountID: "auth_123",
		Email:         "buyer@example.com",
		FullName:      "Aarav Sharma",
		Status:        domain.UserStatusActive,
		AuditFields: domain.AuditFields{
			CreatedBy: "service:auth-service",
			UpdatedBy: "service:auth-service",
			CreatedAt: fixedGRPCTime(),
			UpdatedAt: fixedGRPCTime(),
		},
		StatusAuditFields: domain.StatusAuditFields{
			StatusChangedBy: grpcStringPtr("service:auth-service"),
			StatusChangedAt: grpcTimePtr(fixedGRPCTime()),
		},
	}
}

func grpcValidSeller() domain.SellerProfile {
	return domain.SellerProfile{
		SellerID:  "seller_123",
		UserID:    "user_123",
		StoreName: "Aarav Store",
		Status:    domain.SellerStatusDraft,
		AuditFields: domain.AuditFields{
			CreatedBy: "user_123",
			UpdatedBy: "user_123",
			CreatedAt: fixedGRPCTime(),
			UpdatedAt: fixedGRPCTime(),
		},
		StatusAuditFields: domain.StatusAuditFields{
			StatusChangedBy: grpcStringPtr("user_123"),
			StatusChangedAt: grpcTimePtr(fixedGRPCTime()),
		},
	}
}

func grpcValidAddress() domain.Address {
	return domain.Address{
		AddressID:  "addr_123",
		UserID:     "user_123",
		Name:       "Aarav Sharma",
		Line1:      "221B MG Road",
		City:       "Bengaluru",
		State:      "Karnataka",
		PostalCode: "560001",
		Country:    "IN",
		IsDefault:  true,
		Status:     domain.AddressStatusActive,
		AuditFields: domain.AuditFields{
			CreatedBy: "user_123",
			UpdatedBy: "user_123",
			CreatedAt: fixedGRPCTime(),
			UpdatedAt: fixedGRPCTime(),
		},
	}
}

func grpcValidKYCDocument() domain.KYCDocument {
	return domain.KYCDocument{
		DocumentID:   "doc_123",
		SellerID:     "seller_123",
		DocumentType: domain.KYCDocumentTypePANCard,
		StorageURL:   "s3://private-kyc/seller_123/doc_123.pdf",
		Status:       domain.KYCStatusApproved,
		ReviewedBy:   grpcStringPtr("admin_123"),
		ReviewedAt:   grpcTimePtr(fixedGRPCTime()),
		AuditFields: domain.AuditFields{
			CreatedBy: "user_123",
			UpdatedBy: "admin_123",
			CreatedAt: fixedGRPCTime().Add(-time.Hour),
			UpdatedAt: fixedGRPCTime(),
		},
		StatusAuditFields: domain.StatusAuditFields{
			StatusChangedBy: grpcStringPtr("admin_123"),
			StatusChangedAt: grpcTimePtr(fixedGRPCTime()),
		},
	}
}

func grpcStringPtr(value string) *string {
	return &value
}

func grpcTimePtr(value time.Time) *time.Time {
	return &value
}

func fixedGRPCTime() time.Time {
	return time.Date(2026, 5, 21, 10, 30, 0, 0, time.UTC)
}
