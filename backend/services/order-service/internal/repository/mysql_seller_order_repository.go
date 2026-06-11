package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

type sellerOrderSummary struct {
	OrderID           string
	ParentOrderStatus domain.OrderStatus
	Currency          string
	SellerItemsTotal  int64
	CreatedAt         time.Time
}

func (r *MySQLOrderRepository) ListSellerOrders(ctx context.Context, filter domain.ListSellerOrdersFilter) (domain.SellerOrderPage, error) {
	filter.SellerID = strings.TrimSpace(filter.SellerID)
	if filter.SellerID == "" || filter.PageSize <= 0 {
		return domain.SellerOrderPage{}, domain.ErrInvalidRequest
	}
	summaries, cursor, err := r.listSellerOrderSummaries(ctx, filter)
	if err != nil {
		return domain.SellerOrderPage{}, err
	}
	if len(summaries) == 0 {
		return domain.SellerOrderPage{Orders: []domain.SellerOrderView{}}, nil
	}
	views, err := r.hydrateSellerOrderViews(ctx, filter.SellerID, summaries)
	if err != nil {
		return domain.SellerOrderPage{}, err
	}
	return domain.SellerOrderPage{Orders: views, NextCursor: cursor}, nil
}

func (r *MySQLOrderRepository) UpdateSellerFulfillment(
	ctx context.Context,
	transition domain.SellerFulfillmentTransition,
	event *domain.OutboxEvent,
) (domain.SellerOrderView, error) {
	transition.OrderID = strings.TrimSpace(transition.OrderID)
	transition.SellerID = strings.TrimSpace(transition.SellerID)
	transition.ActorID = strings.TrimSpace(transition.ActorID)
	transition.Carrier = strings.TrimSpace(transition.Carrier)
	transition.TrackingNumber = strings.TrimSpace(transition.TrackingNumber)
	transition.Note = strings.TrimSpace(transition.Note)
	if err := transition.Validate(); err != nil {
		return domain.SellerOrderView{}, err
	}
	if event != nil {
		if err := validateOutboxEventForHistory(*event, transition.OrderID, transition.HistoryID); err != nil {
			return domain.SellerOrderView{}, err
		}
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return domain.SellerOrderView{}, databaseUnavailable("begin seller fulfillment transaction", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	current, err := lockParentOrderStatus(ctx, tx, transition.OrderID)
	if err != nil {
		return domain.SellerOrderView{}, err
	}
	if !domain.CanSellerFulfillParentStatus(current) {
		return domain.SellerOrderView{}, domain.ErrOrderNotPaid
	}
	ownStatuses, err := lockSellerItemStatuses(ctx, tx, transition.OrderID, transition.SellerID)
	if err != nil {
		return domain.SellerOrderView{}, err
	}
	for _, status := range ownStatuses {
		if err := domain.ValidateSellerItemTransition(status, transition.TargetStatus); err != nil {
			return domain.SellerOrderView{}, err
		}
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE order_items
		SET fulfillment_status = ?
		WHERE order_id = ?
		  AND seller_id = ?
	`, string(transition.TargetStatus), transition.OrderID, transition.SellerID)
	if err != nil {
		return domain.SellerOrderView{}, databaseUnavailable("update seller item fulfillment statuses", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.SellerOrderView{}, databaseUnavailable("read seller item fulfillment update result", err)
	}
	if affected != int64(len(ownStatuses)) {
		return domain.SellerOrderView{}, domain.ErrInvalidFulfillmentTransition
	}
	if err := persistSellerShipmentTransition(ctx, tx, transition); err != nil {
		return domain.SellerOrderView{}, err
	}
	allStatuses, err := lockOrderItemStatuses(ctx, tx, transition.OrderID)
	if err != nil {
		return domain.SellerOrderView{}, err
	}
	nextStatus, changed := domain.AggregateParentFulfillmentStatus(current, allStatuses)
	if changed {
		if err := persistAggregatedParentStatus(ctx, tx, transition, current, nextStatus); err != nil {
			return domain.SellerOrderView{}, err
		}
		if nextStatus == domain.OrderStatusDelivered {
			if event == nil {
				return domain.SellerOrderView{}, domain.ErrInvalidOrder
			}
			if err := insertOutboxEvent(ctx, tx, *event); err != nil {
				return domain.SellerOrderView{}, databaseUnavailable("insert seller fulfillment delivered outbox event", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.SellerOrderView{}, databaseUnavailable("commit seller fulfillment transaction", err)
	}
	return r.getSellerOrderView(ctx, transition.SellerID, transition.OrderID)
}

func (r *MySQLOrderRepository) listSellerOrderSummaries(ctx context.Context, filter domain.ListSellerOrdersFilter) ([]sellerOrderSummary, *domain.OrderCursor, error) {
	query := `
		SELECT
			o.order_id,
			o.status,
			o.currency,
			o.created_at,
			COALESCE(SUM(oi.total_amount), 0) AS seller_items_total,
			SUM(CASE WHEN oi.fulfillment_status NOT IN ('cancelled', 'returned') THEN 1 ELSE 0 END) AS active_item_count,
			SUM(CASE WHEN oi.fulfillment_status = 'pending' THEN 1 ELSE 0 END) AS pending_item_count,
			SUM(CASE WHEN oi.fulfillment_status = 'packed' THEN 1 ELSE 0 END) AS packed_item_count,
			SUM(CASE WHEN oi.fulfillment_status = 'shipped' THEN 1 ELSE 0 END) AS shipped_item_count,
			SUM(CASE WHEN oi.fulfillment_status = 'delivered' THEN 1 ELSE 0 END) AS delivered_item_count
		FROM orders AS o
		INNER JOIN order_items AS oi
			ON oi.order_id = o.order_id
		WHERE oi.seller_id = ?`
	args := []any{filter.SellerID}
	if filter.Cursor != nil {
		query += `
		  AND (o.created_at < ? OR (o.created_at = ? AND o.order_id < ?))`
		args = append(args, filter.Cursor.CreatedAt.UTC(), filter.Cursor.CreatedAt.UTC(), filter.Cursor.OrderID)
	}
	query += `
		GROUP BY o.order_id, o.status, o.currency, o.created_at
		HAVING active_item_count > 0`
	if filter.FulfillmentFilter != nil {
		query += sellerFulfillmentHavingClause(*filter.FulfillmentFilter)
	}
	query += `
		ORDER BY o.created_at DESC, o.order_id DESC
		LIMIT ?`
	args = append(args, filter.PageSize+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, databaseUnavailable("list seller order summaries", err)
	}
	defer rows.Close()

	summaries := make([]sellerOrderSummary, 0, filter.PageSize+1)
	for rows.Next() {
		var (
			summary        sellerOrderSummary
			status         string
			activeCount    int64
			pendingCount   int64
			packedCount    int64
			shippedCount   int64
			deliveredCount int64
		)
		if err := rows.Scan(
			&summary.OrderID,
			&status,
			&summary.Currency,
			&summary.CreatedAt,
			&summary.SellerItemsTotal,
			&activeCount,
			&pendingCount,
			&packedCount,
			&shippedCount,
			&deliveredCount,
		); err != nil {
			return nil, nil, fmt.Errorf("scan seller order summary: %w", err)
		}
		parsed, err := domain.ParseOrderStatus(status)
		if err != nil {
			return nil, nil, err
		}
		summary.ParentOrderStatus = parsed
		summary.CreatedAt = summary.CreatedAt.UTC()
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, databaseUnavailable("iterate seller order summaries", err)
	}

	var cursor *domain.OrderCursor
	if len(summaries) > filter.PageSize {
		last := summaries[filter.PageSize-1]
		cursor = &domain.OrderCursor{CreatedAt: last.CreatedAt, OrderID: last.OrderID}
		summaries = summaries[:filter.PageSize]
	}
	return summaries, cursor, nil
}

func sellerFulfillmentHavingClause(filter domain.SellerFulfillmentStatus) string {
	switch filter {
	case domain.SellerFulfillmentStatusPending:
		return " AND pending_item_count = active_item_count"
	case domain.SellerFulfillmentStatusPartiallyPacked:
		return " AND packed_item_count > 0 AND pending_item_count > 0 AND shipped_item_count = 0 AND delivered_item_count = 0"
	case domain.SellerFulfillmentStatusPacked:
		return " AND packed_item_count = active_item_count"
	case domain.SellerFulfillmentStatusPartiallyShipped:
		return " AND shipped_item_count > 0 AND delivered_item_count = 0 AND (pending_item_count + packed_item_count) > 0"
	case domain.SellerFulfillmentStatusShipped:
		return " AND shipped_item_count = active_item_count"
	case domain.SellerFulfillmentStatusPartiallyDelivered:
		return " AND delivered_item_count > 0 AND delivered_item_count < active_item_count"
	case domain.SellerFulfillmentStatusDelivered:
		return " AND delivered_item_count = active_item_count"
	default:
		return ""
	}
}

func (r *MySQLOrderRepository) getSellerOrderView(ctx context.Context, sellerID string, orderID string) (domain.SellerOrderView, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			o.order_id,
			o.status,
			o.currency,
			o.created_at,
			COALESCE(SUM(oi.total_amount), 0) AS seller_items_total,
			SUM(CASE WHEN oi.fulfillment_status NOT IN ('cancelled', 'returned') THEN 1 ELSE 0 END) AS active_item_count
		FROM orders AS o
		INNER JOIN order_items AS oi
			ON oi.order_id = o.order_id
		WHERE oi.seller_id = ?
		  AND o.order_id = ?
		GROUP BY o.order_id, o.status, o.currency, o.created_at
		HAVING active_item_count > 0
	`, strings.TrimSpace(sellerID), strings.TrimSpace(orderID))
	if err != nil {
		return domain.SellerOrderView{}, databaseUnavailable("get seller order summary", err)
	}
	defer rows.Close()

	var summaries []sellerOrderSummary
	for rows.Next() {
		var (
			summary     sellerOrderSummary
			status      string
			activeCount int64
		)
		if err := rows.Scan(
			&summary.OrderID,
			&status,
			&summary.Currency,
			&summary.CreatedAt,
			&summary.SellerItemsTotal,
			&activeCount,
		); err != nil {
			return domain.SellerOrderView{}, fmt.Errorf("scan seller order summary: %w", err)
		}
		parsed, err := domain.ParseOrderStatus(status)
		if err != nil {
			return domain.SellerOrderView{}, err
		}
		summary.ParentOrderStatus = parsed
		summary.CreatedAt = summary.CreatedAt.UTC()
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return domain.SellerOrderView{}, databaseUnavailable("iterate seller order summary", err)
	}
	if len(summaries) == 0 {
		return domain.SellerOrderView{}, domain.ErrOrderNotFound
	}
	views, err := r.hydrateSellerOrderViews(ctx, sellerID, summaries)
	if err != nil {
		return domain.SellerOrderView{}, err
	}
	return views[0], nil
}

func (r *MySQLOrderRepository) hydrateSellerOrderViews(ctx context.Context, sellerID string, summaries []sellerOrderSummary) ([]domain.SellerOrderView, error) {
	orderIDs := make([]string, 0, len(summaries))
	index := make(map[string]int, len(summaries))
	views := make([]domain.SellerOrderView, 0, len(summaries))
	for position, summary := range summaries {
		orderIDs = append(orderIDs, summary.OrderID)
		index[summary.OrderID] = position
		views = append(views, domain.SellerOrderView{
			OrderID:           summary.OrderID,
			ParentOrderStatus: summary.ParentOrderStatus,
			Currency:          summary.Currency,
			SellerItemsTotal:  summary.SellerItemsTotal,
			Items:             []domain.SellerOrderItemView{},
			Shipments:         []domain.SellerShipmentView{},
			CreatedAt:         summary.CreatedAt,
		})
	}
	items, err := r.listSellerItemsForOrders(ctx, sellerID, orderIDs)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		position, exists := index[item.OrderID]
		if !exists {
			return nil, errors.New("seller item returned for an unrequested order")
		}
		views[position].Items = append(views[position].Items, item)
	}
	shipments, err := r.listSellerShipmentsForOrders(ctx, sellerID, orderIDs)
	if err != nil {
		return nil, err
	}
	for _, shipment := range shipments {
		position, exists := index[shipment.OrderID]
		if !exists {
			return nil, errors.New("seller shipment returned for an unrequested order")
		}
		views[position].Shipments = append(views[position].Shipments, shipment)
	}
	for position := range views {
		views[position].SellerFulfillmentStatus = domain.DeriveSellerFulfillmentStatus(views[position].Items)
	}
	return views, nil
}

func (r *MySQLOrderRepository) listSellerItemsForOrders(ctx context.Context, sellerID string, orderIDs []string) ([]domain.SellerOrderItemView, error) {
	if len(orderIDs) == 0 {
		return nil, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(orderIDs)), ",")
	args := make([]any, 0, len(orderIDs)+1)
	args = append(args, strings.TrimSpace(sellerID))
	for _, orderID := range orderIDs {
		args = append(args, orderID)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			order_item_id,
			order_id,
			product_id,
			variant_id,
			sku,
			title_snapshot,
			image_url_snapshot,
			quantity,
			currency,
			unit_amount,
			total_amount,
			fulfillment_status
		FROM order_items
		WHERE seller_id = ?
		  AND order_id IN (`+placeholders+`)
		ORDER BY order_id, created_at, order_item_id
	`, args...)
	if err != nil {
		return nil, databaseUnavailable("list seller order items", err)
	}
	defer rows.Close()

	var items []domain.SellerOrderItemView
	for rows.Next() {
		var (
			item     domain.SellerOrderItemView
			imageURL sql.NullString
			status   string
		)
		if err := rows.Scan(
			&item.OrderItemID,
			&item.OrderID,
			&item.ProductID,
			&item.VariantID,
			&item.SKU,
			&item.TitleSnapshot,
			&imageURL,
			&item.Quantity,
			&item.Currency,
			&item.UnitAmount,
			&item.TotalAmount,
			&status,
		); err != nil {
			return nil, fmt.Errorf("scan seller order item: %w", err)
		}
		if imageURL.Valid {
			item.ImageURLSnapshot = imageURL.String
		}
		item.FulfillmentStatus = domain.FulfillmentStatus(status)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, databaseUnavailable("iterate seller order items", err)
	}
	return items, nil
}

func (r *MySQLOrderRepository) listSellerShipmentsForOrders(ctx context.Context, sellerID string, orderIDs []string) ([]domain.SellerShipmentView, error) {
	if len(orderIDs) == 0 {
		return nil, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(orderIDs)), ",")
	args := make([]any, 0, len(orderIDs)+1)
	args = append(args, strings.TrimSpace(sellerID))
	for _, orderID := range orderIDs {
		args = append(args, orderID)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			shipment_id,
			order_id,
			status,
			carrier,
			tracking_number,
			shipped_at,
			delivered_at
		FROM shipments
		WHERE seller_id = ?
		  AND order_id IN (`+placeholders+`)
		ORDER BY order_id, created_at, shipment_id
	`, args...)
	if err != nil {
		return nil, databaseUnavailable("list seller shipments", err)
	}
	defer rows.Close()

	var shipments []domain.SellerShipmentView
	for rows.Next() {
		var (
			shipment       domain.SellerShipmentView
			status         string
			carrier        sql.NullString
			trackingNumber sql.NullString
			shippedAt      sql.NullTime
			deliveredAt    sql.NullTime
		)
		if err := rows.Scan(
			&shipment.ShipmentID,
			&shipment.OrderID,
			&status,
			&carrier,
			&trackingNumber,
			&shippedAt,
			&deliveredAt,
		); err != nil {
			return nil, fmt.Errorf("scan seller shipment: %w", err)
		}
		shipment.Status = domain.ShipmentStatus(status)
		if carrier.Valid {
			shipment.Carrier = carrier.String
		}
		if trackingNumber.Valid {
			shipment.TrackingNumber = trackingNumber.String
		}
		if shippedAt.Valid {
			value := shippedAt.Time.UTC()
			shipment.ShippedAt = &value
		}
		if deliveredAt.Valid {
			value := deliveredAt.Time.UTC()
			shipment.DeliveredAt = &value
		}
		shipments = append(shipments, shipment)
	}
	if err := rows.Err(); err != nil {
		return nil, databaseUnavailable("iterate seller shipments", err)
	}
	return shipments, nil
}

