package clients

import (
	"context"

	userv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/user/v1"
	"google.golang.org/grpc"
)

type UserClient interface {
	GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.UserProfile, error)
	UpdateUserProfile(ctx context.Context, req *userv1.UpdateUserProfileRequest) (*userv1.UserProfile, error)
	ListUserAddresses(ctx context.Context, req *userv1.ListUserAddressesRequest) (*userv1.AddressListResponse, error)
	CreateAddress(ctx context.Context, req *userv1.CreateAddressRequest) (*userv1.Address, error)
	UpdateAddress(ctx context.Context, req *userv1.UpdateAddressRequest) (*userv1.Address, error)
	DeleteAddress(ctx context.Context, req *userv1.DeleteAddressRequest) (*userv1.SuccessResponse, error)
	GetSellerProfile(ctx context.Context, req *userv1.GetSellerProfileRequest) (*userv1.SellerProfile, error)
	UpdateSellerProfile(ctx context.Context, req *userv1.UpdateSellerProfileRequest) (*userv1.SellerProfile, error)
}

type GRPCUserClient struct {
	client userv1.UserServiceClient
}

func NewGRPCUserClient(conn grpc.ClientConnInterface) *GRPCUserClient {
	return &GRPCUserClient{client: userv1.NewUserServiceClient(conn)}
}

func (c *GRPCUserClient) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.UserProfile, error) {
	return c.client.GetUser(ctx, req)
}

func (c *GRPCUserClient) UpdateUserProfile(ctx context.Context, req *userv1.UpdateUserProfileRequest) (*userv1.UserProfile, error) {
	return c.client.UpdateUserProfile(ctx, req)
}

func (c *GRPCUserClient) ListUserAddresses(ctx context.Context, req *userv1.ListUserAddressesRequest) (*userv1.AddressListResponse, error) {
	return c.client.ListUserAddresses(ctx, req)
}

func (c *GRPCUserClient) CreateAddress(ctx context.Context, req *userv1.CreateAddressRequest) (*userv1.Address, error) {
	return c.client.CreateAddress(ctx, req)
}

func (c *GRPCUserClient) UpdateAddress(ctx context.Context, req *userv1.UpdateAddressRequest) (*userv1.Address, error) {
	return c.client.UpdateAddress(ctx, req)
}

func (c *GRPCUserClient) DeleteAddress(ctx context.Context, req *userv1.DeleteAddressRequest) (*userv1.SuccessResponse, error) {
	return c.client.DeleteAddress(ctx, req)
}

func (c *GRPCUserClient) GetSellerProfile(ctx context.Context, req *userv1.GetSellerProfileRequest) (*userv1.SellerProfile, error) {
	return c.client.GetSellerProfile(ctx, req)
}

func (c *GRPCUserClient) UpdateSellerProfile(ctx context.Context, req *userv1.UpdateSellerProfileRequest) (*userv1.SellerProfile, error) {
	return c.client.UpdateSellerProfile(ctx, req)
}
