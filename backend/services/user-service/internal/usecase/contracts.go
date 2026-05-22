package usecase

import (
	"context"

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	FindUserByID(ctx context.Context, userID string) (domain.User, error)
	FindUserByAuthAccountID(ctx context.Context, authAccountID string) (domain.User, error)
	BatchFindUsers(ctx context.Context, userIDs []string) ([]domain.User, error)
	UpdateUserProfile(ctx context.Context, userID string, patch domain.UserProfilePatch) (domain.User, error)
}

type AddressRepository interface {
	ListAddresses(ctx context.Context, userID string, limit int, offset int) ([]domain.Address, error)
	CreateAddress(ctx context.Context, address domain.Address) (domain.Address, error)
	UpdateAddress(ctx context.Context, address domain.Address) (domain.Address, error)
	DeleteAddress(ctx context.Context, userID string, addressID string) error
	SetDefaultAddress(ctx context.Context, userID string, addressID string) error
}

type SellerRepository interface {
	CreateSellerProfile(ctx context.Context, seller domain.SellerProfile) (domain.SellerProfile, error)
	GetSellerProfileByUserID(ctx context.Context, userID string) (domain.SellerProfile, error)
	GetSellerProfileBySellerID(ctx context.Context, sellerID string) (domain.SellerProfile, error)
	UpdateSellerProfile(ctx context.Context, sellerID string, patch domain.SellerProfilePatch) (domain.SellerProfile, error)
	AddKYCDocument(ctx context.Context, document domain.KYCDocument) (domain.KYCDocument, error)
	ListKYCDocuments(ctx context.Context, sellerID string) ([]domain.KYCDocument, error)
}
