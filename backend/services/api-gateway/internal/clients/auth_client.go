package clients

import "google.golang.org/grpc"

type AuthServiceClient interface {
	OutboundClient
	isAuthServiceClient()
}

type authServiceClient struct {
	OutboundClient
}

func newAuthServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) AuthServiceClient {
	return authServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (authServiceClient) isAuthServiceClient() {}