func lockParentOrderStatus(ctx context.Context, tx *sql.Tx, orderID string) (domain.OrderStatus, error) {
	var statusValue string
	err := tx.QueryRowContext(ctx, `
		SELECT status
		FROM orders
		WHERE order_id = ?
		FOR UPDATE
	`, strings.TrimSpace(orderID)).Scan(&statusValue)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrOrderNotFound
	}
	if err != nil {
		return "", databaseUnavailable("lock parent order for seller fulfillment", err)
	}
	status, err := domain.ParseOrderStatus(statusValue)
	if err != nil {
		return "", err
	}
	return status, nil
}

func lockSellerItemStatuses(ctx context.Context, tx *sql.Tx, orderID string, sellerID string) ([]domain.FulfillmentStatus, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT fulfillment_status
		FROM order_items
		WHERE order_id = ?
		  AND seller_id = ?
		FOR UPDATE
	`, strings.TrimSpace(orderID), strings.TrimSpace(sellerID))
	if err != nil {
		return nil, databaseUnavailable("lock seller order items", err)
	}
	defer rows.Close()

	var statuses []domain.FulfillmentStatus
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return nil, databaseUnavailable("scan seller item status", err)
		}
		statuses = append(statuses, domain.FulfillmentStatus(status))
	}
	if err := rows.Err(); err != nil {
		return nil, databaseUnavailable("iterate seller item statuses", err)
	}
	if len(statuses) == 0 {
		return nil, domain.ErrOrderNotFound
	}
	return statuses, nil
}

func lockOrderItemStatuses(ctx context.Context, tx *sql.Tx, orderID string) ([]domain.FulfillmentStatus, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT fulfillment_status
		FROM order_items
		WHERE order_id = ?
		FOR UPDATE
	`, strings.TrimSpace(orderID))
	if err != nil {
		return nil, databaseUnavailable("lock all order item statuses", err)
	}
	defer rows.Close()

	var statuses []domain.FulfillmentStatus
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return nil, databaseUnavailable("scan order item status", err)
		}
		statuses = append(statuses, domain.FulfillmentStatus(status))
	}
	if err := rows.Err(); err != nil {
		return nil, databaseUnavailable("iterate order item statuses", err)
	}
	if len(statuses) == 0 {
		return nil, domain.ErrInvalidOrder
	}
	return statuses, nil
}

