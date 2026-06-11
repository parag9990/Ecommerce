package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

type ListSellerOrdersUsecase struct {
	orders   SellerOrderRepository
	tokenKey []byte
	logger   *slog.Logger
}

func NewListSellerOrdersUsecase(
	orders SellerOrderRepository,
	tokenKey []byte,
	logger *slog.Logger,
) (*ListSellerOrdersUsecase, error) {
	if orders == nil {
		return nil, errors.New("seller order repository is required")
	}
	if len(tokenKey) < 32 {
		return nil, errors.New("pagination token signing key must be at least 32 bytes")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &ListSellerOrdersUsecase{
		orders:   orders,
		tokenKey: append([]byte(nil), tokenKey...),
		logger:   logger,
	}, nil
}

func (u *ListSellerOrdersUsecase) Execute(ctx context.Context, query ListSellerOrdersQuery) (SellerOrderPage, error) {
	query.SellerID = strings.TrimSpace(query.SellerID)
	query.ActorUserID = strings.TrimSpace(query.ActorUserID)
	if query.SellerID == "" || query.ActorUserID == "" ||
		len(query.SellerID) > maxRequestIDLength || len(query.ActorUserID) > maxRequestIDLength {
		return SellerOrderPage{}, domain.ErrInvalidRequest
	}
	if query.PageSize <= 0 || query.PageSize > MaxOrderPageSize {
		return SellerOrderPage{}, domain.ErrInvalidRequest
	}
	if query.FulfillmentFilter != nil && !domain.IsKnownSellerFulfillmentStatus(*query.FulfillmentFilter) {
		return SellerOrderPage{}, domain.ErrInvalidRequest
	}
	cursor, err := decodePageToken(u.tokenKey, strings.TrimSpace(query.PageToken))
	if err != nil {
		return SellerOrderPage{}, err
	}
	page, err := u.orders.ListSellerOrders(ctx, domain.ListSellerOrdersFilter{
		SellerID:          query.SellerID,
		PageSize:          query.PageSize,
		Cursor:            cursor,
		FulfillmentFilter: query.FulfillmentFilter,
	})
	if err != nil {
		return SellerOrderPage{}, err
	}
	if page.NextCursor != nil {
		token, err := encodePageToken(u.tokenKey, *page.NextCursor)
		if err != nil {
			return SellerOrderPage{}, err
		}
		page.NextCursor = nil
		page.NextPageToken = token
	}
	u.logger.Info("seller.orders.listed",
		slog.String("seller_id", query.SellerID),
		slog.String("actor_user_id", query.ActorUserID),
		slog.Int("result_count", len(page.Orders)),
	)
	return page, nil
}
