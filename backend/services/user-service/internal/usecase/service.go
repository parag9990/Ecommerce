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

	"github.com/parag/ecommerce/backend/services/user-service/internal/audit"
	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
	uservalidation "github.com/parag/ecommerce/backend/services/user-service/internal/validation"
	sharedvalidation "github.com/parag/ecommerce/backend/shared/validation"
)

const (
	profileFieldFullName    = "full_name"
	profileFieldPhone       = "phone"
	profileFieldAvatarURL   = "avatar_url"
	sellerFieldStoreName    = "store_name"
	sellerFieldDisplayName  = "display_name"
	sellerFieldGSTNumber    = "gst_number"
	sellerFieldSupportEmail = "support_email"
	defaultIDBytes          = 16
	maxUserAddresses        = 20
	defaultPage             = 1
	defaultPageSize         = 20
	maxPageSize             = 50
)

type Caller struct {
	UserID      string
	ServiceName string
	ActorID     string
	ActorType   string
	Roles       []string
}

type CreateUserInput struct {
	AuthAccountID string
	Email         string
	Phone         string
	FullName      string
	Caller        Caller
}

type GetUserInput struct {
	UserID string
	Caller Caller
}

type UpdateUserProfileInput struct {
	UserID    string
	FullName  string
	Phone     string
	AvatarURL string
	Mask      []string
	Caller    Caller
}

type ListUserAddressesInput struct {
	UserID   string
	Page     int
	PageSize int
	Caller   Caller
}

type CreateAddressInput struct {
	UserID     string
	Name       string
	Phone      string
	Line1      string
	Line2      string
	City       string
	State      string
	PostalCode string
	Country    string
	IsDefault  bool
	Caller     Caller
}

type UpdateAddressInput struct {
	UserID     string
	AddressID  string
	Name       string
	Phone      string
	Line1      string
	Line2      string
	City       string
	State      string
	PostalCode string
	Country    string
	IsDefault  bool
	Caller     Caller
}

type DeleteAddressInput struct {
	UserID    string
	AddressID string
	Caller    Caller
}

type GetSellerProfileInput struct {
	SellerID string
	UserID   string
	Caller   Caller
}

type UpdateSellerProfileInput struct {
	SellerID     string
	UserID       string
	StoreName    string
	DisplayName  string
	GSTNumber    string
	SupportEmail string
	Mask         []string
	Caller       Caller
}

type UpdateUserStatusInput struct {
	UserID string
	Status string
	Caller Caller
}

type UpdateSellerStatusInput struct {
	SellerID string
	Status   string
	Reason   string
	Caller   Caller
}

type ReviewKYCDocumentInput struct {
	SellerID        string
	DocumentID      string
	Status          string
	RejectionReason string
	Caller          Caller
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewUserID(ctx context.Context) (string, error)
}

type AddressIDGenerator interface {
	NewAddressID(ctx context.Context) (string, error)
}

type Service struct {
	users      UserRepository
	addresses  AddressRepository
	sellers    SellerRepository
	events     EventRecorder
	unitOfWork UnitOfWork
	validator  InputValidator
	clock      Clock
	ids        IDGenerator
	addressIDs AddressIDGenerator
	logger     *slog.Logger
}

type Option func(*Service)

func WithLogger(logger *slog.Logger) Option {
	return func(service *Service) {
		if logger != nil {
			service.logger = logger
		}
	}
}

func WithClock(clock Clock) Option {
	return func(service *Service) {
		if clock != nil {
			service.clock = clock
		}
	}
}

func WithIDGenerator(ids IDGenerator) Option {
	return func(service *Service) {
		if ids != nil {
			service.ids = ids
		}
	}
}

func WithAddressIDGenerator(ids AddressIDGenerator) Option {
	return func(service *Service) {
		if ids != nil {
			service.addressIDs = ids
		}
	}
}

func WithValidationPhoneRegion(region string) Option {
	return func(service *Service) {
		service.validator = uservalidation.NewUserValidator(uservalidation.Config{PhoneRegion: strings.TrimSpace(region)})
	}
}

func WithInputValidator(validator InputValidator) Option {
	return func(service *Service) {
		if validator != nil {
			service.validator = validator
		}
	}
}

func WithEventRecorder(recorder EventRecorder) Option {
	return func(service *Service) {
		if recorder != nil {
			service.events = recorder
		}
	}
}

func WithUnitOfWork(unitOfWork UnitOfWork) Option {
	return func(service *Service) {
		if unitOfWork != nil {
			service.unitOfWork = unitOfWork
		}
	}
}

