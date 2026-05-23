package clients

import "google.golang.org/grpc"

type SearchServiceClient interface {
	OutboundClient
	isSearchServiceClient()
}

type searchServiceClient struct {
	OutboundClient
}

func newSearchServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) SearchServiceClient {
	return searchServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (searchServiceClient) isSearchServiceClient() {}
