package clients

import "google.golang.org/grpc"

type ProductServiceClient interface {
	OutboundClient
	isProductServiceClient()
}

type productServiceClient struct {
	OutboundClient
}

func newProductServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) ProductServiceClient {
	return productServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (productServiceClient) isProductServiceClient() {}
