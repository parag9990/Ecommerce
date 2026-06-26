package clients

import "google.golang.org/grpc"

// RecommendationServiceClient stays transport-agnostic because the gRPC-Web
// facade forwards the canonical protobuf frames without translating messages.
type RecommendationServiceClient interface {
	OutboundClient
	isRecommendationServiceClient()
}

type recommendationServiceClient struct {
	OutboundClient
}

func newRecommendationServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) RecommendationServiceClient {
	return recommendationServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (recommendationServiceClient) isRecommendationServiceClient() {}
