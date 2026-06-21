package clients

import (
	"context"

	userv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/user/v1"
	"google.golang.org/grpc"
)

type UserClient interface {
	GetUser(context.Context, *userv1.GetUserRequest) (*userv1.UserProfile, error)
	UpdateUserProfile(context.Context, *userv1.UpdateUserProfileRequest) (*userv1.UserProfile, error)
	ListUserAddresses(context.Context, *userv1.ListUserAddressesRequest) (*userv1.AddressListResponse, error)
	CreateAddress(context.Context, *userv1.CreateAddressRequest) (*userv1.Address, error)
	UpdateAddress(context.Context, *userv1.UpdateAddressRequest) (*userv1.Address, error)
	DeleteAddress(context.Context, *userv1.DeleteAddressRequest) (*userv1.SuccessResponse, error)
	GetSellerProfile(context.Context, *userv1.GetSellerProfileRequest) (*userv1.SellerProfile, error)
	UpdateSellerProfile(context.Context, *userv1.UpdateSellerProfileRequest) (*userv1.SellerProfile, error)
}

type UserServiceClient interface {
	OutboundClient
	UserClient
	isUserServiceClient()
}

type userServiceClient struct {
	OutboundClient
	rpc userv1.UserServiceClient
}

func newUserServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) UserServiceClient {
	return userServiceClient{
		OutboundClient: newOutboundClient(descriptor, conn),
		rpc:            userv1.NewUserServiceClient(conn),
	}
}

func (userServiceClient) isUserServiceClient() {}

func (c userServiceClient) GetUser(ctx context.Context, request *userv1.GetUserRequest) (*userv1.UserProfile, error) {
	return c.rpc.GetUser(ctx, request)
}

func (c userServiceClient) UpdateUserProfile(ctx context.Context, request *userv1.UpdateUserProfileRequest) (*userv1.UserProfile, error) {
	return c.rpc.UpdateUserProfile(ctx, request)
}

func (c userServiceClient) ListUserAddresses(ctx context.Context, request *userv1.ListUserAddressesRequest) (*userv1.AddressListResponse, error) {
	return c.rpc.ListUserAddresses(ctx, request)
}

func (c userServiceClient) CreateAddress(ctx context.Context, request *userv1.CreateAddressRequest) (*userv1.Address, error) {
	return c.rpc.CreateAddress(ctx, request)
}

func (c userServiceClient) UpdateAddress(ctx context.Context, request *userv1.UpdateAddressRequest) (*userv1.Address, error) {
	return c.rpc.UpdateAddress(ctx, request)
}

func (c userServiceClient) DeleteAddress(ctx context.Context, request *userv1.DeleteAddressRequest) (*userv1.SuccessResponse, error) {
	return c.rpc.DeleteAddress(ctx, request)
}

func (c userServiceClient) GetSellerProfile(ctx context.Context, request *userv1.GetSellerProfileRequest) (*userv1.SellerProfile, error) {
	return c.rpc.GetSellerProfile(ctx, request)
}

func (c userServiceClient) UpdateSellerProfile(ctx context.Context, request *userv1.UpdateSellerProfileRequest) (*userv1.SellerProfile, error) {
	return c.rpc.UpdateSellerProfile(ctx, request)
}
