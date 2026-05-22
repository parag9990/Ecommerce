package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
)

func TestServiceCreateUserUsesGeneratedID(t *testing.T) {
	users := &fakeUserRepository{}
	service := newTestService(t, users, &fakeSellerRepository{})

	got, err := service.CreateUser(context.Background(), CreateUserInput{
		AuthAccountID: "auth_123",
		Email:         "buyer@example.com",
		Phone:         "+919999999999",
		FullName:      "Aarav Sharma",
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
	if got.UserID != "user_fixed" {
		t.Fatalf("UserID = %q, want user_fixed", got.UserID)
	}
	if users.created.UserID != "user_fixed" {
		t.Fatalf("repository saw UserID = %q", users.created.UserID)
	}
}

func TestServiceUpdateUserProfileRequiresMask(t *testing.T) {
	service := newTestService(t, &fakeUserRepository{byID: validUser()}, &fakeSellerRepository{})

	_, err := service.UpdateUserProfile(context.Background(), UpdateUserProfileInput{
		UserID:   "user_123",
		FullName: "Aarav S.",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestServiceUpdateUserProfileClearsAvatarWithFieldMask(t *testing.T) {
	users := &fakeUserRepository{byID: validUser()}
	service := newTestService(t, users, &fakeSellerRepository{})

	_, err := service.UpdateUserProfile(context.Background(), UpdateUserProfileInput{
		UserID:    "user_123",
		AvatarURL: "",
		Mask:      []string{"avatar_url"},
		Caller:    Caller{UserID: "user_123"},
	})
	if err != nil {
		t.Fatalf("UpdateUserProfile returned error: %v", err)
	}
	if users.updatePatch.AvatarURL == nil {
		t.Fatal("expected avatar_url to be included in repository patch")
	}
	if *users.updatePatch.AvatarURL != "" {
		t.Fatalf("AvatarURL patch = %q, want empty string clear marker", *users.updatePatch.AvatarURL)
	}
}

func TestServiceGetUserRejectsDifferentActor(t *testing.T) {
	service := newTestService(t, &fakeUserRepository{byID: validUser()}, &fakeSellerRepository{})

	_, err := service.GetUser(context.Background(), GetUserInput{
		UserID: "user_123",
		Caller: Caller{
			UserID: "user_other",
		},
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestServiceGetSellerProfileRejectsMismatchedUserID(t *testing.T) {
	service := newTestService(t, &fakeUserRepository{}, &fakeSellerRepository{
		bySellerID: validSeller(),
	})

	_, err := service.GetSellerProfile(context.Background(), GetSellerProfileInput{
		SellerID: "seller_123",
		UserID:   "user_other",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func newTestService(t *testing.T, users UserRepository, sellers SellerRepository) *Service {
	t.Helper()

	service, err := NewService(
		users,
		sellers,
		WithClock(fixedClock{at: fixedUsecaseTime()}),
		WithIDGenerator(fixedIDGenerator{id: "user_fixed"}),
	)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	return service
}

func validUser() domain.User {
	avatar := "https://cdn.example.com/avatar.png"
	return domain.User{
		UserID:        "user_123",
		AuthAccountID: "auth_123",
		Email:         "buyer@example.com",
		FullName:      "Aarav Sharma",
		AvatarURL:     &avatar,
		Status:        domain.UserStatusActive,
		CreatedAt:     fixedUsecaseTime(),
		UpdatedAt:     fixedUsecaseTime(),
	}
}

func validSeller() domain.SellerProfile {
	return domain.SellerProfile{
		SellerID:  "seller_123",
		UserID:    "user_123",
		StoreName: "Aarav Store",
		Status:    domain.SellerStatusDraft,
		CreatedAt: fixedUsecaseTime(),
		UpdatedAt: fixedUsecaseTime(),
	}
}

func fixedUsecaseTime() time.Time {
	return time.Date(2026, 5, 21, 10, 30, 0, 0, time.UTC)
}

type fixedClock struct {
	at time.Time
}

func (c fixedClock) Now() time.Time {
	return c.at
}

type fixedIDGenerator struct {
	id string
}

func (g fixedIDGenerator) NewUserID(context.Context) (string, error) {
	return g.id, nil
}

type fakeUserRepository struct {
	created     domain.User
	byID        domain.User
	createErr   error
	findErr     error
	updateErr   error
	updatePatch domain.UserProfilePatch
}

func (r *fakeUserRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	r.created = user
	if r.createErr != nil {
		return domain.User{}, r.createErr
	}
	return user, nil
}

func (r *fakeUserRepository) FindUserByID(ctx context.Context, userID string) (domain.User, error) {
	if r.findErr != nil {
		return domain.User{}, r.findErr
	}
	return r.byID, nil
}

func (r *fakeUserRepository) FindUserByAuthAccountID(ctx context.Context, authAccountID string) (domain.User, error) {
	return domain.User{}, domain.ErrUserNotFound
}

func (r *fakeUserRepository) BatchFindUsers(ctx context.Context, userIDs []string) ([]domain.User, error) {
	return nil, nil
}

func (r *fakeUserRepository) UpdateUserProfile(ctx context.Context, userID string, patch domain.UserProfilePatch) (domain.User, error) {
	r.updatePatch = patch
	if r.updateErr != nil {
		return domain.User{}, r.updateErr
	}
	updated := r.byID
	if patch.FullName != nil {
		updated.FullName = *patch.FullName
	}
	if patch.Phone != nil {
		updated.Phone = patch.Phone
	}
	if patch.AvatarURL != nil {
		updated.AvatarURL = patch.AvatarURL
	}
	return updated, nil
}

type fakeSellerRepository struct {
	bySellerID domain.SellerProfile
	byUserID   domain.SellerProfile
	err        error
}

func (r *fakeSellerRepository) CreateSellerProfile(ctx context.Context, seller domain.SellerProfile) (domain.SellerProfile, error) {
	return seller, nil
}

func (r *fakeSellerRepository) GetSellerProfileByUserID(ctx context.Context, userID string) (domain.SellerProfile, error) {
	if r.err != nil {
		return domain.SellerProfile{}, r.err
	}
	return r.byUserID, nil
}

func (r *fakeSellerRepository) GetSellerProfileBySellerID(ctx context.Context, sellerID string) (domain.SellerProfile, error) {
	if r.err != nil {
		return domain.SellerProfile{}, r.err
	}
	return r.bySellerID, nil
}

func (r *fakeSellerRepository) UpdateSellerProfile(ctx context.Context, sellerID string, patch domain.SellerProfilePatch) (domain.SellerProfile, error) {
	return domain.SellerProfile{}, nil
}

func (r *fakeSellerRepository) AddKYCDocument(ctx context.Context, document domain.KYCDocument) (domain.KYCDocument, error) {
	return document, nil
}

func (r *fakeSellerRepository) ListKYCDocuments(ctx context.Context, sellerID string) ([]domain.KYCDocument, error) {
	return nil, nil
}
