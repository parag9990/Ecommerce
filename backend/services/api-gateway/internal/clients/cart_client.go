package clients

import "google.golang.org/grpc"

type CartServiceClient interface {
	OutboundClient
	isCartServiceClient()
}

type cartServiceClient struct {
	OutboundClient
}

func newCartServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) CartServiceClient {
	return cartServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (cartServiceClient) isCartServiceClient() {}
