package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

func TestMySQLOrderRepositoryListsOnlySellerScopedItemsAndShipments(t *testing.T) {
	db, mock := newMockDB(t)
	repository, _ := NewMySQLOrderRepository(db)
	now := time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC)

	mock.ExpectQuery("FROM orders AS o").
		WithArgs("seller_1", 3).
		WillReturnRows(sqlmock.NewRows([]string{
			"order_id", "status", "currency", "created_at", "seller_items_total",
			"active_item_count", "pending_item_count", "packed_item_count", "shipped_item_count", "delivered_item_count",
		}).AddRow("ord_2", "paid", "INR", now.Add(time.Hour), int64(2000), int64(1), int64(0), int64(1), int64(0), int64(0)).
			AddRow("ord_1", "paid", "INR", now, int64(1000), int64(1), int64(1), int64(0), int64(0), int64(0)))
	mock.ExpectQuery("FROM order_items").
		WithArgs("seller_1", "ord_2", "ord_1").
		WillReturnRows(sqlmock.NewRows([]string{
			"order_item_id", "order_id", "product_id", "variant_id", "sku", "title_snapshot",
			"image_url_snapshot", "quantity", "currency", "unit_amount", "total_amount", "fulfillment_status",
		}).AddRow("oi_2", "ord_2", "product_2", "variant_2", "SKU-2", "Seller Product 2", nil, int32(1), "INR", int64(2000), int64(2000), "packed").
			AddRow("oi_1", "ord_1", "product_1", "variant_1", "SKU-1", "Seller Product 1", nil, int32(1), "INR", int64(1000), int64(1000), "pending"))
	mock.ExpectQuery("FROM shipments").
		WithArgs("seller_1", "ord_2", "ord_1").
		WillReturnRows(sqlmock.NewRows([]string{
			"shipment_id", "order_id", "status", "carrier", "tracking_number", "shipped_at", "delivered_at",
		}).AddRow("shp_2", "ord_2", "packed", nil, nil, nil, nil))

	page, err := repository.ListSellerOrders(context.Background(), domain.ListSellerOrdersFilter{
		SellerID: "seller_1", PageSize: 2,
	})
	if err != nil {
		t.Fatalf("ListSellerOrders() error = %v", err)
	}
	if len(page.Orders) != 2 || page.Orders[0].OrderID != "ord_2" ||
		page.Orders[0].SellerItemsTotal != 2000 || len(page.Orders[0].Items) != 1 ||
		len(page.Orders[0].Shipments) != 1 ||
		page.Orders[1].SellerFulfillmentStatus != domain.SellerFulfillmentStatusPending {
		t.Fatalf("page = %+v, want seller-scoped hydrated views", page)
	}
	assertMockExpectations(t, mock)
}

