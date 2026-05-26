package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

type MySQLOrderRepository struct {
	db *sql.DB
}

func NewMySQLOrderRepository(db *sql.DB) (*MySQLOrderRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return &MySQLOrderRepository{db: db}, nil
}

func (r *MySQLOrderRepository) CreateOrderAndAttachClaim(ctx context.Context, order domain.Order, claim domain.IdempotencyClaim, event domain.OutboxEvent) error {
	if err := order.ValidateForCreation(); err != nil {
		return err
	}
	if err := claim.ValidateForOrderBinding(order.UserID, claim.Key); err != nil {
		return err
	}
	if err := validateOutboxEventForHistory(event, order.OrderID, order.InitialHistory.ID); err != nil {
		return err
	}
	addressJSON, err := json.Marshal(order.AddressSnapshot)
	if err != nil {
		return fmt.Errorf("marshal address snapshot: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin create order transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := insertOrder(ctx, tx, order, addressJSON); err != nil {
		return err
	}
	for _, item := range order.Items {
		if err := insertOrderItem(ctx, tx, item); err != nil {
			return err
		}
	}
	if err := insertStatusHistory(ctx, tx, order.InitialHistory); err != nil {
		return err
	}
	if err := insertOutboxEvent(ctx, tx, event); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE order_idempotency_keys
		SET order_id = ?
		WHERE user_id = ?
		  AND idempotency_key = ?
		  AND request_hash = ?
		  AND status = 'processing'
		  AND order_id IS NULL
	`,
		order.OrderID,
		claim.UserID,
		claim.Key,
		claim.RequestHash,
	)
	if err != nil {
		return databaseUnavailable("attach order to idempotency claim", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return databaseUnavailable("read idempotency claim binding result", err)
	}
	if affected != 1 {
		return domain.ErrCheckoutInProgress
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: commit create order transaction: %w", domain.ErrOrderCreateOutcomeUnknown, err)
	}
	return nil
}

func insertOrder(ctx context.Context, tx *sql.Tx, order domain.Order, addressJSON []byte) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO orders (
			order_id,
			user_id,
			cart_id,
			status,
			currency,
			subtotal_amount,
			discount_amount,
			shipping_amount,
			tax_amount,
			total_amount,
			address_snapshot,
			coupon_code,
			payment_id,
			inventory_reservation_id,
			inventory_reserved_until
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, ?, ?)
	`,
		order.OrderID,
		order.UserID,
		order.CartID,
		order.Status.String(),
		order.Currency,
		order.SubtotalAmount,
		order.DiscountAmount,
		order.ShippingAmount,
		order.TaxAmount,
		order.TotalAmount,
		addressJSON,
		nullableString(order.CouponCode),
		order.InventoryReservationID,
		order.InventoryReservedUntil.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	return nil
}

func insertOrderItem(ctx context.Context, tx *sql.Tx, item domain.OrderItem) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO order_items (
			order_item_id,
			order_id,
			seller_id,
			product_id,
			variant_id,
			sku,
			title_snapshot,
			image_url_snapshot,
			quantity,
			currency,
			unit_amount,
			discount_amount,
			tax_amount,
			total_amount,
			fulfillment_status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		item.OrderItemID,
		item.OrderID,
		item.SellerID,
		item.ProductID,
		item.VariantID,
		item.SKU,
		item.TitleSnapshot,
		nullableString(item.ImageURLSnapshot),
		item.Quantity,
		item.Currency,
		item.UnitAmount,
		item.DiscountAmount,
		item.TaxAmount,
		item.TotalAmount,
		string(item.FulfillmentStatus),
	)
	if err != nil {
		return fmt.Errorf("insert order item: %w", err)
	}
	return nil
}