func NewService(users UserRepository, addresses AddressRepository, sellers SellerRepository, options ...Option) (*Service, error) {
	if users == nil {
		return nil, errors.New("user repository is required")
	}
	if addresses == nil {
		return nil, errors.New("address repository is required")
	}
	if sellers == nil {
		return nil, errors.New("seller repository is required")
	}

	service := &Service{
		users:      users,
		addresses:  addresses,
		sellers:    sellers,
		events:     noopEventRecorder{},
		validator:  uservalidation.NewUserValidator(uservalidation.Config{}),
		clock:      systemClock{},
		ids:        SecureIDGenerator{Prefix: "user", Bytes: defaultIDBytes},
		addressIDs: SecureIDGenerator{Prefix: "addr", Bytes: defaultIDBytes},
		logger:     slog.Default(),
	}
	for _, option := range options {
		option(service)
	}
	return service, nil
}

func (s *Service) CreateUser(ctx context.Context, input CreateUserInput) (domain.User, error) {
	validated, validationErr := s.validator.NormalizeAndValidateCreateUser(uservalidation.CreateUserInput{
		AuthAccountID: input.AuthAccountID,
		Email:         input.Email,
		Phone:         input.Phone,
		FullName:      input.FullName,
	})
	if validationErr.HasErrors() {
		return domain.User{}, toDomainValidationError(validationErr)
	}
	input.AuthAccountID = validated.AuthAccountID
	input.Email = validated.Email
	input.Phone = validated.Phone
	input.FullName = validated.FullName

	actor, err := s.writeActor(ctx, input.Caller)
	if err != nil {
		return domain.User{}, err
	}
	actorID := actor.AuditID()

	userID, err := s.ids.NewUserID(ctx)
	if err != nil {
		s.logError(ctx, "create_user.generate_id", err)
		return domain.User{}, fmt.Errorf("generate user id: %w", err)
	}

	phone := optionalString(input.Phone)
	user, err := domain.NewUser(domain.NewUserParams{
		UserID:        userID,
		AuthAccountID: input.AuthAccountID,
		Email:         input.Email,
		Phone:         phone,
		FullName:      input.FullName,
		CreatedBy:     actorID,
		CreatedAt:     s.clock.Now(),
	})
	if err != nil {
		return domain.User{}, err
	}

	var created domain.User
	if err := s.withWriteRepositories(ctx, func(ctx context.Context, repositories TransactionRepositories) error {
		var err error
		created, err = repositories.Users.CreateUser(ctx, user)
		if err != nil {
			s.logError(ctx, "create_user.persist", err, slog.String("user_id", user.UserID))
			return err
		}
		if err := repositories.Events.RecordUserCreated(ctx, buildUserCreatedPayload(created)); err != nil {
			s.logError(ctx, "create_user.record_event", err, slog.String("user_id", user.UserID))
			return fmt.Errorf("record user created event: %w", err)
		}
		return nil
	}); err != nil {
		return domain.User{}, err
	}
	return created, nil
}

func (s *Service) GetUser(ctx context.Context, input GetUserInput) (domain.User, error) {
	userID, err := requiredID("user_id", input.UserID)
	if err != nil {
		return domain.User{}, err
	}
	if err := requireSelfOrService(input.Caller, userID); err != nil {
		return domain.User{}, err
	}

	user, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		s.logError(ctx, "get_user", err, slog.String("user_id", userID))
		return domain.User{}, err
	}
	return user, nil
}

func (s *Service) UpdateUserProfile(ctx context.Context, input UpdateUserProfileInput) (domain.User, error) {
	userID, err := requiredID("user_id", input.UserID)
	if err != nil {
		return domain.User{}, err
	}
	if err := requireSelfOrService(input.Caller, userID); err != nil {
		return domain.User{}, err
	}

	included, err := parseProfileMask(input.Mask)
	if err != nil {
		return domain.User{}, err
	}

	validated, validationErr := s.validator.NormalizeAndValidateUpdateProfile(uservalidation.UpdateProfileInput{
		FullName:  input.FullName,
		Phone:     input.Phone,
		AvatarURL: input.AvatarURL,
	}, profileUpdateFields(included))
	if validationErr.HasErrors() {
		return domain.User{}, toDomainValidationError(validationErr)
	}
	input.FullName = validated.FullName
	input.Phone = validated.Phone
	input.AvatarURL = validated.AvatarURL

	actor, err := s.writeActor(ctx, input.Caller)
	if err != nil {
		return domain.User{}, err
	}

	current, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		s.logError(ctx, "update_user_profile.find", err, slog.String("user_id", userID))
		return domain.User{}, err
	}

	patch := buildRequestedProfilePatch(input, included, actor.AuditID(), s.clock.Now())
	next := current
	if err := next.ApplyProfilePatch(patch); err != nil {
		return domain.User{}, err
	}

	repositoryPatch := buildRepositoryProfilePatch(next, included, patch.UpdatedBy, patch.UpdatedAt)
	updated, err := s.users.UpdateUserProfile(ctx, userID, repositoryPatch)
	if err != nil {
		s.logError(ctx, "update_user_profile.persist", err, slog.String("user_id", userID))
		return domain.User{}, err
	}
	return updated, nil
}

