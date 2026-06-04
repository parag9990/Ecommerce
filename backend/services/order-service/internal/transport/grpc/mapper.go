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

func mapSellerFulfillmentStatusToProto(value domain.SellerFulfillmentStatus) orderv1.SellerFulfillmentStatus {
	switch value {
	case domain.SellerFulfillmentStatusPending:
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PENDING
	case domain.SellerFulfillmentStatusPartiallyPacked:
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_PACKED
	case domain.SellerFulfillmentStatusPacked:
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PACKED
	case domain.SellerFulfillmentStatusPartiallyShipped:
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_SHIPPED
	case domain.SellerFulfillmentStatusShipped:
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_SHIPPED
	case domain.SellerFulfillmentStatusPartiallyDelivered:
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_DELIVERED
	case domain.SellerFulfillmentStatusDelivered:
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_DELIVERED
	default:
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_UNSPECIFIED
	}
}

func mapSellerFulfillmentStatusFromProto(value orderv1.SellerFulfillmentStatus) (*domain.SellerFulfillmentStatus, bool) {
	switch value {
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_UNSPECIFIED:
		return nil, true
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PENDING:
		status := domain.SellerFulfillmentStatusPending
		return &status, true
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_PACKED:
		status := domain.SellerFulfillmentStatusPartiallyPacked
		return &status, true
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PACKED:
		status := domain.SellerFulfillmentStatusPacked
		return &status, true
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_SHIPPED:
		status := domain.SellerFulfillmentStatusPartiallyShipped
		return &status, true
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_SHIPPED:
		status := domain.SellerFulfillmentStatusShipped
		return &status, true
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_DELIVERED:
		status := domain.SellerFulfillmentStatusPartiallyDelivered
		return &status, true
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_DELIVERED:
		status := domain.SellerFulfillmentStatusDelivered
		return &status, true
	default:
		return nil, false
	}
}

func mapItemFulfillmentStatusToProto(value domain.FulfillmentStatus) orderv1.ItemFulfillmentStatus {
	switch value {
	case domain.FulfillmentStatusPending:
		return orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_PENDING
	case domain.FulfillmentStatusPacked:
		return orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_PACKED
	case domain.FulfillmentStatusShipped:
		return orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_SHIPPED
	case domain.FulfillmentStatusDelivered:
		return orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_DELIVERED
	case domain.FulfillmentStatusCancelled:
		return orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_CANCELLED
	case domain.FulfillmentStatusReturned:
		return orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_RETURNED
	default:
		return orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_UNSPECIFIED
	}
}

func mapShipmentStatusToProto(value domain.ShipmentStatus) orderv1.ShipmentStatus {
	switch value {
	case domain.ShipmentStatusPending:
		return orderv1.ShipmentStatus_SHIPMENT_STATUS_PENDING
	case domain.ShipmentStatusPacked:
		return orderv1.ShipmentStatus_SHIPMENT_STATUS_PACKED
	case domain.ShipmentStatusShipped:
		return orderv1.ShipmentStatus_SHIPMENT_STATUS_SHIPPED
	case domain.ShipmentStatusDelivered:
		return orderv1.ShipmentStatus_SHIPMENT_STATUS_DELIVERED
	case domain.ShipmentStatusFailed:
		return orderv1.ShipmentStatus_SHIPMENT_STATUS_FAILED
	default:
		return orderv1.ShipmentStatus_SHIPMENT_STATUS_UNSPECIFIED
	}
}

func mapSellerOrderView(order domain.SellerOrderView) *orderv1.SellerOrderView {
	items := make([]*orderv1.SellerOrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, &orderv1.SellerOrderItem{
			OrderItemId:       item.OrderItemID,
			ProductId:         item.ProductID,
			VariantId:         item.VariantID,
			Sku:               item.SKU,
			TitleSnapshot:     item.TitleSnapshot,
			ImageUrlSnapshot:  item.ImageURLSnapshot,
			Quantity:          item.Quantity,
			UnitPrice:         mapMoney(item.UnitAmount, item.Currency),
			LineTotal:         mapMoney(item.TotalAmount, item.Currency),
			FulfillmentStatus: mapItemFulfillmentStatusToProto(item.FulfillmentStatus),
		})
	}
	shipments := make([]*orderv1.SellerShipment, 0, len(order.Shipments))
	for _, shipment := range order.Shipments {
		mapped := &orderv1.SellerShipment{
			ShipmentId:     shipment.ShipmentID,
			Status:         mapShipmentStatusToProto(shipment.Status),
			Carrier:        shipment.Carrier,
			TrackingNumber: shipment.TrackingNumber,
		}
		if shipment.ShippedAt != nil && !shipment.ShippedAt.IsZero() {
			mapped.ShippedAt = timestamppb.New(*shipment.ShippedAt)
		}
		if shipment.DeliveredAt != nil && !shipment.DeliveredAt.IsZero() {
			mapped.DeliveredAt = timestamppb.New(*shipment.DeliveredAt)
		}
		shipments = append(shipments, mapped)
	}
	result := &orderv1.SellerOrderView{
		OrderId:                 order.OrderID,
		ParentOrderStatus:       mapStatusToProto(order.ParentOrderStatus),
		SellerFulfillmentStatus: mapSellerFulfillmentStatusToProto(order.SellerFulfillmentStatus),
		SellerItemsTotal:        mapMoney(order.SellerItemsTotal, order.Currency),
		Items:                   items,
		Shipments:               shipments,
	}
	if !order.CreatedAt.IsZero() {
		result.CreatedAt = timestamppb.New(order.CreatedAt)
	}
	return result
}

func mapSellerOrderViews(orders []domain.SellerOrderView) []*orderv1.SellerOrderView {
	result := make([]*orderv1.SellerOrderView, 0, len(orders))
	for _, order := range orders {
		result = append(result, mapSellerOrderView(order))
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
