package clients

import "google.golang.org/grpc"

type OutboundClient interface {
	grpc.ClientConnInterface
	Descriptor() ServiceDescriptor
	Conn() grpc.ClientConnInterface
}

type downstreamClient struct {
	descriptor ServiceDescriptor
	grpc.ClientConnInterface
}

func newOutboundClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) OutboundClient {
	return &downstreamClient{
		descriptor:          descriptor,
		ClientConnInterface: conn,
	}
}

func (c *downstreamClient) Descriptor() ServiceDescriptor {
	return c.descriptor
}

func (c *downstreamClient) Conn() grpc.ClientConnInterface {
	return c.ClientConnInterface
}
