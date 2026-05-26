package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

const (
	DefaultOrderPageSize = 20
	MaxOrderPageSize     = 100
	maxPageTokenLength   = 1024
)

type ListOrdersUsecase struct {
	orders   OrderReadRepository
	tokenKey []byte
}

type pageTokenPayload struct {
	CreatedAt string `json:"created_at"`
	OrderID   string `json:"order_id"`
}

func NewListOrdersUsecase(orders OrderReadRepository, tokenKey []byte) (*ListOrdersUsecase, error) {
	if orders == nil {
		return nil, errors.New("order read repository is required")
	}
	if len(tokenKey) < 32 {
		return nil, errors.New("pagination token signing key must be at least 32 bytes")
	}
	return &ListOrdersUsecase{orders: orders, tokenKey: append([]byte(nil), tokenKey...)}, nil
}

func (u *ListOrdersUsecase) Execute(ctx context.Context, query ListOrdersQuery) (OrderPage, error) {
	query.UserID = strings.TrimSpace(query.UserID)
	if query.UserID == "" || len(query.UserID) > maxRequestIDLength {
		return OrderPage{}, domain.ErrInvalidRequest
	}
	if query.PageSize <= 0 || query.PageSize > MaxOrderPageSize {
		return OrderPage{}, domain.ErrInvalidRequest
	}
	if query.StatusFilter != nil && !domain.IsKnownOrderStatus(*query.StatusFilter) {
		return OrderPage{}, domain.ErrInvalidRequest
	}
	cursor, err := u.decodeToken(strings.TrimSpace(query.PageToken))
	if err != nil {
		return OrderPage{}, err
	}
	page, err := u.orders.ListOrders(ctx, ListOrdersFilter{
		UserID:       query.UserID,
		PageSize:     query.PageSize,
		Cursor:       cursor,
		StatusFilter: query.StatusFilter,
	})
	if err != nil {
		return OrderPage{}, err
	}
	if page.NextCursor == nil {
		return page, nil
	}
	token, err := u.encodeToken(*page.NextCursor)
	if err != nil {
		return OrderPage{}, err
	}
	page.NextCursor = nil
	page.NextPageToken = token
	return page, nil
}

func (u *ListOrdersUsecase) encodeToken(cursor OrderCursor) (string, error) {
	payload, err := json.Marshal(pageTokenPayload{
		CreatedAt: cursor.CreatedAt.UTC().Format(time.RFC3339Nano),
		OrderID:   cursor.OrderID,
	})
	if err != nil {
		return "", err
	}
	signature := hmac.New(sha256.New, u.tokenKey)
	_, _ = signature.Write(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString(signature.Sum(nil)), nil
}

func (u *ListOrdersUsecase) decodeToken(token string) (*OrderCursor, error) {
	if token == "" {
		return nil, nil
	}
	if len(token) > maxPageTokenLength {
		return nil, domain.ErrInvalidPageToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, domain.ErrInvalidPageToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, domain.ErrInvalidPageToken
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, domain.ErrInvalidPageToken
	}
	expected := hmac.New(sha256.New, u.tokenKey)
	_, _ = expected.Write(payload)
	if !hmac.Equal(signature, expected.Sum(nil)) {
		return nil, domain.ErrInvalidPageToken
	}
	var decoded pageTokenPayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return nil, domain.ErrInvalidPageToken
	}
	createdAt, err := time.Parse(time.RFC3339Nano, decoded.CreatedAt)
	if err != nil || strings.TrimSpace(decoded.OrderID) == "" || len(decoded.OrderID) > maxRequestIDLength {
		return nil, domain.ErrInvalidPageToken
	}
	return &OrderCursor{CreatedAt: createdAt.UTC(), OrderID: strings.TrimSpace(decoded.OrderID)}, nil
}

func hasAnyRole(roles []string, allowed ...string) bool {
	for _, role := range roles {
		role = strings.ToLower(strings.TrimSpace(role))
		for _, candidate := range allowed {
			if role == candidate {
				return true
			}
		}
	}
	return false
}
