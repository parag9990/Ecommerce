package clients

import "google.golang.org/grpc"

type SuperadminServiceClient interface {
	OutboundClient
	isSuperadminServiceClient()
}

type superadminServiceClient struct {
	OutboundClient
}

func newSuperadminServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) SuperadminServiceClient {
	return superadminServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (superadminServiceClient) isSuperadminServiceClient() {}
