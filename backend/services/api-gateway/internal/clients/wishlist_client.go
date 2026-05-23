package clients

import "google.golang.org/grpc"

type WishlistServiceClient interface {
	OutboundClient
	isWishlistServiceClient()
}

type wishlistServiceClient struct {
	OutboundClient
}

func newWishlistServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) WishlistServiceClient {
	return wishlistServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (wishlistServiceClient) isWishlistServiceClient() {}
