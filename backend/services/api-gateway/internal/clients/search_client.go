package clients

import (
	searchv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/search/v1"
	"google.golang.org/grpc"
)

type SearchServiceClient interface {
	OutboundClient
	searchv1.SearchServiceClient
	isSearchServiceClient()
}

type searchServiceClient struct {
	OutboundClient
	searchv1.SearchServiceClient
}

func newSearchServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) SearchServiceClient {
	return searchServiceClient{
		OutboundClient:      newOutboundClient(descriptor, conn),
		SearchServiceClient: searchv1.NewSearchServiceClient(conn),
	}
}

func (searchServiceClient) isSearchServiceClient() {}
