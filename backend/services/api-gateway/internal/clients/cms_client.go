package clients

import "google.golang.org/grpc"

type CMSServiceClient interface {
	OutboundClient
	isCMSServiceClient()
}

type cmsServiceClient struct {
	OutboundClient
}

func newCMSServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) CMSServiceClient {
	return cmsServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (cmsServiceClient) isCMSServiceClient() {}
