package clients

import "google.golang.org/grpc"

type OrderServiceClient interface {
	OutboundClient
	isOrderServiceClient()
}

type orderServiceClient struct {
	OutboundClient
}

func newOrderServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) OrderServiceClient {
	return orderServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (orderServiceClient) isOrderServiceClient() {}
