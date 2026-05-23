package clients

import "google.golang.org/grpc"

type PaymentServiceClient interface {
	OutboundClient
	isPaymentServiceClient()
}

type paymentServiceClient struct {
	OutboundClient
}

func newPaymentServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) PaymentServiceClient {
	return paymentServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (paymentServiceClient) isPaymentServiceClient() {}