func (s *Service) ListUserAddresses(ctx context.Context, input ListUserAddressesInput) ([]domain.Address, error) {
	userID, err := requiredID("user_id", input.UserID)
	if err != nil {
		return nil, err
	}
	if err := requireSelfOrService(input.Caller, userID); err != nil {
		return nil, err
	}

	page := normalizePage(input.Page)
	pageSize := normalizePageSize(input.PageSize)
	offset := (page - 1) * pageSize

	addresses, err := s.addresses.ListAddresses(ctx, userID, pageSize, offset)
	if err != nil {
		s.logError(ctx, "list_user_addresses", err, slog.String("user_id", userID))
		return nil, err
	}
	return addresses, nil
}

func (s *Service) CreateAddress(ctx context.Context, input CreateAddressInput) (domain.Address, error) {
	userID, err := requiredID("user_id", input.UserID)
	if err != nil {
		return domain.Address{}, err
	}
	if err := requireSelfOrService(input.Caller, userID); err != nil {
		return domain.Address{}, err
	}
	actor, err := s.writeActor(ctx, input.Caller)
	if err != nil {
		return domain.Address{}, err
	}
	actorID := actor.AuditID()

	validated, validationErr := s.validator.NormalizeAndValidateAddress(uservalidation.AddressInput{
		Name:       input.Name,
		Phone:      input.Phone,
		Line1:      input.Line1,
		Line2:      input.Line2,
		City:       input.City,
		State:      input.State,
		PostalCode: input.PostalCode,
		Country:    input.Country,
	})
	if validationErr.HasErrors() {
		return domain.Address{}, toDomainValidationError(validationErr)
	}
	input.Name = validated.Name
	input.Phone = validated.Phone
	input.Line1 = validated.Line1
	input.Line2 = validated.Line2
	input.City = validated.City
	input.State = validated.State
	input.PostalCode = validated.PostalCode
	input.Country = validated.Country

	var created domain.Address
	if err := s.withWriteRepositories(ctx, func(ctx context.Context, repositories TransactionRepositories) error {
		existing, err := repositories.Addresses.ListAddresses(ctx, userID, maxUserAddresses+1, 0)
		if err != nil {
			s.logError(ctx, "create_address.count_existing", err, slog.String("user_id", userID))
			return err
		}
		if len(existing) >= maxUserAddresses {
			return domain.ErrAddressLimitExceeded
		}

		addressID, err := s.addressIDs.NewAddressID(ctx)
		if err != nil {
			s.logError(ctx, "create_address.generate_id", err, slog.String("user_id", userID))
			return fmt.Errorf("generate address id: %w", err)
		}

		address, err := domain.NewAddress(domain.NewAddressParams{
			AddressID:  addressID,
			UserID:     userID,
			Name:       input.Name,
			Phone:      optionalString(input.Phone),
			Line1:      input.Line1,
			Line2:      optionalString(input.Line2),
			City:       input.City,
			State:      input.State,
			PostalCode: input.PostalCode,
			Country:    input.Country,
			IsDefault:  input.IsDefault || len(existing) == 0,
			CreatedBy:  actorID,
			CreatedAt:  s.clock.Now(),
		})
		if err != nil {
			return err
		}

		created, err = repositories.Addresses.CreateAddress(ctx, address)
		if err != nil {
			s.logError(ctx, "create_address.persist", err, slog.String("user_id", userID), slog.String("address_id", address.AddressID))
			return err
		}
		if err := repositories.Events.RecordAddressUpdated(ctx, buildAddressUpdatedPayload(created, domain.AddressChangeCreated, actorID, created.UpdatedAt)); err != nil {
			s.logError(ctx, "create_address.record_event", err, slog.String("user_id", userID), slog.String("address_id", created.AddressID))
			return fmt.Errorf("record address updated event: %w", err)
		}
		return nil
	}); err != nil {
		return domain.Address{}, err
	}
	return created, nil
}

