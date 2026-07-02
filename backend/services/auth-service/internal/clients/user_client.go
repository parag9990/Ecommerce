package clients

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	userv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type UserRPC interface {
	CreateUser(ctx context.Context, in *userv1.CreateUserRequest, opts ...grpc.CallOption) (*userv1.UserProfile, error)
}

type GRPCUserClient struct {
	client  UserRPC
	timeout time.Duration
}

func DialGRPCUserClient(address string, timeout time.Duration) (*GRPCUserClient, *grpc.ClientConn, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, nil, errors.New("user gRPC address is required")
	}
	if timeout <= 0 {
		return nil, nil, errors.New("user timeout must be greater than zero")
	}

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("create user gRPC client: %w", err)
	}
	return NewGRPCUserClient(userv1.NewUserServiceClient(conn), timeout), conn, nil
}

func NewGRPCUserClient(client UserRPC, timeout time.Duration) *GRPCUserClient {
	return &GRPCUserClient{client: client, timeout: timeout}
}

func (c *GRPCUserClient) CreateUser(ctx context.Context, authAccountID string, email string, phone string, fullName string, requestID string) (string, error) {
	if c == nil || c.client == nil {
		return "", errors.New("user client is not initialized")
	}
	if c.timeout <= 0 {
		return "", errors.New("user timeout must be greater than zero")
	}
	authAccountID = strings.TrimSpace(authAccountID)
	email = strings.TrimSpace(email)
	fullName = strings.TrimSpace(fullName)
	if authAccountID == "" || email == "" || fullName == "" {
		return "", errors.New("user create request is incomplete")
	}

	md := []string{
		"x-service-name", "auth-service",
		"x-actor-id", "auth-service",
		"x-actor-type", "service",
	}
	if requestID = strings.TrimSpace(requestID); requestID != "" {
		md = append(md, "x-request-id", requestID)
	}

	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	callCtx = metadata.AppendToOutgoingContext(callCtx, md...)

	response, err := c.client.CreateUser(callCtx, &userv1.CreateUserRequest{
		AuthAccountId: authAccountID,
		Email:         email,
		Phone:         strings.TrimSpace(phone),
		FullName:      fullName,
	})
	if err != nil {
		return "", fmt.Errorf("create user profile: %w", err)
	}
	if response == nil || strings.TrimSpace(response.GetUserId()) == "" {
		return "", errors.New("user create response is incomplete")
	}
	return strings.TrimSpace(response.GetUserId()), nil
}
