package clients

import "google.golang.org/grpc"

type NotificationServiceClient interface {
	OutboundClient
	isNotificationServiceClient()
}

type notificationServiceClient struct {
	OutboundClient
}

func newNotificationServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) NotificationServiceClient {
	return notificationServiceClient{OutboundClient: newOutboundClient(descriptor, conn)}
}

func (notificationServiceClient) isNotificationServiceClient() {}
