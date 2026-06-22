package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

type AdminOrderPage struct {
	Orders []domain.Order
	Total  int64
}

func scanStatusHistoryEntry(row scanner) (domain.OrderStatusHistoryEntry, error) {
	var entry domain.OrderStatusHistoryEntry
	var fromStatus, reason, changedBy, metadata sql.NullString
	var toStatus, actorType string
	if err := row.Scan(&entry.ID, &entry.OrderID, &fromStatus, &toStatus, &reason, &actorType, &changedBy, &metadata, &entry.CreatedAt); err != nil {
		return domain.OrderStatusHistoryEntry{}, err
	}
	parsedTo, err := domain.ParseOrderStatus(toStatus)
	if err != nil {
		return domain.OrderStatusHistoryEntry{}, err
	}
	entry.ToStatus = parsedTo
	if fromStatus.Valid {
		parsed, err := domain.ParseOrderStatus(fromStatus.String)
		if err != nil {
			return domain.OrderStatusHistoryEntry{}, err
		}
		entry.FromStatus = &parsed
	}
	parsedActor, err := domain.ParseOrderStatusActorType(actorType)
	if err != nil {
		return domain.OrderStatusHistoryEntry{}, err
	}
	entry.ActorType, entry.Reason, entry.ActorID, entry.CreatedAt = parsedActor, reason.String, changedBy.String, entry.CreatedAt.UTC()
	if metadata.Valid && strings.TrimSpace(metadata.String) != "" {
		if err := json.Unmarshal([]byte(metadata.String), &entry.Metadata); err != nil {
			return domain.OrderStatusHistoryEntry{}, err
		}
	}
	return entry, nil
}

func (r *MySQLOrderRepository) ListOrdersForAdmin(ctx context.Context, status, userID, sellerID string, limit, offset int) (AdminOrderPage, error) {
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 5)
	if status = strings.TrimSpace(status); status != "" {
		conditions = append(conditions, "o.status = ?")
		args = append(args, status)
	}
	if userID = strings.TrimSpace(userID); userID != "" {
		conditions = append(conditions, "o.user_id = ?")
		args = append(args, userID)
	}
	if sellerID = strings.TrimSpace(sellerID); sellerID != "" {
		conditions = append(conditions, "EXISTS (SELECT 1 FROM order_items oi WHERE oi.order_id = o.order_id AND oi.seller_id = ?)")
		args = append(args, sellerID)
	}
	where := " WHERE " + strings.Join(conditions, " AND ")
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders o"+where, args...).Scan(&total); err != nil {
		return AdminOrderPage{}, fmt.Errorf("count admin orders: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, orderHeaderSelect+" o"+where+" ORDER BY o.created_at DESC, o.order_id DESC LIMIT ? OFFSET ?", append(args, limit, offset)...)
	if err != nil {
		return AdminOrderPage{}, databaseUnavailable("list admin orders", err)
	}
	defer rows.Close()
	orders := make([]domain.Order, 0, limit)
	for rows.Next() {
		order, err := scanOrderHeader(rows)
		if err != nil {
			return AdminOrderPage{}, fmt.Errorf("scan admin order: %w", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return AdminOrderPage{}, databaseUnavailable("iterate admin orders", err)
	}
	if err := r.loadOrderItems(ctx, orders); err != nil {
		return AdminOrderPage{}, err
	}
	return AdminOrderPage{Orders: orders, Total: total}, nil
}

func (r *MySQLOrderRepository) GetOrderStatusHistoryForAdmin(ctx context.Context, orderID string) ([]domain.OrderStatusHistoryEntry, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT history_id, order_id, from_status, to_status, reason, actor_type, changed_by, metadata, created_at FROM order_status_history WHERE order_id = ? ORDER BY created_at ASC, id ASC`, strings.TrimSpace(orderID))
	if err != nil {
		return nil, databaseUnavailable("list order status history", err)
	}
	defer rows.Close()
	history := make([]domain.OrderStatusHistoryEntry, 0)
	for rows.Next() {
		entry, err := scanStatusHistoryEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("scan order status history: %w", err)
		}
		history = append(history, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, databaseUnavailable("iterate order status history", err)
	}
	return history, nil
}
