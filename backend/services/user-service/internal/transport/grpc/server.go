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
	ListUserAddresses(ctx context.Context, input usecase.ListUserAddressesInput) ([]domain.Address, error)
	CreateAddress(ctx context.Context, input usecase.CreateAddressInput) (domain.Address, error)
	UpdateAddress(ctx context.Context, input usecase.UpdateAddressInput) (domain.Address, error)
	DeleteAddress(ctx context.Context, input usecase.DeleteAddressInput) error
	GetSellerProfile(ctx context.Context, input usecase.GetSellerProfileInput) (domain.SellerProfile, error)
	UpdateSellerProfile(ctx context.Context, input usecase.UpdateSellerProfileInput) (domain.SellerProfile, error)
	UpdateUserStatus(ctx context.Context, input usecase.UpdateUserStatusInput) (domain.User, error)
	UpdateSellerStatus(ctx context.Context, input usecase.UpdateSellerStatusInput) (domain.SellerProfile, error)
	ReviewKYCDocument(ctx context.Context, input usecase.ReviewKYCDocumentInput) (domain.KYCDocument, error)
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

func (s *Server) ListUserAddresses(ctx context.Context, req *userv1.ListUserAddressesRequest) (*userv1.AddressListResponse, error) {
	if strings.TrimSpace(req.GetUserId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	addresses, err := s.users.ListUserAddresses(ctx, usecase.ListUserAddressesInput{
		UserID:   req.GetUserId(),
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
		Caller:   callerFromContext(ctx),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return mapAddressList(addresses, req.GetPage(), req.GetPageSize()), nil
}

func (s *Server) CreateAddress(ctx context.Context, req *userv1.CreateAddressRequest) (*userv1.Address, error) {
	if strings.TrimSpace(req.GetUserId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetAddress() == nil {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	address, err := s.users.CreateAddress(ctx, usecase.CreateAddressInput{
		UserID:     req.GetUserId(),
		Name:       req.GetAddress().GetName(),
		Phone:      req.GetAddress().GetPhone(),
		Line1:      req.GetAddress().GetLine1(),
		Line2:      req.GetAddress().GetLine2(),
		City:       req.GetAddress().GetCity(),
		State:      req.GetAddress().GetState(),
		PostalCode: req.GetAddress().GetPostalCode(),
		Country:    req.GetAddress().GetCountry(),
		IsDefault:  req.GetAddress().GetIsDefault(),
		Caller:     callerFromContext(ctx),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return mapAddress(address), nil
}

func (s *Server) UpdateAddress(ctx context.Context, req *userv1.UpdateAddressRequest) (*userv1.Address, error) {
	if strings.TrimSpace(req.GetUserId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if strings.TrimSpace(req.GetAddressId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "address_id is required")
	}
	if req.GetAddress() == nil {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	address, err := s.users.UpdateAddress(ctx, usecase.UpdateAddressInput{
		UserID:     req.GetUserId(),
		AddressID:  req.GetAddressId(),
		Name:       req.GetAddress().GetName(),
		Phone:      req.GetAddress().GetPhone(),
		Line1:      req.GetAddress().GetLine1(),
		Line2:      req.GetAddress().GetLine2(),
		City:       req.GetAddress().GetCity(),
		State:      req.GetAddress().GetState(),
		PostalCode: req.GetAddress().GetPostalCode(),
		Country:    req.GetAddress().GetCountry(),
		IsDefault:  req.GetAddress().GetIsDefault(),
		Caller:     callerFromContext(ctx),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return mapAddress(address), nil
}

func (s *Server) DeleteAddress(ctx context.Context, req *userv1.DeleteAddressRequest) (*userv1.SuccessResponse, error) {
	if strings.TrimSpace(req.GetUserId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if strings.TrimSpace(req.GetAddressId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "address_id is required")
	}

	if err := s.users.DeleteAddress(ctx, usecase.DeleteAddressInput{
		UserID:    req.GetUserId(),
		AddressID: req.GetAddressId(),
		Caller:    callerFromContext(ctx),
	}); err != nil {
		return nil, mapError(err)
	}
	return &userv1.SuccessResponse{Success: true}, nil
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

func (s *Server) UpdateSellerProfile(ctx context.Context, req *userv1.UpdateSellerProfileRequest) (*userv1.SellerProfile, error) {
	if strings.TrimSpace(req.GetSellerId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "seller_id is required")
	}

	seller, err := s.users.UpdateSellerProfile(ctx, usecase.UpdateSellerProfileInput{
		SellerID:     req.GetSellerId(),
		UserID:       req.GetUserId(),
		StoreName:    req.GetStoreName(),
		DisplayName:  req.GetDisplayName(),
		GSTNumber:    req.GetGstNumber(),
		SupportEmail: req.GetSupportEmail(),
		Mask:         req.GetUpdateMask().GetPaths(),
		Caller:       callerFromContext(ctx),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return mapSellerProfile(seller), nil
}

func (s *Server) UpdateUserStatus(ctx context.Context, req *userv1.UpdateUserStatusRequest) (*userv1.UserProfile, error) {
	if strings.TrimSpace(req.GetUserId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if strings.TrimSpace(req.GetStatus()) == "" {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	user, err := s.users.UpdateUserStatus(ctx, usecase.UpdateUserStatusInput{
		UserID: req.GetUserId(),
		Status: req.GetStatus(),
		Caller: callerFromContext(ctx),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return mapUserProfile(user), nil
}

func (s *Server) UpdateSellerStatus(ctx context.Context, req *userv1.UpdateSellerStatusRequest) (*userv1.SellerProfile, error) {
	if strings.TrimSpace(req.GetSellerId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "seller_id is required")
	}
	if strings.TrimSpace(req.GetStatus()) == "" {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	seller, err := s.users.UpdateSellerStatus(ctx, usecase.UpdateSellerStatusInput{
		SellerID: req.GetSellerId(),
		Status:   req.GetStatus(),
		Reason:   req.GetReason(),
		Caller:   callerFromContext(ctx),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return mapSellerProfile(seller), nil
}

func (s *Server) ReviewKYCDocument(ctx context.Context, req *userv1.ReviewKYCDocumentRequest) (*userv1.KYCDocument, error) {
	if strings.TrimSpace(req.GetSellerId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "seller_id is required")
	}
	if strings.TrimSpace(req.GetDocumentId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "document_id is required")
	}
	if strings.TrimSpace(req.GetStatus()) == "" {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	document, err := s.users.ReviewKYCDocument(ctx, usecase.ReviewKYCDocumentInput{
		SellerID:        req.GetSellerId(),
		DocumentID:      req.GetDocumentId(),
		Status:          req.GetStatus(),
		RejectionReason: req.GetRejectionReason(),
		Caller:          callerFromContext(ctx),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return mapKYCDocument(document), nil
}