func persistSellerShipmentTransition(ctx context.Context, tx *sql.Tx, transition domain.SellerFulfillmentTransition) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT shipment_id
		FROM shipments
		WHERE order_id = ?
		  AND seller_id = ?
		FOR UPDATE
	`, transition.OrderID, transition.SellerID)
	if err != nil {
		return databaseUnavailable("lock seller shipments", err)
	}
	var existing []string
	for rows.Next() {
		var shipmentID string
		if err := rows.Scan(&shipmentID); err != nil {
			_ = rows.Close()
			return databaseUnavailable("scan seller shipment", err)
		}
		existing = append(existing, shipmentID)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return databaseUnavailable("iterate seller shipments", err)
	}
	if err := rows.Close(); err != nil {
		return databaseUnavailable("close seller shipment rows", err)
	}
	if len(existing) > 1 {
		return domain.ErrMultiSellerFulfillmentPending
	}

	shippedAt := any(nil)
	deliveredAt := any(nil)
	if transition.TargetStatus == domain.FulfillmentStatusShipped {
		shippedAt = transition.OccurredAt.UTC()
	}
	if transition.TargetStatus == domain.FulfillmentStatusDelivered {
		deliveredAt = transition.OccurredAt.UTC()
	}
	status := string(transition.TargetStatus)
	if len(existing) == 0 {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO shipments (
				shipment_id, order_id, seller_id, carrier, tracking_number, status, shipped_at, delivered_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
			transition.ShipmentID,
			transition.OrderID,
			transition.SellerID,
			nullableString(transition.Carrier),
			nullableString(transition.TrackingNumber),
			status,
			shippedAt,
			deliveredAt,
		)
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE shipments
			SET
				carrier = COALESCE(?, carrier),
				tracking_number = COALESCE(?, tracking_number),
				status = ?,
				shipped_at = COALESCE(?, shipped_at),
				delivered_at = COALESCE(?, delivered_at)
			WHERE shipment_id = ?
			  AND order_id = ?
			  AND seller_id = ?
		`,
			nullableString(transition.Carrier),
			nullableString(transition.TrackingNumber),
			status,
			shippedAt,
			deliveredAt,
			existing[0],
			transition.OrderID,
			transition.SellerID,
		)
	}
	if err != nil {
		return databaseUnavailable("persist seller shipment transition", err)
	}
	return nil
}