func (r *MySQLOrderRepository) GetForPayment(ctx context.Context, orderID string) (domain.Order, error) {
	var (
		order         domain.Order
		status        string
		paymentID     sql.NullString
		reservationID sql.NullString
		reservedUntil sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT
			order_id,
			user_id,
			status,
			currency,
			total_amount,
			payment_id,
			inventory_reservation_id,
			inventory_reserved_until
		FROM orders
		WHERE order_id = ?
	`, strings.TrimSpace(orderID)).Scan(
		&order.OrderID,
		&order.UserID,
		&status,
		&order.Currency,
		&order.TotalAmount,
		&paymentID,
		&reservationID,
		&reservedUntil,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if err != nil {
		return domain.Order{}, databaseUnavailable("get order for payment", err)
	}
	order.Status, err = domain.ParseOrderStatus(status)
	if err != nil {
		return domain.Order{}, fmt.Errorf("parse order payment status: %w", err)
	}
	if paymentID.Valid {
		order.PaymentID = paymentID.String
	}
	if reservationID.Valid {
		order.InventoryReservationID = reservationID.String
	}
	if reservedUntil.Valid {
		order.InventoryReservedUntil = reservedUntil.Time.UTC()
	}
	return order, nil
}

const orderHeaderSelect = `
	SELECT
		order_id,
		user_id,
		cart_id,
		status,
		currency,
		subtotal_amount,
		discount_amount,
		shipping_amount,
		tax_amount,
		total_amount,
		address_snapshot,
		coupon_code,
		payment_id,
		inventory_reservation_id,
		inventory_reserved_until,
		created_at,
		updated_at
	FROM orders
`

func (r *MySQLOrderRepository) GetOrder(ctx context.Context, orderID string) (domain.Order, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return domain.Order{}, domain.ErrInvalidRequest
	}
	order, err := scanOrderHeader(r.db.QueryRowContext(ctx, orderHeaderSelect+" WHERE order_id = ?", orderID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}
	orders := []domain.Order{order}
	if err := r.loadOrderItems(ctx, orders); err != nil {
		return domain.Order{}, err
	}
	return orders[0], nil
}

func (r *MySQLOrderRepository) ListOrders(ctx context.Context, filter domain.ListOrdersFilter) (domain.OrderPage, error) {
	if strings.TrimSpace(filter.UserID) == "" || filter.PageSize <= 0 {
		return domain.OrderPage{}, domain.ErrInvalidRequest
	}
	query := orderHeaderSelect + " WHERE user_id = ?"
	args := []any{strings.TrimSpace(filter.UserID)}
	if filter.StatusFilter != nil {
		query += " AND status = ?"
		args = append(args, filter.StatusFilter.String())
	}
	if filter.Cursor != nil {
		query += " AND (created_at < ? OR (created_at = ? AND order_id < ?))"
		args = append(args, filter.Cursor.CreatedAt.UTC(), filter.Cursor.CreatedAt.UTC(), filter.Cursor.OrderID)
	}
	query += " ORDER BY created_at DESC, order_id DESC LIMIT ?"
	args = append(args, filter.PageSize+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.OrderPage{}, databaseUnavailable("list orders", err)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0, filter.PageSize+1)
	for rows.Next() {
		order, scanErr := scanOrderHeader(rows)
		if scanErr != nil {
			return domain.OrderPage{}, fmt.Errorf("scan listed order: %w", scanErr)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return domain.OrderPage{}, databaseUnavailable("iterate listed orders", err)
	}

	var cursor *domain.OrderCursor
	if len(orders) > filter.PageSize {
		last := orders[filter.PageSize-1]
		cursor = &domain.OrderCursor{CreatedAt: last.CreatedAt, OrderID: last.OrderID}
		orders = orders[:filter.PageSize]
	}
	if err := r.loadOrderItems(ctx, orders); err != nil {
		return domain.OrderPage{}, err
	}
	return domain.OrderPage{Orders: orders, NextCursor: cursor}, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanOrderHeader(row scanner) (domain.Order, error) {
	var (
		order         domain.Order
		cartID        sql.NullString
		status        string
		addressJSON   []byte
		couponCode    sql.NullString
		paymentID     sql.NullString
		reservationID sql.NullString
		reservedUntil sql.NullTime
	)
	err := row.Scan(
		&order.OrderID,
		&order.UserID,
		&cartID,
		&status,
		&order.Currency,
		&order.SubtotalAmount,
		&order.DiscountAmount,
		&order.ShippingAmount,
		&order.TaxAmount,
		&order.TotalAmount,
		&addressJSON,
		&couponCode,
		&paymentID,
		&reservationID,
		&reservedUntil,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return domain.Order{}, err
	}
	order.Status, err = domain.ParseOrderStatus(status)
	if err != nil {
		return domain.Order{}, err
	}
	if err := json.Unmarshal(addressJSON, &order.AddressSnapshot); err != nil {
		return domain.Order{}, fmt.Errorf("unmarshal address snapshot: %w", err)
	}
	if cartID.Valid {
		order.CartID = cartID.String
	}
	if couponCode.Valid {
		order.CouponCode = couponCode.String
	}
	if paymentID.Valid {
		order.PaymentID = paymentID.String
	}
	if reservationID.Valid {
		order.InventoryReservationID = reservationID.String
	}
	if reservedUntil.Valid {
		order.InventoryReservedUntil = reservedUntil.Time.UTC()
	}
	order.CreatedAt = order.CreatedAt.UTC()
	order.UpdatedAt = order.UpdatedAt.UTC()
	return order, nil
}

func (r *MySQLOrderRepository) loadOrderItems(ctx context.Context, orders []domain.Order) error {
	if len(orders) == 0 {
		return nil
	}
	index := make(map[string]int, len(orders))
	args := make([]any, 0, len(orders))
	for position := range orders {
		index[orders[position].OrderID] = position
		args = append(args, orders[position].OrderID)
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(orders)), ",")
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			order_item_id, order_id, seller_id, product_id, variant_id, sku,
			title_snapshot, image_url_snapshot, quantity, currency, unit_amount,
			discount_amount, tax_amount, total_amount, fulfillment_status
		FROM order_items
		WHERE order_id IN (`+placeholders+`)
		ORDER BY id ASC
	`, args...)
	if err != nil {
		return databaseUnavailable("list order items", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			item     domain.OrderItem
			imageURL sql.NullString
			status   string
		)
		if err := rows.Scan(
			&item.OrderItemID, &item.OrderID, &item.SellerID, &item.ProductID, &item.VariantID,
			&item.SKU, &item.TitleSnapshot, &imageURL, &item.Quantity, &item.Currency,
			&item.UnitAmount, &item.DiscountAmount, &item.TaxAmount, &item.TotalAmount, &status,
		); err != nil {
			return fmt.Errorf("scan order item: %w", err)
		}
		if imageURL.Valid {
			item.ImageURLSnapshot = imageURL.String
		}
		item.FulfillmentStatus = domain.FulfillmentStatus(status)
		position, exists := index[item.OrderID]
		if !exists {
			return errors.New("order item returned for an unrequested order")
		}
		orders[position].Items = append(orders[position].Items, item)
	}
	if err := rows.Err(); err != nil {
		return databaseUnavailable("iterate order items", err)
	}
	return nil
}

