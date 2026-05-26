package ordergrpc

import (
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/usecase"
	orderv1 "github.com/example/ecommerce-platform/backend/shared/gen/go/ecommerce/order/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapStatusToProto(value domain.OrderStatus) orderv1.OrderStatus {
	switch value {
	case domain.OrderStatusCreated:
		return orderv1.OrderStatus_ORDER_STATUS_CREATED
	case domain.OrderStatusPendingPayment:
		return orderv1.OrderStatus_ORDER_STATUS_PENDING_PAYMENT
	case domain.OrderStatusPaid:
		return orderv1.OrderStatus_ORDER_STATUS_PAID
	case domain.OrderStatusPacked:
		return orderv1.OrderStatus_ORDER_STATUS_PACKED
	case domain.OrderStatusShipped:
		return orderv1.OrderStatus_ORDER_STATUS_SHIPPED
	case domain.OrderStatusDelivered:
		return orderv1.OrderStatus_ORDER_STATUS_DELIVERED
	case domain.OrderStatusCancelled:
		return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
	case domain.OrderStatusRefunded:
		return orderv1.OrderStatus_ORDER_STATUS_REFUNDED
	case domain.OrderStatusPaymentFailed:
		return orderv1.OrderStatus_ORDER_STATUS_PAYMENT_FAILED
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func mapStatusFromProto(value orderv1.OrderStatus) (domain.OrderStatus, bool) {
	switch value {
	case orderv1.OrderStatus_ORDER_STATUS_CREATED:
		return domain.OrderStatusCreated, true
	case orderv1.OrderStatus_ORDER_STATUS_PENDING_PAYMENT:
		return domain.OrderStatusPendingPayment, true
	case orderv1.OrderStatus_ORDER_STATUS_PAID:
		return domain.OrderStatusPaid, true
	case orderv1.OrderStatus_ORDER_STATUS_PACKED:
		return domain.OrderStatusPacked, true
	case orderv1.OrderStatus_ORDER_STATUS_SHIPPED:
		return domain.OrderStatusShipped, true
	case orderv1.OrderStatus_ORDER_STATUS_DELIVERED:
		return domain.OrderStatusDelivered, true
	case orderv1.OrderStatus_ORDER_STATUS_CANCELLED:
		return domain.OrderStatusCancelled, true
	case orderv1.OrderStatus_ORDER_STATUS_REFUNDED:
		return domain.OrderStatusRefunded, true
	case orderv1.OrderStatus_ORDER_STATUS_PAYMENT_FAILED:
		return domain.OrderStatusPaymentFailed, true
	default:
		return "", false
	}
}

func mapOrder(order domain.Order) *orderv1.Order {
	items := make([]*orderv1.OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, &orderv1.OrderItem{
			OrderItemId: item.OrderItemID,
			ProductId:   item.ProductID,
			VariantId:   item.VariantID,
			SellerId:    item.SellerID,
			ProductName: item.TitleSnapshot,
			Quantity:    item.Quantity,
			UnitPrice:   mapMoney(item.UnitAmount, item.Currency),
			LineTotal:   mapMoney(item.TotalAmount, item.Currency),
		})
	}
	result := &orderv1.Order{
		OrderId:   order.OrderID,
		UserId:    order.UserID,
		Status:    mapStatusToProto(order.Status),
		Items:     items,
		Total:     mapMoney(order.TotalAmount, order.Currency),
		PaymentId: order.PaymentID,
		ShippingAddress: &orderv1.AddressSnapshot{
			RecipientName: order.AddressSnapshot.RecipientName,
			Phone:         order.AddressSnapshot.Phone,
			Line1:         order.AddressSnapshot.Line1,
			Line2:         order.AddressSnapshot.Line2,
			City:          order.AddressSnapshot.City,
			State:         order.AddressSnapshot.State,
			PostalCode:    order.AddressSnapshot.PostalCode,
			CountryCode:   order.AddressSnapshot.Country,
		},
	}
	if !order.CreatedAt.IsZero() {
		result.CreatedAt = timestamppb.New(order.CreatedAt)
	}
	if !order.UpdatedAt.IsZero() {
		result.UpdatedAt = timestamppb.New(order.UpdatedAt)
	}
	return result
}

func mapOrders(orders []domain.Order) []*orderv1.Order {
	result := make([]*orderv1.Order, 0, len(orders))
	for _, order := range orders {
		result = append(result, mapOrder(order))
	}
	return result
}

func mapMoney(amount int64, currency string) *orderv1.Money {
	return &orderv1.Money{MinorUnits: amount, Currency: currency}
}

func mapPaymentAction(action *usecase.PaymentActionView) *orderv1.PaymentAction {
	if action == nil {
		return nil
	}
	result := &orderv1.PaymentAction{
		PaymentId:         action.PaymentID,
		ClientActionToken: action.ClientActionToken,
	}
	if !action.ExpiresAt.IsZero() {
		result.ExpiresAt = timestamppb.New(action.ExpiresAt)
	}
	return result
}