func (s *Service) UpdateAddress(ctx context.Context, input UpdateAddressInput) (domain.Address, error) {
	userID, err := requiredID("user_id", input.UserID)
	if err != nil {
		return domain.Address{}, err
	}
	addressID, err := requiredID("address_id", input.AddressID)
	if err != nil {
		return domain.Address{}, err
	}
	if err := requireSelfOrService(input.Caller, userID); err != nil {
		return domain.Address{}, err
	}
	actor, err := s.writeActor(ctx, input.Caller)
	if err != nil {
		return domain.Address{}, err
	}

	validated, validationErr := s.validator.NormalizeAndValidateAddress(uservalidation.AddressInput{
		Name:       input.Name,
		Phone:      input.Phone,
		Line1:      input.Line1,
		Line2:      input.Line2,
		City:       input.City,
		State:      input.State,
		PostalCode: input.PostalCode,
		Country:    input.Country,
	})
	if validationErr.HasErrors() {
		return domain.Address{}, toDomainValidationError(validationErr)
	}
	input.Name = validated.Name
	input.Phone = validated.Phone
	input.Line1 = validated.Line1
	input.Line2 = validated.Line2
	input.City = validated.City
	input.State = validated.State
	input.PostalCode = validated.PostalCode
	input.Country = validated.Country

	var updated domain.Address
	if err := s.withWriteRepositories(ctx, func(ctx context.Context, repositories TransactionRepositories) error {
		current, err := repositories.Addresses.FindAddress(ctx, userID, addressID)
		if err != nil {
			s.logError(ctx, "update_address.find", err, slog.String("user_id", userID), slog.String("address_id", addressID))
			return err
		}
		wasDefault := current.IsDefault

		patch := domain.AddressPatch{
			Name:       &input.Name,
			Phone:      &input.Phone,
			Line1:      &input.Line1,
			Line2:      &input.Line2,
			City:       &input.City,
			State:      &input.State,
			PostalCode: &input.PostalCode,
			Country:    &input.Country,
			UpdatedBy:  actor.AuditID(),
			UpdatedAt:  s.clock.Now(),
		}
		if err := current.ApplyPatch(patch); err != nil {
			return err
		}

		updated, err = repositories.Addresses.UpdateAddress(ctx, current)
		if err != nil {
			s.logError(ctx, "update_address.persist", err, slog.String("user_id", userID), slog.String("address_id", addressID))
			return err
		}

		changeType := domain.AddressChangeUpdated
		if input.IsDefault {
			if err := repositories.Addresses.SetDefaultAddress(ctx, userID, addressID, domain.NewMutationAudit(actor.AuditID(), patch.UpdatedAt)); err != nil {
				s.logError(ctx, "update_address.set_default", err, slog.String("user_id", userID), slog.String("address_id", addressID))
				return err
			}
			updated, err = repositories.Addresses.FindAddress(ctx, userID, addressID)
			if err != nil {
				s.logError(ctx, "update_address.find_after_default", err, slog.String("user_id", userID), slog.String("address_id", addressID))
				return err
			}
			if !wasDefault {
				changeType = domain.AddressChangeDefaultChanged
			}
		}

		if err := repositories.Events.RecordAddressUpdated(ctx, buildAddressUpdatedPayload(updated, changeType, actor.AuditID(), updated.UpdatedAt)); err != nil {
			s.logError(ctx, "update_address.record_event", err, slog.String("user_id", userID), slog.String("address_id", addressID))
			return fmt.Errorf("record address updated event: %w", err)
		}
		return nil
	}); err != nil {
		return domain.Address{}, err
	}
	return updated, nil
}

func (s *Service) DeleteAddress(ctx context.Context, input DeleteAddressInput) error {
	userID, err := requiredID("user_id", input.UserID)
	if err != nil {
		return err
	}
	addressID, err := requiredID("address_id", input.AddressID)
	if err != nil {
		return err
	}
	if err := requireSelfOrService(input.Caller, userID); err != nil {
		return err
	}
	actor, err := s.writeActor(ctx, input.Caller)
	if err != nil {
		return err
	}

	auditRecord := domain.NewMutationAudit(actor.AuditID(), s.clock.Now())
	return s.withWriteRepositories(ctx, func(ctx context.Context, repositories TransactionRepositories) error {
		current, err := repositories.Addresses.FindAddress(ctx, userID, addressID)
		if err != nil {
			s.logError(ctx, "delete_address.find", err, slog.String("user_id", userID), slog.String("address_id", addressID))
			return err
		}
		if err := repositories.Addresses.DeleteAddress(ctx, userID, addressID, auditRecord); err != nil {
			s.logError(ctx, "delete_address", err, slog.String("user_id", userID), slog.String("address_id", addressID))
			return err
		}

		current.IsDefault = false
		current.UpdatedBy = auditRecord.ActorID
		current.UpdatedAt = auditRecord.At
		if err := repositories.Events.RecordAddressUpdated(ctx, buildAddressUpdatedPayload(current, domain.AddressChangeDeleted, actor.AuditID(), auditRecord.At)); err != nil {
			s.logError(ctx, "delete_address.record_event", err, slog.String("user_id", userID), slog.String("address_id", addressID))
			return fmt.Errorf("record address updated event: %w", err)
		}
		return nil
	})
}

