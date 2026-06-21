package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/parag/ecommerce/backend/services/user-service/internal/audit"
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
		Caller:        Caller{ServiceName: "auth-service"},
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
	if users.created.CreatedBy != "service:auth-service" {
		t.Fatalf("CreatedBy = %q, want service:auth-service", users.created.CreatedBy)
	}
}

func TestServiceCreateUserRecordsUserCreatedEvent(t *testing.T) {
	events := &fakeEventRecorder{}
	service := newTestServiceWithOptions(t, &fakeUserRepository{}, &fakeAddressRepository{}, &fakeSellerRepository{}, WithEventRecorder(events))

	_, err := service.CreateUser(context.Background(), CreateUserInput{
		AuthAccountID: "auth_123",
		Email:         "Buyer@Example.COM",
		Phone:         "+919999999999",
		FullName:      "Aarav Sharma",
		Caller:        Caller{ServiceName: "auth-service"},
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
	if events.userCreated.UserID != "user_fixed" {
		t.Fatalf("UserCreated user_id = %q, want user_fixed", events.userCreated.UserID)
	}
	if events.userCreated.EmailHash == "" || events.userCreated.EmailHash == "buyer@example.com" {
		t.Fatalf("EmailHash should be populated without raw email, got %q", events.userCreated.EmailHash)
	}
	if events.userCreated.PhoneHash == "" || events.userCreated.PhoneHash == "+919999999999" {
		t.Fatalf("PhoneHash should be populated without raw phone, got %q", events.userCreated.PhoneHash)
	}
}

func TestServiceCreateUserRejectsInvalidEmailBeforeRepository(t *testing.T) {
	users := &fakeUserRepository{}
	service := newTestService(t, users, &fakeSellerRepository{})

	_, err := service.CreateUser(context.Background(), CreateUserInput{
		AuthAccountID: "auth_123",
		Email:         "not-an-email",
		FullName:      "Aarav Sharma",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if users.created.UserID != "" {
		t.Fatalf("repository should not be called, created = %#v", users.created)
	}
}

func TestServiceUpdateUserProfileRequiresMask(t *testing.T) {
	service := newTestService(t, &fakeUserRepository{byID: validUser()}, &fakeSellerRepository{})

	_, err := service.UpdateUserProfile(context.Background(), UpdateUserProfileInput{
		UserID:   "user_123",
		FullName: "Aarav S.",
		Caller:   Caller{UserID: "user_123"},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestServiceCreateUserRequiresAuditActor(t *testing.T) {
	users := &fakeUserRepository{}
	service := newTestService(t, users, &fakeSellerRepository{})

	_, err := service.CreateUser(context.Background(), CreateUserInput{
		AuthAccountID: "auth_123",
		Email:         "buyer@example.com",
		FullName:      "Aarav Sharma",
	})
	if err == nil {
		t.Fatal("expected missing actor error")
	}
	if users.created.UserID != "" {
		t.Fatalf("repository should not be called, created = %#v", users.created)
	}
}

func TestServiceCreateUserRejectsEndUserCaller(t *testing.T) {
	users := &fakeUserRepository{}
	service := newTestService(t, users, &fakeSellerRepository{})

	_, err := service.CreateUser(context.Background(), CreateUserInput{
		AuthAccountID: "auth_123",
		Email:         "buyer@example.com",
		FullName:      "Aarav Sharma",
		Caller:        Caller{UserID: "user_123"},
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
	if users.created.UserID != "" {
		t.Fatalf("repository should not be called, created = %#v", users.created)
	}
}

func TestServiceUpdateUserProfileRejectsInvalidAvatarBeforeRepository(t *testing.T) {
	findErr := errors.New("repository find should not be called")
	service := newTestService(t, &fakeUserRepository{byID: validUser(), findErr: findErr}, &fakeSellerRepository{})

	_, err := service.UpdateUserProfile(context.Background(), UpdateUserProfileInput{
		UserID:    "user_123",
		AvatarURL: "http://cdn.example.com/avatar.png",
		Mask:      []string{"avatar_url"},
		Caller:    Caller{UserID: "user_123"},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if errors.Is(err, findErr) {
		t.Fatal("repository find should not be called for invalid input")
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

func TestServiceUpdateUserProfileNormalizesPhone(t *testing.T) {
	users := &fakeUserRepository{byID: validUser()}
	service := newTestService(t, users, &fakeSellerRepository{})

	_, err := service.UpdateUserProfile(context.Background(), UpdateUserProfileInput{
		UserID: "user_123",
		Phone:  "99999 99999",
		Mask:   []string{"phone"},
		Caller: Caller{UserID: "user_123"},
	})
	if err != nil {
		t.Fatalf("UpdateUserProfile returned error: %v", err)
	}
	if users.updatePatch.Phone == nil || *users.updatePatch.Phone != "+919999999999" {
		t.Fatalf("Phone patch = %#v, want normalized phone", users.updatePatch.Phone)
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

func TestServiceGetUserRequiresCallerIdentity(t *testing.T) {
	service := newTestService(t, &fakeUserRepository{byID: validUser()}, &fakeSellerRepository{})

	_, err := service.GetUser(context.Background(), GetUserInput{UserID: "user_123"})
	if !errors.Is(err, audit.ErrMissingActor) {
		t.Fatalf("expected missing actor error, got %v", err)
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

func TestServiceCreateAddressFirstAddressBecomesDefault(t *testing.T) {
	addresses := &fakeAddressRepository{}
	service := newTestServiceWithAddresses(t, &fakeUserRepository{}, addresses, &fakeSellerRepository{})

	got, err := service.CreateAddress(context.Background(), CreateAddressInput{
		UserID:     "user_123",
		Name:       "Aarav Sharma",
		Line1:      "221B MG Road",
		City:       "Bengaluru",
		State:      "Karnataka",
		PostalCode: "560001",
		Country:    "IN",
		Caller:     Caller{UserID: "user_123"},
	})
	if err != nil {
		t.Fatalf("CreateAddress returned error: %v", err)
	}
	if got.AddressID != "addr_fixed" {
		t.Fatalf("AddressID = %q, want addr_fixed", got.AddressID)
	}
	if !addresses.created.IsDefault {
		t.Fatal("first address should be created as default")
	}
}

func TestServiceCreateAddressRecordsAddressUpdatedEvent(t *testing.T) {
	events := &fakeEventRecorder{}
	addresses := &fakeAddressRepository{}
	service := newTestServiceWithOptions(t, &fakeUserRepository{}, addresses, &fakeSellerRepository{}, WithEventRecorder(events))

	_, err := service.CreateAddress(context.Background(), CreateAddressInput{
		UserID:     "user_123",
		Name:       "Aarav Sharma",
		Line1:      "221B MG Road",
		City:       "Bengaluru",
		State:      "Karnataka",
		PostalCode: "560001",
		Country:    "IN",
		Caller:     Caller{UserID: "user_123"},
	})
	if err != nil {
		t.Fatalf("CreateAddress returned error: %v", err)
	}
	if events.addressUpdated.ChangeType != string(domain.AddressChangeCreated) {
		t.Fatalf("AddressUpdated change_type = %q, want created", events.addressUpdated.ChangeType)
	}
	if events.addressUpdated.City != "Bengaluru" || events.addressUpdated.Country != "IN" {
		t.Fatalf("AddressUpdated location = %#v", events.addressUpdated)
	}
}

func TestServiceCreateAddressRejectsInvalidPostalCodeBeforeRepository(t *testing.T) {
	addresses := &fakeAddressRepository{}
	service := newTestServiceWithAddresses(t, &fakeUserRepository{}, addresses, &fakeSellerRepository{})

	_, err := service.CreateAddress(context.Background(), CreateAddressInput{
		UserID:     "user_123",
		Name:       "Aarav Sharma",
		Line1:      "221B MG Road",
		City:       "Bengaluru",
		State:      "Karnataka",
		PostalCode: "012345",
		Country:    "IN",
		Caller:     Caller{UserID: "user_123"},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if addresses.listCalls != 0 || addresses.created.AddressID != "" {
		t.Fatalf("repository should not be called, listCalls=%d created=%#v", addresses.listCalls, addresses.created)
	}
}

func TestServiceCreateAddressRejectsLimitExceeded(t *testing.T) {
	addresses := &fakeAddressRepository{list: make([]domain.Address, maxUserAddresses)}
	service := newTestServiceWithAddresses(t, &fakeUserRepository{}, addresses, &fakeSellerRepository{})

	_, err := service.CreateAddress(context.Background(), CreateAddressInput{
		UserID:     "user_123",
		Name:       "Aarav Sharma",
		Line1:      "221B MG Road",
		City:       "Bengaluru",
		State:      "Karnataka",
		PostalCode: "560001",
		Country:    "IN",
		Caller:     Caller{UserID: "user_123"},
	})
	if !errors.Is(err, domain.ErrAddressLimitExceeded) {
		t.Fatalf("expected ErrAddressLimitExceeded, got %v", err)
	}
}

func TestServiceUpdateSellerProfileUsesFieldMask(t *testing.T) {
	sellers := &fakeSellerRepository{bySellerID: validSeller()}
	service := newTestService(t, &fakeUserRepository{}, sellers)

	_, err := service.UpdateSellerProfile(context.Background(), UpdateSellerProfileInput{
		SellerID:  "seller_123",
		UserID:    "user_123",
		StoreName: "Updated Store",
		Mask:      []string{"store_name"},
		Caller:    Caller{UserID: "user_123"},
	})
	if err != nil {
		t.Fatalf("UpdateSellerProfile returned error: %v", err)
	}
	if sellers.updatePatch.StoreName == nil || *sellers.updatePatch.StoreName != "Updated Store" {
		t.Fatalf("StoreName patch = %#v, want Updated Store", sellers.updatePatch.StoreName)
	}
}

func TestServiceUpdateSellerProfileRejectsInvalidGSTINBeforeWrite(t *testing.T) {
	sellers := &fakeSellerRepository{bySellerID: validSeller()}
	service := newTestService(t, &fakeUserRepository{}, sellers)

	_, err := service.UpdateSellerProfile(context.Background(), UpdateSellerProfileInput{
		SellerID:  "seller_123",
		UserID:    "user_123",
		GSTNumber: "99AAAAA0000A1Z5",
		Mask:      []string{"gst_number"},
		Caller:    Caller{UserID: "user_123"},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if sellers.updatePatch.GSTNumber != nil {
		t.Fatalf("repository update should not be called, patch = %#v", sellers.updatePatch)
	}
}

func TestServiceUpdateSellerProfileNormalizesEmailAndGSTIN(t *testing.T) {
	sellers := &fakeSellerRepository{bySellerID: validSeller()}
	service := newTestService(t, &fakeUserRepository{}, sellers)

	_, err := service.UpdateSellerProfile(context.Background(), UpdateSellerProfileInput{
		SellerID:     "seller_123",
		UserID:       "user_123",
		GSTNumber:    "22aaaaa0000a1z5",
		SupportEmail: " SUPPORT@EXAMPLE.COM ",
		Mask:         []string{"gst_number", "support_email"},
		Caller:       Caller{UserID: "user_123"},
	})
	if err != nil {
		t.Fatalf("UpdateSellerProfile returned error: %v", err)
	}
	if sellers.updatePatch.GSTNumber == nil || *sellers.updatePatch.GSTNumber != "22AAAAA0000A1Z5" {
		t.Fatalf("GSTNumber patch = %#v, want normalized GSTIN", sellers.updatePatch.GSTNumber)
	}
	if sellers.updatePatch.SupportEmail == nil || *sellers.updatePatch.SupportEmail != "support@example.com" {
		t.Fatalf("SupportEmail patch = %#v, want normalized email", sellers.updatePatch.SupportEmail)
	}
}

func TestServiceUpdateSellerStatusRecordsSellerApprovedEvent(t *testing.T) {
	events := &fakeEventRecorder{}
	seller := validSeller()
	seller.Status = domain.SellerStatusPendingReview
	sellers := &fakeSellerRepository{bySellerID: seller}
	service := newTestServiceWithOptions(t, &fakeUserRepository{}, &fakeAddressRepository{}, sellers, WithEventRecorder(events))

	_, err := service.UpdateSellerStatus(context.Background(), UpdateSellerStatusInput{
		SellerID: "seller_123",
		Status:   string(domain.SellerStatusActive),
		Reason:   "KYC verified",
		Caller:   Caller{ActorID: "admin_123", ActorType: "admin"},
	})
	if err != nil {
		t.Fatalf("UpdateSellerStatus returned error: %v", err)
	}
	if events.sellerApproved.SellerID != "seller_123" {
		t.Fatalf("SellerApproved seller_id = %q, want seller_123", events.sellerApproved.SellerID)
	}
	if events.sellerApproved.PreviousStatus != string(domain.SellerStatusPendingReview) || events.sellerApproved.CurrentStatus != string(domain.SellerStatusActive) {
		t.Fatalf("SellerApproved statuses = %#v", events.sellerApproved)
	}
	if events.sellerApproved.ApprovedBy != "admin_123" {
		t.Fatalf("SellerApproved approved_by = %q, want admin_123", events.sellerApproved.ApprovedBy)
	}
}

func newTestService(t *testing.T, users UserRepository, sellers SellerRepository) *Service {
	return newTestServiceWithAddresses(t, users, &fakeAddressRepository{}, sellers)
}

func newTestServiceWithAddresses(t *testing.T, users UserRepository, addresses AddressRepository, sellers SellerRepository) *Service {
	return newTestServiceWithOptions(t, users, addresses, sellers)
}

func newTestServiceWithOptions(t *testing.T, users UserRepository, addresses AddressRepository, sellers SellerRepository, options ...Option) *Service {
	t.Helper()

	baseOptions := []Option{
		WithClock(fixedClock{at: fixedUsecaseTime()}),
		WithIDGenerator(fixedIDGenerator{id: "user_fixed"}),
		WithAddressIDGenerator(fixedAddressIDGenerator{id: "addr_fixed"}),
	}
	baseOptions = append(baseOptions, options...)
	service, err := NewService(
		users,
		addresses,
		sellers,
		baseOptions...,
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
		AuditFields: domain.AuditFields{
			CreatedBy: "service:backfill",
			UpdatedBy: "service:backfill",
			CreatedAt: fixedUsecaseTime(),
			UpdatedAt: fixedUsecaseTime(),
		},
		StatusAuditFields: domain.StatusAuditFields{
			StatusChangedBy: strPtrUsecase("service:backfill"),
			StatusChangedAt: timePtrUsecase(fixedUsecaseTime()),
		},
	}
}

func validSeller() domain.SellerProfile {
	return domain.SellerProfile{
		SellerID:  "seller_123",
		UserID:    "user_123",
		StoreName: "Aarav Store",
		Status:    domain.SellerStatusDraft,
		AuditFields: domain.AuditFields{
			CreatedBy: "user_123",
			UpdatedBy: "user_123",
			CreatedAt: fixedUsecaseTime(),
			UpdatedAt: fixedUsecaseTime(),
		},
		StatusAuditFields: domain.StatusAuditFields{
			StatusChangedBy: strPtrUsecase("user_123"),
			StatusChangedAt: timePtrUsecase(fixedUsecaseTime()),
		},
	}
}

func validAddress() domain.Address {
	return domain.Address{
		AddressID:  "addr_123",
		UserID:     "user_123",
		Name:       "Aarav Sharma",
		Line1:      "221B MG Road",
		City:       "Bengaluru",
		State:      "Karnataka",
		PostalCode: "560001",
		Country:    "IN",
		Status:     domain.AddressStatusActive,
		AuditFields: domain.AuditFields{
			CreatedBy: "user_123",
			UpdatedBy: "user_123",
			CreatedAt: fixedUsecaseTime(),
			UpdatedAt: fixedUsecaseTime(),
		},
	}
}

func strPtrUsecase(value string) *string {
	return &value
}

func timePtrUsecase(value time.Time) *time.Time {
	return &value
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

type fixedAddressIDGenerator struct {
	id string
}

func (g fixedAddressIDGenerator) NewAddressID(context.Context) (string, error) {
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

func (r *fakeUserRepository) UpdateUserStatus(ctx context.Context, user domain.User) (domain.User, error) {
	return user, nil
}

type fakeAddressRepository struct {
	list                []domain.Address
	found               domain.Address
	created             domain.Address
	updated             domain.Address
	listCalls           int
	listErr             error
	findErr             error
	createErr           error
	updateErr           error
	deleteErr           error
	setDefaultErr       error
	setDefaultAddressID string
	lockCalls           int
	lockErr             error
}

func (r *fakeAddressRepository) LockUserForAddressMutation(ctx context.Context, userID string) error {
	r.lockCalls++
	return r.lockErr
}

func (r *fakeAddressRepository) ListAddresses(ctx context.Context, userID string, limit int, offset int) ([]domain.Address, error) {
	r.listCalls++
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.list, nil
}

func (r *fakeAddressRepository) FindAddress(ctx context.Context, userID string, addressID string) (domain.Address, error) {
	if r.findErr != nil {
		return domain.Address{}, r.findErr
	}
	if r.found.AddressID != "" {
		return r.found, nil
	}
	return validAddress(), nil
}

func (r *fakeAddressRepository) CreateAddress(ctx context.Context, address domain.Address) (domain.Address, error) {
	r.created = address
	if r.createErr != nil {
		return domain.Address{}, r.createErr
	}
	return address, nil
}

func (r *fakeAddressRepository) UpdateAddress(ctx context.Context, address domain.Address) (domain.Address, error) {
	r.updated = address
	if r.updateErr != nil {
		return domain.Address{}, r.updateErr
	}
	return address, nil
}

func (r *fakeAddressRepository) DeleteAddress(ctx context.Context, userID string, addressID string, audit domain.MutationAudit) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	return nil
}

func (r *fakeAddressRepository) SetDefaultAddress(ctx context.Context, userID string, addressID string, audit domain.MutationAudit) error {
	r.setDefaultAddressID = addressID
	if r.setDefaultErr != nil {
		return r.setDefaultErr
	}
	return nil
}

type fakeSellerRepository struct {
	bySellerID  domain.SellerProfile
	byUserID    domain.SellerProfile
	err         error
	updatePatch domain.SellerProfilePatch
}

type fakeEventRecorder struct {
	userCreated    domain.UserCreatedPayload
	sellerApproved domain.SellerApprovedPayload
	addressUpdated domain.AddressUpdatedPayload
	err            error
}

func (r *fakeEventRecorder) RecordUserCreated(ctx context.Context, payload domain.UserCreatedPayload) error {
	r.userCreated = payload
	return r.err
}

func (r *fakeEventRecorder) RecordSellerApproved(ctx context.Context, payload domain.SellerApprovedPayload) error {
	r.sellerApproved = payload
	return r.err
}

func (r *fakeEventRecorder) RecordAddressUpdated(ctx context.Context, payload domain.AddressUpdatedPayload) error {
	r.addressUpdated = payload
	return r.err
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
	r.updatePatch = patch
	updated := r.bySellerID
	if patch.StoreName != nil {
		updated.StoreName = *patch.StoreName
	}
	if patch.DisplayName != nil {
		updated.DisplayName = patch.DisplayName
	}
	if patch.GSTNumber != nil {
		updated.GSTNumber = patch.GSTNumber
	}
	if patch.SupportEmail != nil {
		updated.SupportEmail = patch.SupportEmail
	}
	return updated, nil
}

func (r *fakeSellerRepository) UpdateSellerStatus(ctx context.Context, seller domain.SellerProfile) (domain.SellerProfile, error) {
	return seller, nil
}

func (r *fakeSellerRepository) AddKYCDocument(ctx context.Context, document domain.KYCDocument) (domain.KYCDocument, error) {
	return document, nil
}

func (r *fakeSellerRepository) GetKYCDocument(ctx context.Context, sellerID string, documentID string) (domain.KYCDocument, error) {
	return domain.KYCDocument{}, domain.ErrKYCDocumentNotFound
}

func (r *fakeSellerRepository) ReviewKYCDocument(ctx context.Context, document domain.KYCDocument) (domain.KYCDocument, error) {
	return document, nil
}

func (r *fakeSellerRepository) ListKYCDocuments(ctx context.Context, sellerID string) ([]domain.KYCDocument, error) {
	return nil, nil
}