func (r *MySQLOrderRepository) AttachPaymentIntent(
	ctx context.Context,
	orderID string,
	paymentID string,
	history domain.OrderStatusHistoryEntry,
) (bool, error) {
	if strings.TrimSpace(paymentID) == "" {
		return false, domain.ErrInvalidPaymentIntentResponse
	}
	if err := validateTransitionHistory(history, orderID, domain.OrderStatusCreated, domain.OrderStatusPendingPayment); err != nil {
		return false, err
	}
	return r.transitionWithHistory(ctx, history, nil, `
		UPDATE orders
		SET payment_id = ?, status = ?
		WHERE order_id = ? AND status = ? AND payment_id IS NULL
	`, strings.TrimSpace(paymentID), domain.OrderStatusPendingPayment.String(), strings.TrimSpace(orderID), domain.OrderStatusCreated.String())
}

func (r *MySQLOrderRepository) MarkPaymentFailedFromCreated(
	ctx context.Context,
	orderID string,
	history domain.OrderStatusHistoryEntry,
) (bool, error) {
	if err := validateTransitionHistory(history, orderID, domain.OrderStatusCreated, domain.OrderStatusPaymentFailed); err != nil {
		return false, err
	}
	return r.transitionWithHistory(ctx, history, nil, `
		UPDATE orders
		SET status = ?
		WHERE order_id = ? AND status = ? AND payment_id IS NULL
	`, domain.OrderStatusPaymentFailed.String(), strings.TrimSpace(orderID), domain.OrderStatusCreated.String())
}

func (r *MySQLOrderRepository) TransitionPaymentResult(
	ctx context.Context,
	orderID string,
	paymentID string,
	from domain.OrderStatus,
	to domain.OrderStatus,
	history domain.OrderStatusHistoryEntry,
	event *domain.OutboxEvent,
) (bool, error) {
	if strings.TrimSpace(paymentID) == "" {
		return false, domain.ErrPaymentOrderMismatch
	}
	if err := validateTransitionHistory(history, orderID, from, to); err != nil {
		return false, err
	}
	return r.transitionWithHistory(ctx, history, event, `
		UPDATE orders
		SET status = ?
		WHERE order_id = ? AND payment_id = ? AND status = ?
	`, to.String(), strings.TrimSpace(orderID), strings.TrimSpace(paymentID), from.String())
}