func (s *Service) GetSellerProfile(ctx context.Context, input GetSellerProfileInput) (domain.SellerProfile, error) {
	sellerID := strings.TrimSpace(input.SellerID)
	userID := strings.TrimSpace(input.UserID)
	if sellerID == "" && userID == "" {
		return domain.SellerProfile{}, validationError("seller_id", "seller_id or user_id is required")
	}

	if sellerID != "" {
		seller, err := s.sellers.GetSellerProfileBySellerID(ctx, sellerID)
		if err != nil {
			s.logError(ctx, "get_seller_profile.by_seller_id", err, slog.String("seller_id", sellerID))
			return domain.SellerProfile{}, err
		}
		if userID != "" && seller.UserID != userID {
			return domain.SellerProfile{}, domain.ErrForbidden
		}
		return seller, nil
	}

	seller, err := s.sellers.GetSellerProfileByUserID(ctx, userID)
	if err != nil {
		s.logError(ctx, "get_seller_profile.by_user_id", err, slog.String("user_id", userID))
		return domain.SellerProfile{}, err
	}
	return seller, nil
}

func (s *Service) UpdateSellerProfile(ctx context.Context, input UpdateSellerProfileInput) (domain.SellerProfile, error) {
	sellerID, err := requiredID("seller_id", input.SellerID)
	if err != nil {
		return domain.SellerProfile{}, err
	}

	current, err := s.sellers.GetSellerProfileBySellerID(ctx, sellerID)
	if err != nil {
		s.logError(ctx, "update_seller_profile.find", err, slog.String("seller_id", sellerID))
		return domain.SellerProfile{}, err
	}

	userID := strings.TrimSpace(input.UserID)
	if userID != "" && current.UserID != userID {
		return domain.SellerProfile{}, domain.ErrForbidden
	}
	if err := requireSelfOrService(input.Caller, current.UserID); err != nil {
		return domain.SellerProfile{}, err
	}
	actor, err := s.writeActor(ctx, input.Caller)
	if err != nil {
		return domain.SellerProfile{}, err
	}

	included, err := parseSellerMask(input.Mask)
	if err != nil {
		return domain.SellerProfile{}, err
	}

	validated, validationErr := s.validator.NormalizeAndValidateUpdateSellerProfile(uservalidation.UpdateSellerProfileInput{
		StoreName:    input.StoreName,
		DisplayName:  input.DisplayName,
		GSTNumber:    input.GSTNumber,
		SupportEmail: input.SupportEmail,
	}, sellerUpdateFields(included))
	if validationErr.HasErrors() {
		return domain.SellerProfile{}, toDomainValidationError(validationErr)
	}
	input.StoreName = validated.StoreName
	input.DisplayName = validated.DisplayName
	input.GSTNumber = validated.GSTNumber
	input.SupportEmail = validated.SupportEmail

	patch := buildRequestedSellerPatch(input, included, actor.AuditID(), s.clock.Now())
	next := current
	if err := next.ApplyPatch(patch); err != nil {
		return domain.SellerProfile{}, err
	}

	repositoryPatch := buildRepositorySellerPatch(next, included, patch.UpdatedBy, patch.UpdatedAt)
	updated, err := s.sellers.UpdateSellerProfile(ctx, sellerID, repositoryPatch)
	if err != nil {
		s.logError(ctx, "update_seller_profile.persist", err, slog.String("seller_id", sellerID))
		return domain.SellerProfile{}, err
	}
	return updated, nil
}

func (s *Service) UpdateUserStatus(ctx context.Context, input UpdateUserStatusInput) (domain.User, error) {
	userID, err := requiredID("user_id", input.UserID)
	if err != nil {
		return domain.User{}, err
	}
	actor, err := s.writeActor(ctx, input.Caller)
	if err != nil {
		return domain.User{}, err
	}
	if !isPrivilegedActor(actor) {
		return domain.User{}, domain.ErrForbidden
	}

	to := domain.UserStatus(strings.TrimSpace(input.Status))
	if !to.Valid() {
		return domain.User{}, validationError("status", "is not supported")
	}

	current, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		s.logError(ctx, "update_user_status.find", err, slog.String("user_id", userID))
		return domain.User{}, err
	}

	next := current
	if err := next.TransitionStatus(to, actor.AuditID(), s.clock.Now()); err != nil {
		return domain.User{}, err
	}

	updated, err := s.users.UpdateUserStatus(ctx, next)
	if err != nil {
		s.logError(ctx, "update_user_status.persist", err, slog.String("user_id", userID))
		return domain.User{}, err
	}
	return updated, nil
}