func persistAggregatedParentStatus(
	ctx context.Context,
	tx *sql.Tx,
	transition domain.SellerFulfillmentTransition,
	from domain.OrderStatus,
	to domain.OrderStatus,
) error {
	if err := domain.ValidateOrderStatusTransition(from, to); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE orders
		SET status = ?
		WHERE order_id = ?
		  AND status = ?
	`, to.String(), transition.OrderID, from.String())
	if err != nil {
		return databaseUnavailable("update aggregated parent fulfillment status", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return databaseUnavailable("read parent fulfillment status result", err)
	}
	if affected != 1 {
		return domain.ErrInvalidOrderStatusTransition
	}
	history := domain.OrderStatusHistoryEntry{
		ID:         transition.HistoryID,
		OrderID:    transition.OrderID,
		FromStatus: &from,
		ToStatus:   to,
		Reason:     "seller_fulfillment_" + to.String(),
		ActorType:  domain.OrderStatusActorSeller,
		ActorID:    transition.ActorID,
		Metadata: map[string]any{
			"seller_id":         transition.SellerID,
			"carrier":           transition.Carrier,
			"tracking_number":   transition.TrackingNumber,
			"seller_item_state": string(transition.TargetStatus),
			"note":              transition.Note,
		},
		CreatedAt: transition.OccurredAt.UTC(),
	}
	return insertStatusHistory(ctx, tx, history)
}
