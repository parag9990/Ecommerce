package clients

import (
	"context"

	notificationv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/notification/v1"
	"google.golang.org/grpc"
)

type NotificationClient interface {
	GetNotificationPreference(context.Context, *notificationv1.GetNotificationPreferenceRequest) (*notificationv1.NotificationPreference, error)
	UpdateNotificationPreference(context.Context, *notificationv1.UpdateNotificationPreferenceRequest) (*notificationv1.NotificationPreference, error)
}

type NotificationServiceClient interface {
	OutboundClient
	NotificationClient
	isNotificationServiceClient()
}

type notificationServiceClient struct {
	OutboundClient
	rpc notificationv1.NotificationServiceClient
}

func newNotificationServiceClient(descriptor ServiceDescriptor, conn grpc.ClientConnInterface) NotificationServiceClient {
	return notificationServiceClient{
		OutboundClient: newOutboundClient(descriptor, conn),
		rpc:            notificationv1.NewNotificationServiceClient(conn),
	}
}

func (notificationServiceClient) isNotificationServiceClient() {}

func (c notificationServiceClient) GetNotificationPreference(
	ctx context.Context,
	request *notificationv1.GetNotificationPreferenceRequest,
) (*notificationv1.NotificationPreference, error) {
	return c.rpc.GetNotificationPreference(ctx, request)
}

func (c notificationServiceClient) UpdateNotificationPreference(
	ctx context.Context,
	request *notificationv1.UpdateNotificationPreferenceRequest,
) (*notificationv1.NotificationPreference, error) {
	return c.rpc.UpdateNotificationPreference(ctx, request)
}
