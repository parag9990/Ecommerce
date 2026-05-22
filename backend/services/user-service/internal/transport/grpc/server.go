package grpc

import (
	"context"
	"log/slog"
	"strings"

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
	"github.com/parag/ecommerce/backend/services/user-service/internal/usecase"
	userv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserUsecase interface {
	CreateUser(ctx context.Context, input usecase.CreateUserInput) (domain.User, error)
	GetUser(ctx context.Context, input usecase.GetUserInput) (domain.User, error)
	UpdateUserProfile(ctx context.Context, input usecase.UpdateUserProfileInput) (domain.User, error)
	GetSellerProfile(ctx context.Context, input usecase.GetSellerProfileInput) (domain.SellerProfile, error)
}

type Server struct {
	userv1.UnimplementedUserServiceServer

	users  UserUsecase
	logger *slog.Logger
}

type Option func(*Server)

func WithLogger(logger *slog.Logger) Option {
	return func(server *Server) {
		if logger != nil {
			server.logger = logger
		}
	}
}

func NewServer(users UserUsecase, options ...Option) *Server {
	if users == nil {
		panic("user usecase is required")
	}

	server := &Server{users: users, logger: slog.Default()}
	for _, option := range options {
		option(server)
	}
	return server
}

func (s *Server) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.UserProfile, error) {
	if strings.TrimSpace(req.GetAuthAccountId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "auth_account_id is required")
	}
	if strings.TrimSpace(req.GetEmail()) == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if strings.TrimSpace(req.GetFullName()) == "" {
		return nil, status.Error(codes.InvalidArgument, "full_name is required")
	}

	user, err := s.users.CreateUser(ctx, usecase.CreateUserInput{
		AuthAccountID: req.GetAuthAccountId(),
		Email:         req.GetEmail(),
		Phone:         req.GetPhone(),
		FullName:      req.GetFullName(),
		Caller:        callerFromContext(ctx),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return mapUserProfile(user), nil
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.UserProfile, error) {
	if strings.TrimSpace(req.GetUserId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := s.users.GetUser(ctx, usecase.GetUserInput{
		UserID: req.GetUserId(),
		Caller: callerFromContext(ctx),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return mapUserProfile(user), nil
}

func (s *Server) UpdateUserProfile(ctx context.Context, req *userv1.UpdateUserProfileRequest) (*userv1.UserProfile, error) {
	if strings.TrimSpace(req.GetUserId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := s.users.UpdateUserProfile(ctx, usecase.UpdateUserProfileInput{
		UserID:    req.GetUserId(),
		FullName:  req.GetFullName(),
		Phone:     req.GetPhone(),
		AvatarURL: req.GetAvatarUrl(),
		Mask:      req.GetUpdateMask().GetPaths(),
		Caller:    callerFromContext(ctx),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return mapUserProfile(user), nil
}

func (s *Server) GetSellerProfile(ctx context.Context, req *userv1.GetSellerProfileRequest) (*userv1.SellerProfile, error) {
	if strings.TrimSpace(req.GetSellerId()) == "" && strings.TrimSpace(req.GetUserId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "seller_id or user_id is required")
	}

	seller, err := s.users.GetSellerProfile(ctx, usecase.GetSellerProfileInput{
		SellerID: req.GetSellerId(),
		UserID:   req.GetUserId(),
		Caller:   callerFromContext(ctx),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return mapSellerProfile(seller), nil
}