func (s *Service) UpdateSellerStatus(ctx context.Context, input UpdateSellerStatusInput) (domain.SellerProfile, error) {
	sellerID, err := requiredID("seller_id", input.SellerID)
	if err != nil {
		return domain.SellerProfile{}, err
	}
	actor, err := s.writeActor(ctx, input.Caller)
	if err != nil {
		return domain.SellerProfile{}, err
	}

	to := domain.SellerStatus(strings.TrimSpace(input.Status))
	if !to.Valid() {
		return domain.SellerProfile{}, validationError("status", "is not supported")
	}

	var updated domain.SellerProfile
	if err := s.withWriteRepositories(ctx, func(ctx context.Context, repositories TransactionRepositories) error {
		current, err := repositories.Sellers.GetSellerProfileBySellerID(ctx, sellerID)
		if err != nil {
			s.logError(ctx, "update_seller_status.find", err, slog.String("seller_id", sellerID))
			return err
		}
		if err := authorizeSellerStatusActor(actor, current, to); err != nil {
			return err
		}

		now := s.clock.Now()
		next := current
		actorID := actor.AuditID()
		switch to {
		case domain.SellerStatusPendingReview:
			err = next.SubmitForReview(actorID, now)
		case domain.SellerStatusActive:
			if current.Status == domain.SellerStatusSuspended {
				err = next.Reactivate(actorID, input.Reason, now)
			} else {
				err = next.Approve(actorID, now)
			}
		case domain.SellerStatusRejected:
			err = next.Reject(actorID, input.Reason, now)
		case domain.SellerStatusSuspended:
			err = next.Suspend(actorID, input.Reason, now)
		default:
			err = domain.ErrInvalidTransition
		}
		if err != nil {
			return err
		}

		updated, err = repositories.Sellers.UpdateSellerStatus(ctx, next)
		if err != nil {
			s.logError(ctx, "update_seller_status.persist", err, slog.String("seller_id", sellerID))
			return err
		}
		if sellerApprovalTriggersEvent(current.Status, updated.Status) {
			if err := repositories.Events.RecordSellerApproved(ctx, buildSellerApprovedPayload(current, updated, input.Reason)); err != nil {
				s.logError(ctx, "update_seller_status.record_event", err, slog.String("seller_id", sellerID))
				return fmt.Errorf("record seller approved event: %w", err)
			}
		}
		return nil
	}); err != nil {
		return domain.SellerProfile{}, err
	}
	return updated, nil
}

func (s *Service) ReviewKYCDocument(ctx context.Context, input ReviewKYCDocumentInput) (domain.KYCDocument, error) {
	sellerID, err := requiredID("seller_id", input.SellerID)
	if err != nil {
		return domain.KYCDocument{}, err
	}
	documentID, err := requiredID("document_id", input.DocumentID)
	if err != nil {
		return domain.KYCDocument{}, err
	}
	actor, err := s.writeActor(ctx, input.Caller)
	if err != nil {
		return domain.KYCDocument{}, err
	}
	if !isPrivilegedActor(actor) {
		return domain.KYCDocument{}, domain.ErrForbidden
	}

	to := domain.KYCStatus(strings.TrimSpace(input.Status))
	if to != domain.KYCStatusApproved && to != domain.KYCStatusRejected {
		return domain.KYCDocument{}, validationError("status", "must be approved or rejected")
	}

	current, err := s.sellers.GetKYCDocument(ctx, sellerID, documentID)
	if err != nil {
		s.logError(ctx, "review_kyc_document.find", err, slog.String("seller_id", sellerID), slog.String("document_id", documentID))
		return domain.KYCDocument{}, err
	}

	next := current
	actorID := actor.AuditID()
	now := s.clock.Now()
	if to == domain.KYCStatusApproved {
		err = next.Approve(actorID, now)
	} else {
		err = next.Reject(actorID, input.RejectionReason, now)
	}
	if err != nil {
		return domain.KYCDocument{}, err
	}

	updated, err := s.sellers.ReviewKYCDocument(ctx, next)
	if err != nil {
		s.logError(ctx, "review_kyc_document.persist", err, slog.String("seller_id", sellerID), slog.String("document_id", documentID))
		return domain.KYCDocument{}, err
	}
	return updated, nil
}

