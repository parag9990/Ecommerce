package usecase

import (
	"context"

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
	uservalidation "github.com/parag/ecommerce/backend/services/user-service/internal/validation"
	sharedvalidation "github.com/parag/ecommerce/backend/shared/validation"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	FindUserByID(ctx context.Context, userID string) (domain.User, error)
	FindUserByAuthAccountID(ctx context.Context, authAccountID string) (domain.User, error)
	BatchFindUsers(ctx context.Context, userIDs []string) ([]domain.User, error)
	UpdateUserProfile(ctx context.Context, userID string, patch domain.UserProfilePatch) (domain.User, error)
	UpdateUserStatus(ctx context.Context, user domain.User) (domain.User, error)
}

type AddressRepository interface {
	ListAddresses(ctx context.Context, userID string, limit int, offset int) ([]domain.Address, error)
	FindAddress(ctx context.Context, userID string, addressID string) (domain.Address, error)
	CreateAddress(ctx context.Context, address domain.Address) (domain.Address, error)
	UpdateAddress(ctx context.Context, address domain.Address) (domain.Address, error)
	DeleteAddress(ctx context.Context, userID string, addressID string, audit domain.MutationAudit) error
	SetDefaultAddress(ctx context.Context, userID string, addressID string, audit domain.MutationAudit) error
}

type SellerRepository interface {
	CreateSellerProfile(ctx context.Context, seller domain.SellerProfile) (domain.SellerProfile, error)
	GetSellerProfileByUserID(ctx context.Context, userID string) (domain.SellerProfile, error)
	GetSellerProfileBySellerID(ctx context.Context, sellerID string) (domain.SellerProfile, error)
	UpdateSellerProfile(ctx context.Context, sellerID string, patch domain.SellerProfilePatch) (domain.SellerProfile, error)
	UpdateSellerStatus(ctx context.Context, seller domain.SellerProfile) (domain.SellerProfile, error)
	AddKYCDocument(ctx context.Context, document domain.KYCDocument) (domain.KYCDocument, error)
	GetKYCDocument(ctx context.Context, sellerID string, documentID string) (domain.KYCDocument, error)
	ReviewKYCDocument(ctx context.Context, document domain.KYCDocument) (domain.KYCDocument, error)
	ListKYCDocuments(ctx context.Context, sellerID string) ([]domain.KYCDocument, error)
}

type EventRecorder interface {
	RecordUserCreated(ctx context.Context, payload domain.UserCreatedPayload) error
	RecordSellerApproved(ctx context.Context, payload domain.SellerApprovedPayload) error
	RecordAddressUpdated(ctx context.Context, payload domain.AddressUpdatedPayload) error
}

type TransactionRepositories struct {
	Users     UserRepository
	Addresses AddressRepository
	Sellers   SellerRepository
	Events    EventRecorder
}

type UnitOfWork interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, repositories TransactionRepositories) error) error
}

type InputValidator interface {
	NormalizeAndValidateCreateUser(input uservalidation.CreateUserInput) (uservalidation.CreateUserInput, sharedvalidation.Error)
	NormalizeAndValidateUpdateProfile(input uservalidation.UpdateProfileInput, fields uservalidation.ProfileUpdateFields) (uservalidation.UpdateProfileInput, sharedvalidation.Error)
	NormalizeAndValidateAddress(input uservalidation.AddressInput) (uservalidation.AddressInput, sharedvalidation.Error)
	NormalizeAndValidateUpdateSellerProfile(input uservalidation.UpdateSellerProfileInput, fields uservalidation.SellerUpdateFields) (uservalidation.UpdateSellerProfileInput, sharedvalidation.Error)
}
