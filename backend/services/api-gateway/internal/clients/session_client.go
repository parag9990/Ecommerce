package clients

import "google.golang.org/grpc"

type SessionServiceClient interface {
	OutboundClient
	isSessionServiceClient()
}

type sessionServiceClient struct {
	OutboundClient
}

func newSessionServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) SessionServiceClient {
	return sessionServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (sessionServiceClient) isSessionServiceClient() {}