func parseProfileMask(paths []string) (map[string]struct{}, error) {
	if len(paths) == 0 {
		return nil, validationError("update_mask", "at least one path is required")
	}

	included := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		normalized := strings.TrimSpace(path)
		switch normalized {
		case profileFieldFullName, profileFieldPhone, profileFieldAvatarURL:
			included[normalized] = struct{}{}
		default:
			return nil, validationError("update_mask", fmt.Sprintf("unsupported path %q", normalized))
		}
	}
	if len(included) == 0 {
		return nil, validationError("update_mask", "at least one path is required")
	}
	return included, nil
}

func buildRequestedProfilePatch(input UpdateUserProfileInput, included map[string]struct{}, updatedBy string, updatedAt time.Time) domain.UserProfilePatch {
	patch := domain.UserProfilePatch{UpdatedBy: updatedBy, UpdatedAt: updatedAt}
	if hasPath(included, profileFieldFullName) {
		value := input.FullName
		patch.FullName = &value
	}
	if hasPath(included, profileFieldPhone) {
		value := input.Phone
		patch.Phone = &value
	}
	if hasPath(included, profileFieldAvatarURL) {
		value := input.AvatarURL
		patch.AvatarURL = &value
	}
	return patch
}

func buildRepositoryProfilePatch(user domain.User, included map[string]struct{}, updatedBy string, updatedAt time.Time) domain.UserProfilePatch {
	patch := domain.UserProfilePatch{UpdatedBy: updatedBy, UpdatedAt: updatedAt}
	if hasPath(included, profileFieldFullName) {
		value := user.FullName
		patch.FullName = &value
	}
	if hasPath(included, profileFieldPhone) {
		patch.Phone = nullableUpdateValue(user.Phone)
	}
	if hasPath(included, profileFieldAvatarURL) {
		patch.AvatarURL = nullableUpdateValue(user.AvatarURL)
	}
	return patch
}

func profileUpdateFields(included map[string]struct{}) uservalidation.ProfileUpdateFields {
	return uservalidation.ProfileUpdateFields{
		FullName:  hasPath(included, profileFieldFullName),
		Phone:     hasPath(included, profileFieldPhone),
		AvatarURL: hasPath(included, profileFieldAvatarURL),
	}
}

func parseSellerMask(paths []string) (map[string]struct{}, error) {
	if len(paths) == 0 {
		return nil, validationError("update_mask", "at least one path is required")
	}

	included := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		normalized := strings.TrimSpace(path)
		switch normalized {
		case sellerFieldStoreName, sellerFieldDisplayName, sellerFieldGSTNumber, sellerFieldSupportEmail:
			included[normalized] = struct{}{}
		default:
			return nil, validationError("update_mask", fmt.Sprintf("unsupported path %q", normalized))
		}
	}
	if len(included) == 0 {
		return nil, validationError("update_mask", "at least one path is required")
	}
	return included, nil
}

func buildRequestedSellerPatch(input UpdateSellerProfileInput, included map[string]struct{}, updatedBy string, updatedAt time.Time) domain.SellerProfilePatch {
	patch := domain.SellerProfilePatch{UpdatedBy: updatedBy, UpdatedAt: updatedAt}
	if hasPath(included, sellerFieldStoreName) {
		value := input.StoreName
		patch.StoreName = &value
	}
	if hasPath(included, sellerFieldDisplayName) {
		value := input.DisplayName
		patch.DisplayName = &value
	}
	if hasPath(included, sellerFieldGSTNumber) {
		value := input.GSTNumber
		patch.GSTNumber = &value
	}
	if hasPath(included, sellerFieldSupportEmail) {
		value := input.SupportEmail
		patch.SupportEmail = &value
	}
	return patch
}

func buildRepositorySellerPatch(seller domain.SellerProfile, included map[string]struct{}, updatedBy string, updatedAt time.Time) domain.SellerProfilePatch {
	patch := domain.SellerProfilePatch{UpdatedBy: updatedBy, UpdatedAt: updatedAt}
	if hasPath(included, sellerFieldStoreName) {
		value := seller.StoreName
		patch.StoreName = &value
	}
	if hasPath(included, sellerFieldDisplayName) {
		patch.DisplayName = nullableUpdateValue(seller.DisplayName)
	}
	if hasPath(included, sellerFieldGSTNumber) {
		patch.GSTNumber = nullableUpdateValue(seller.GSTNumber)
	}
	if hasPath(included, sellerFieldSupportEmail) {
		patch.SupportEmail = nullableUpdateValue(seller.SupportEmail)
	}
	return patch
}