func (r *MySQLOrderRepository) UpdateFulfillment(ctx context.Context, transition domain.FulfillmentTransition, event *domain.OutboxEvent) (domain.Order, error) {
	if event != nil {
		if err := validateOutboxEventForHistory(*event, transition.OrderID, transition.History.ID); err != nil {
			return domain.Order{}, err
		}
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return domain.Order{}, databaseUnavailable("begin fulfillment transaction", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var currentValue string
	err = tx.QueryRowContext(ctx, `
		SELECT status
		FROM orders
		WHERE order_id = ?
		FOR UPDATE
	`, strings.TrimSpace(transition.OrderID)).Scan(&currentValue)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if err != nil {
		return domain.Order{}, databaseUnavailable("load order for fulfillment", err)
	}
	current, err := domain.ParseOrderStatus(currentValue)
	if err != nil {
		return domain.Order{}, err
	}
	if transition.History.FromStatus == nil || *transition.History.FromStatus != current {
		return domain.Order{}, domain.ErrInvalidOrderStatusTransition
	}
	if err := transition.Validate(current); err != nil {
		return domain.Order{}, err
	}

	sellerID, err := singleShipmentSeller(ctx, tx, transition.OrderID)
	if err != nil {
		return domain.Order{}, err
	}
	if transition.RequiredSellerID != "" && sellerID != transition.RequiredSellerID {
		return domain.Order{}, domain.ErrForbidden
	}
	if err := persistShipmentTransition(ctx, tx, transition, sellerID); err != nil {
		return domain.Order{}, err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE orders
		SET status = ?
		WHERE order_id = ? AND status = ?
	`, transition.TargetStatus.String(), transition.OrderID, current.String())
	if err != nil {
		return domain.Order{}, databaseUnavailable("update fulfillment order status", err)
	}
	updated, err := result.RowsAffected()
	if err != nil || updated != 1 {
		return domain.Order{}, domain.ErrInvalidOrderStatusTransition
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE order_items
		SET fulfillment_status = ?
		WHERE order_id = ?
	`, transition.TargetStatus.String(), transition.OrderID); err != nil {
		return domain.Order{}, databaseUnavailable("update order item fulfillment statuses", err)
	}
	if err := insertStatusHistory(ctx, tx, transition.History); err != nil {
		return domain.Order{}, databaseUnavailable("insert fulfillment status history", err)
	}
	if event != nil {
		if err := insertOutboxEvent(ctx, tx, *event); err != nil {
			return domain.Order{}, databaseUnavailable("insert fulfillment outbox event", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.Order{}, databaseUnavailable("commit fulfillment transaction", err)
	}
	return r.GetOrder(ctx, transition.OrderID)
}

func singleShipmentSeller(ctx context.Context, tx *sql.Tx, orderID string) (string, error) {
	var (
		sellerID       sql.NullString
		itemCount      int
		distinctSeller int
	)
	err := tx.QueryRowContext(ctx, `
		SELECT MIN(seller_id), COUNT(*), COUNT(DISTINCT seller_id)
		FROM order_items
		WHERE order_id = ?
	`, strings.TrimSpace(orderID)).Scan(&sellerID, &itemCount, &distinctSeller)
	if err != nil {
		return "", databaseUnavailable("load fulfillment sellers", err)
	}
	if itemCount == 0 {
		return "", domain.ErrInvalidOrder
	}
	if distinctSeller != 1 || !sellerID.Valid {
		return "", domain.ErrMultiSellerFulfillmentPending
	}
	return sellerID.String, nil
}

func persistShipmentTransition(ctx context.Context, tx *sql.Tx, transition domain.FulfillmentTransition, sellerID string) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT shipment_id
		FROM shipments
		WHERE order_id = ?
		FOR UPDATE
	`, transition.OrderID)
	if err != nil {
		return databaseUnavailable("lock order shipments", err)
	}
	var existing []string
	for rows.Next() {
		var shipmentID string
		if err := rows.Scan(&shipmentID); err != nil {
			_ = rows.Close()
			return databaseUnavailable("scan shipment", err)
		}
		existing = append(existing, shipmentID)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return databaseUnavailable("iterate shipments", err)
	}
	if err := rows.Close(); err != nil {
		return databaseUnavailable("close shipment rows", err)
	}
	if len(existing) > 1 {
		return domain.ErrMultiSellerFulfillmentPending
	}
	shippedAt := any(nil)
	deliveredAt := any(nil)
	if transition.TargetStatus == domain.OrderStatusShipped {
		shippedAt = transition.OccurredAt.UTC()
	}
	if transition.TargetStatus == domain.OrderStatusDelivered {
		deliveredAt = transition.OccurredAt.UTC()
	}
	if len(existing) == 0 {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO shipments (
				shipment_id, order_id, seller_id, carrier, tracking_number, status, shipped_at, delivered_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
			transition.ShipmentID,
			transition.OrderID,
			sellerID,
			nullableString(transition.Carrier),
			nullableString(transition.TrackingNumber),
			transition.TargetStatus.String(),
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
		`,
			nullableString(transition.Carrier),
			nullableString(transition.TrackingNumber),
			transition.TargetStatus.String(),
			shippedAt,
			deliveredAt,
			existing[0],
		)
	}
	if err != nil {
		return databaseUnavailable("persist shipment transition", err)
	}
	return nil
}

func (r *MySQLOrderRepository) transitionWithHistory(
	ctx context.Context,
	history domain.OrderStatusHistoryEntry,
	event *domain.OutboxEvent,
	query string,
	args ...any,
) (bool, error) {
	if event != nil {
		if err := validateOutboxEventForHistory(*event, history.OrderID, history.ID); err != nil {
			return false, err
		}
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return false, fmt.Errorf("begin payment status transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("transition payment status: %w", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read payment status transition result: %w", err)
	}
	if changed == 0 {
		return false, nil
	}
	if changed != 1 {
		return false, fmt.Errorf("transition payment status changed %d orders", changed)
	}
	if err := insertStatusHistory(ctx, tx, history); err != nil {
		return false, err
	}
	if event != nil {
		if err := insertOutboxEvent(ctx, tx, *event); err != nil {
			return false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit payment status transaction: %w", err)
	}
	return true, nil
}

func validateOutboxEventForHistory(event domain.OutboxEvent, orderID string, historyID string) error {
	orderID = strings.TrimSpace(orderID)
	historyID = strings.TrimSpace(historyID)
	if event.AggregateID != orderID || event.SourceHistoryID != historyID {
		return domain.ErrInvalidOrder
	}
	return event.ValidateForInsert()
}

func validateTransitionHistory(history domain.OrderStatusHistoryEntry, orderID string, from domain.OrderStatus, to domain.OrderStatus) error {
	if err := domain.ValidateOrderStatusTransition(from, to); err != nil {
		return err
	}
	if strings.TrimSpace(history.ID) == "" || history.OrderID != strings.TrimSpace(orderID) ||
		history.FromStatus == nil || *history.FromStatus != from || history.ToStatus != to ||
		strings.TrimSpace(history.Reason) == "" || !domain.IsKnownOrderStatusActorType(history.ActorType) ||
		history.CreatedAt.IsZero() {
		return domain.ErrInvalidOrder
	}
	return nil
}

func insertStatusHistory(ctx context.Context, tx *sql.Tx, history domain.OrderStatusHistoryEntry) error {
	metadataJSON, err := nullableJSON(history.Metadata)
	if err != nil {
		return fmt.Errorf("marshal status history metadata: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO order_status_history (
			history_id,
			order_id,
			from_status,
			to_status,
			reason,
			actor_type,
			changed_by,
			metadata,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		history.ID,
		history.OrderID,
		nullableStatus(history.FromStatus),
		history.ToStatus.String(),
		nullableString(history.Reason),
		history.ActorType.String(),
		nullableString(history.ActorID),
		metadataJSON,
		history.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert order status history: %w", err)
	}
	return nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableStatus(status *domain.OrderStatus) any {
	if status == nil {
		return nil
	}
	return status.String()
}

func nullableJSON(value map[string]any) (any, error) {
	if value == nil {
		return nil, nil
	}
	serialized, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return serialized, nil
}

func databaseUnavailable(operation string, err error) error {
	return fmt.Errorf("%w: %s: %w", domain.ErrTemporarilyUnavailable, operation, err)
}
