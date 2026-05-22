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

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
)

const (
	profileFieldFullName  = "full_name"
	profileFieldPhone     = "phone"
	profileFieldAvatarURL = "avatar_url"
	defaultIDBytes        = 16
)

type Caller struct {
	UserID      string
	ServiceName string
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

type GetSellerProfileInput struct {
	SellerID string
	UserID   string
	Caller   Caller
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewUserID(ctx context.Context) (string, error)
}

type Service struct {
	users   UserRepository
	sellers SellerRepository
	clock   Clock
	ids     IDGenerator
	logger  *slog.Logger
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

func NewService(users UserRepository, sellers SellerRepository, options ...Option) (*Service, error) {
	if users == nil {
		return nil, errors.New("user repository is required")
	}
	if sellers == nil {
		return nil, errors.New("seller repository is required")
	}

	service := &Service{
		users:   users,
		sellers: sellers,
		clock:   systemClock{},
		ids:     SecureIDGenerator{Prefix: "user", Bytes: defaultIDBytes},
		logger:  slog.Default(),
	}
	for _, option := range options {
		option(service)
	}
	return service, nil
}

func (s *Service) CreateUser(ctx context.Context, input CreateUserInput) (domain.User, error) {
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
		CreatedAt:     s.clock.Now(),
	})
	if err != nil {
		return domain.User{}, err
	}

	created, err := s.users.CreateUser(ctx, user)
	if err != nil {
		s.logError(ctx, "create_user.persist", err, slog.String("user_id", user.UserID))
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

	current, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		s.logError(ctx, "update_user_profile.find", err, slog.String("user_id", userID))
		return domain.User{}, err
	}

	patch := buildRequestedProfilePatch(input, included, s.clock.Now())
	next := current
	if err := next.ApplyProfilePatch(patch); err != nil {
		return domain.User{}, err
	}

	repositoryPatch := buildRepositoryProfilePatch(next, included, patch.UpdatedAt)
	updated, err := s.users.UpdateUserProfile(ctx, userID, repositoryPatch)
	if err != nil {
		s.logError(ctx, "update_user_profile.persist", err, slog.String("user_id", userID))
		return domain.User{}, err
	}
	return updated, nil
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

func buildRequestedProfilePatch(input UpdateUserProfileInput, included map[string]struct{}, updatedAt time.Time) domain.UserProfilePatch {
	patch := domain.UserProfilePatch{UpdatedAt: updatedAt}
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

func buildRepositoryProfilePatch(user domain.User, included map[string]struct{}, updatedAt time.Time) domain.UserProfilePatch {
	patch := domain.UserProfilePatch{UpdatedAt: updatedAt}
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

func requireSelfOrService(caller Caller, userID string) error {
	callerUserID := strings.TrimSpace(caller.UserID)
	if callerUserID == "" || callerUserID == userID {
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
		prefix = "user"
	}
	return prefix + "_" + hex.EncodeToString(random), nil
}
