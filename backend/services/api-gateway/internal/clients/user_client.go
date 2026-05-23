package clients

import "google.golang.org/grpc"

type UserServiceClient interface {
	OutboundClient
	isUserServiceClient()
}

type userServiceClient struct {
	OutboundClient
}

func newUserServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) UserServiceClient {
	return userServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (userServiceClient) isUserServiceClient() {}
