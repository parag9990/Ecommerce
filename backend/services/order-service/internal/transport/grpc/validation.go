package ordergrpc

import (
	"strings"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/usecase"
	orderv1 "github.com/example/ecommerce-platform/backend/shared/gen/go/ecommerce/order/v1"
)

func validateCreateOrderRequest(request *orderv1.CreateOrderRequest) error {
	if request == nil || strings.TrimSpace(request.GetCartId()) == "" || request.GetShippingAddress() == nil {
		return domain.ErrInvalidRequest
	}
	if err := domain.ValidateIdempotencyKey(request.GetIdempotencyKey()); err != nil {
		return err
	}
	address := domain.AddressSnapshot{
		RecipientName: request.GetShippingAddress().GetRecipientName(),
		Phone:         request.GetShippingAddress().GetPhone(),
		Line1:         request.GetShippingAddress().GetLine1(),
		Line2:         request.GetShippingAddress().GetLine2(),
		City:          request.GetShippingAddress().GetCity(),
		State:         request.GetShippingAddress().GetState(),
		PostalCode:    request.GetShippingAddress().GetPostalCode(),
		Country:       request.GetShippingAddress().GetCountryCode(),
	}
	if err := address.Validate(); err != nil {
		return err
	}
	return nil
}

func validateListOrdersRequest(request *orderv1.ListOrdersRequest) (int, *domain.OrderStatus, string, error) {
	if request == nil {
		return usecase.DefaultOrderPageSize, nil, "", nil
	}
	pageSize := int(request.GetPageSize())
	if pageSize == 0 {
		pageSize = usecase.DefaultOrderPageSize
	}
	if pageSize < 1 || pageSize > usecase.MaxOrderPageSize {
		return 0, nil, "", domain.ErrInvalidRequest
	}
	var filter *domain.OrderStatus
	if request.GetStatusFilter() != orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		status, ok := mapStatusFromProto(request.GetStatusFilter())
		if !ok {
			return 0, nil, "", domain.ErrInvalidRequest
		}
		filter = &status
	}
	return pageSize, filter, request.GetPageToken(), nil
}

func validateFulfillmentRequest(request *orderv1.UpdateFulfillmentRequest) (domain.OrderStatus, error) {
	if request == nil || strings.TrimSpace(request.GetOrderId()) == "" {
		return "", domain.ErrInvalidRequest
	}
	status, ok := mapStatusFromProto(request.GetTargetStatus())
	if !ok || !domain.IsFulfillmentTargetStatus(status) {
		return "", domain.ErrInvalidRequest
	}
	if status == domain.OrderStatusShipped &&
		(strings.TrimSpace(request.GetTrackingNumber()) == "" || strings.TrimSpace(request.GetCarrier()) == "") {
		return "", domain.ErrTrackingRequired
	}
	return status, nil
}
