package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

type GetOrderUsecase struct {
	orders OrderReadRepository
}

func NewGetOrderUsecase(orders OrderReadRepository) (*GetOrderUsecase, error) {
	if orders == nil {
		return nil, errors.New("order read repository is required")
	}
	return &GetOrderUsecase{orders: orders}, nil
}

func (u *GetOrderUsecase) Execute(ctx context.Context, query GetOrderQuery) (domain.Order, error) {
	query.ActorID = strings.TrimSpace(query.ActorID)
	query.OrderID = strings.TrimSpace(query.OrderID)
	if query.ActorID == "" || query.OrderID == "" ||
		len(query.ActorID) > maxRequestIDLength || len(query.OrderID) > maxRequestIDLength {
		return domain.Order{}, domain.ErrInvalidRequest
	}
	order, err := u.orders.GetOrder(ctx, query.OrderID)
	if err != nil {
		return domain.Order{}, err
	}
	if order.UserID != query.ActorID && !hasAnyRole(query.Roles, "admin", "order_manager") {
		return domain.Order{}, domain.ErrOrderForbidden
	}
	return order, nil
}