func TestMySQLOrderRepositoryUpdatesOnlyAuthenticatedSellerFulfillment(t *testing.T) {
	db, mock := newMockDB(t)
	repository, _ := NewMySQLOrderRepository(db)
	now := time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("FROM orders").
		WithArgs("ord_1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("paid"))
	mock.ExpectQuery("FROM order_items").
		WithArgs("ord_1", "seller_1").
		WillReturnRows(sqlmock.NewRows([]string{"fulfillment_status"}).AddRow("pending"))
	mock.ExpectExec("UPDATE order_items").
		WithArgs("packed", "ord_1", "seller_1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("FROM shipments").
		WithArgs("ord_1", "seller_1").
		WillReturnRows(sqlmock.NewRows([]string{"shipment_id"}))
	mock.ExpectExec("INSERT INTO shipments").
		WithArgs("shp_1", "ord_1", "seller_1", nil, nil, "packed", nil, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("FROM order_items").
		WithArgs("ord_1").
		WillReturnRows(sqlmock.NewRows([]string{"fulfillment_status"}).AddRow("packed").AddRow("packed"))
	mock.ExpectExec("UPDATE orders").
		WithArgs("packed", "ord_1", "paid").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO order_status_history").
		WithArgs("osh_1", "ord_1", "paid", "packed", "seller_fulfillment_packed", "seller", "user_1", sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("FROM orders AS o").
		WithArgs("seller_1", "ord_1").
		WillReturnRows(sqlmock.NewRows([]string{
			"order_id", "status", "currency", "created_at", "seller_items_total", "active_item_count",
		}).AddRow("ord_1", "packed", "INR", now, int64(1000), int64(1)))
	mock.ExpectQuery("FROM order_items").
		WithArgs("seller_1", "ord_1").
		WillReturnRows(sqlmock.NewRows([]string{
			"order_item_id", "order_id", "product_id", "variant_id", "sku", "title_snapshot",
			"image_url_snapshot", "quantity", "currency", "unit_amount", "total_amount", "fulfillment_status",
		}).AddRow("oi_1", "ord_1", "product_1", "variant_1", "SKU-1", "Seller Product", nil, int32(1), "INR", int64(1000), int64(1000), "packed"))
	mock.ExpectQuery("FROM shipments").
		WithArgs("seller_1", "ord_1").
		WillReturnRows(sqlmock.NewRows([]string{
			"shipment_id", "order_id", "status", "carrier", "tracking_number", "shipped_at", "delivered_at",
		}).AddRow("shp_1", "ord_1", "packed", nil, nil, nil, nil))

	view, err := repository.UpdateSellerFulfillment(context.Background(), domain.SellerFulfillmentTransition{
		OrderID: "ord_1", SellerID: "seller_1", ShipmentID: "shp_1", TargetStatus: domain.FulfillmentStatusPacked,
		ActorID: "user_1", OccurredAt: now, HistoryID: "osh_1",
	}, nil)
	if err != nil {
		t.Fatalf("UpdateSellerFulfillment() error = %v", err)
	}
	if view.OrderID != "ord_1" || view.ParentOrderStatus != domain.OrderStatusPacked ||
		view.SellerFulfillmentStatus != domain.SellerFulfillmentStatusPacked || len(view.Items) != 1 {
		t.Fatalf("view = %+v, want seller packed projection", view)
	}
	assertMockExpectations(t, mock)
}

func TestMySQLOrderRepositoryEmitsDeliveredEventWhenSellerAggregationDeliversParent(t *testing.T) {
	db, mock := newMockDB(t)
	repository, _ := NewMySQLOrderRepository(db)
	now := time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC)
	event := domain.OutboxEvent{
		EventID: "evt_delivered", EventType: "OrderDelivered", Version: 1,
		AggregateType: "order", AggregateID: "ord_1", RoutingKey: "order.delivered",
		DeduplicationKey: "osh_delivered:OrderDelivered", SourceHistoryID: "osh_delivered",
		Payload: []byte(`{"event_id":"evt_delivered","event_type":"OrderDelivered"}`),
		Status:  domain.OutboxStatusPending, CreatedAt: now, UpdatedAt: now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery("FROM orders").
		WithArgs("ord_1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("shipped"))
	mock.ExpectQuery("FROM order_items").
		WithArgs("ord_1", "seller_1").
		WillReturnRows(sqlmock.NewRows([]string{"fulfillment_status"}).AddRow("shipped"))
	mock.ExpectExec("UPDATE order_items").
		WithArgs("delivered", "ord_1", "seller_1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("FROM shipments").
		WithArgs("ord_1", "seller_1").
		WillReturnRows(sqlmock.NewRows([]string{"shipment_id"}))
	mock.ExpectExec("INSERT INTO shipments").
		WithArgs("shp_1", "ord_1", "seller_1", nil, nil, "delivered", nil, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("FROM order_items").
		WithArgs("ord_1").
		WillReturnRows(sqlmock.NewRows([]string{"fulfillment_status"}).AddRow("delivered").AddRow("delivered"))
	mock.ExpectExec("UPDATE orders").
		WithArgs("delivered", "ord_1", "shipped").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO order_status_history").
		WithArgs("osh_delivered", "ord_1", "shipped", "delivered", "seller_fulfillment_delivered", "seller", "user_1", sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_outbox_events").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("FROM orders AS o").
		WithArgs("seller_1", "ord_1").
		WillReturnRows(sqlmock.NewRows([]string{
			"order_id", "status", "currency", "created_at", "seller_items_total", "active_item_count",
		}).AddRow("ord_1", "delivered", "INR", now, int64(1000), int64(1)))
	mock.ExpectQuery("FROM order_items").
		WithArgs("seller_1", "ord_1").
		WillReturnRows(sqlmock.NewRows([]string{
			"order_item_id", "order_id", "product_id", "variant_id", "sku", "title_snapshot",
			"image_url_snapshot", "quantity", "currency", "unit_amount", "total_amount", "fulfillment_status",
		}).AddRow("oi_1", "ord_1", "product_1", "variant_1", "SKU-1", "Seller Product", nil, int32(1), "INR", int64(1000), int64(1000), "delivered"))
	mock.ExpectQuery("FROM shipments").
		WithArgs("seller_1", "ord_1").
		WillReturnRows(sqlmock.NewRows([]string{
			"shipment_id", "order_id", "status", "carrier", "tracking_number", "shipped_at", "delivered_at",
		}).AddRow("shp_1", "ord_1", "delivered", nil, nil, nil, now))

	view, err := repository.UpdateSellerFulfillment(context.Background(), domain.SellerFulfillmentTransition{
		OrderID: "ord_1", SellerID: "seller_1", ShipmentID: "shp_1", TargetStatus: domain.FulfillmentStatusDelivered,
		ActorID: "user_1", OccurredAt: now, HistoryID: "osh_delivered",
	}, &event)
	if err != nil {
		t.Fatalf("UpdateSellerFulfillment() error = %v", err)
	}
	if view.ParentOrderStatus != domain.OrderStatusDelivered ||
		view.SellerFulfillmentStatus != domain.SellerFulfillmentStatusDelivered {
		t.Fatalf("view = %+v, want delivered parent and seller status", view)
	}
	assertMockExpectations(t, mock)
}