func sellerUpdateFields(included map[string]struct{}) uservalidation.SellerUpdateFields {
	return uservalidation.SellerUpdateFields{
		StoreName:    hasPath(included, sellerFieldStoreName),
		DisplayName:  hasPath(included, sellerFieldDisplayName),
		GSTNumber:    hasPath(included, sellerFieldGSTNumber),
		SupportEmail: hasPath(included, sellerFieldSupportEmail),
	}
}

func normalizePage(page int) int {
	if page < 1 {
		return defaultPage
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize < 1 {
		return defaultPageSize
	}
	if pageSize > maxPageSize {
		return maxPageSize
	}
	return pageSize
}

func nullableUpdateValue(value *string) *string {
	if value == nil {
		empty := ""
		return &empty
	}
	copied := *value
	return &copied
}

func hasPath(included map[string]struct{}, path string) bool {
	_, ok := included[path]
	return ok
}

func requiredID(field string, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", validationError(field, "is required")
	}
	return trimmed, nil
}

func validationError(field string, message string) error {
	return domain.ValidationError{Fields: []domain.FieldError{{Field: field, Message: message}}}
}

func toDomainValidationError(validationErr sharedvalidation.Error) error {
	return domain.ValidationError{Fields: validationErr.Fields}
}

func requireSelfOrService(caller Caller, userID string) error {
	callerUserID := strings.TrimSpace(caller.UserID)
	if callerUserID == "" || callerUserID == userID {
		return nil
	}
	return domain.ErrForbidden
}

func (s *Service) writeActor(ctx context.Context, caller Caller) (audit.Actor, error) {
	actor, err := audit.ActorFromContext(ctx)
	if err == nil {
		return actor, nil
	}
	if !errors.Is(err, audit.ErrMissingActor) {
		return audit.Actor{}, err
	}
	return actorFromCaller(caller)
}

func actorFromCaller(caller Caller) (audit.Actor, error) {
	actorID := strings.TrimSpace(caller.ActorID)
	actorType := audit.ActorType(strings.TrimSpace(caller.ActorType))
	if actorID != "" || actorType != "" {
		if actorType == "" {
			if actorID == strings.TrimSpace(caller.UserID) {
				actorType = audit.ActorTypeUser
			} else if actorID == strings.TrimSpace(caller.ServiceName) {
				actorType = audit.ActorTypeService
			}
		}
		return audit.NewActor(actorID, actorType)
	}

	if userID := strings.TrimSpace(caller.UserID); userID != "" {
		return audit.NewActor(userID, audit.ActorTypeUser)
	}
	if serviceName := strings.TrimSpace(caller.ServiceName); serviceName != "" {
		return audit.NewActor(serviceName, audit.ActorTypeService)
	}
	return audit.Actor{}, audit.ErrMissingActor
}

func isPrivilegedActor(actor audit.Actor) bool {
	return actor.Type == audit.ActorTypeAdmin || actor.Type == audit.ActorTypeService || actor.Type == audit.ActorTypeSystem
}

func authorizeSellerStatusActor(actor audit.Actor, seller domain.SellerProfile, to domain.SellerStatus) error {
	if isPrivilegedActor(actor) {
		return nil
	}
	if actor.Type == audit.ActorTypeUser && actor.ID == seller.UserID && to == domain.SellerStatusPendingReview {
		return nil
	}
	return domain.ErrForbidden
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (s *Service) logError(ctx context.Context, operation string, err error, attrs ...slog.Attr) {
	if s.logger == nil || err == nil {
		return
	}

	args := []any{
		slog.String("operation", operation),
		slog.String("error_type", fmt.Sprintf("%T", err)),
	}
	for _, attr := range attrs {
		args = append(args, attr)
	}
	s.logger.ErrorContext(ctx, "user_usecase_error", args...)
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

type SecureIDGenerator struct {
	Prefix string
	Bytes  int
}

func (g SecureIDGenerator) NewUserID(ctx context.Context) (string, error) {
	return g.newID(ctx, "user")
}

func (g SecureIDGenerator) NewAddressID(ctx context.Context) (string, error) {
	return g.newID(ctx, "addr")
}

func (g SecureIDGenerator) newID(ctx context.Context, fallbackPrefix string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	size := g.Bytes
	if size <= 0 {
		size = defaultIDBytes
	}
	random := make([]byte, size)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}

	prefix := strings.TrimSpace(g.Prefix)
	if prefix == "" {
		prefix = fallbackPrefix
	}
	return prefix + "_" + hex.EncodeToString(random), nil
}
