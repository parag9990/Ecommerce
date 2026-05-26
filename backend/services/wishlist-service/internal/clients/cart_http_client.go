package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
	"ecommerce/backend/services/wishlist-service/internal/usecase"
)

const cartResponseBodyLimit = 1 << 20

type HTTPCartClient struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	logger     *slog.Logger
}

type HTTPCartClientOption func(*HTTPCartClient)

func WithCartHTTPClient(httpClient *http.Client) HTTPCartClientOption {
	return func(c *HTTPCartClient) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func NewHTTPCartClient(baseURL string, timeout time.Duration, logger *slog.Logger, options ...HTTPCartClientOption) (*HTTPCartClient, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("cart service base URL is required")
	}
	parsed, err := neturl.ParseRequestURI(baseURL)
	if err != nil {
		return nil, fmt.Errorf("cart service base URL is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("cart service base URL must use http or https")
	}
	if parsed.Host == "" {
		return nil, errors.New("cart service base URL host is required")
	}
	if timeout <= 0 {
		return nil, errors.New("cart service timeout must be positive")
	}
	if logger == nil {
		logger = slog.Default()
	}

	client := &HTTPCartClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
		logger:  logger,
	}
	for _, option := range options {
		option(client)
	}
	return client, nil
}

func (c *HTTPCartClient) AddItem(ctx context.Context, input usecase.CartAddItemInput) (*usecase.Cart, error) {
	if c == nil || c.httpClient == nil {
		return nil, errors.New("cart client is not initialized")
	}

	userID := strings.TrimSpace(input.UserID)
	productID := strings.TrimSpace(input.ProductID)
	variantID := strings.TrimSpace(input.VariantID)
	if userID == "" {
		return nil, usecase.ErrUnauthenticated
	}
	if productID == "" {
		return nil, usecase.ValidationError{Field: "product_id", Message: "is required"}
	}
	if variantID == "" {
		return nil, usecase.ValidationError{Field: "variant_id", Message: "is required"}
	}
	if input.Quantity < 1 {
		return nil, usecase.ValidationError{Field: "quantity", Message: "must be greater than or equal to 1"}
	}

	body, err := json.Marshal(cartAddItemRequest{
		ProductID: productID,
		VariantID: variantID,
		Quantity:  input.Quantity,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: encode cart request: %v", usecase.ErrCartServiceUnavailable, err)
	}

	ctx = contextOrBackground(ctx)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/cart/items", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: build cart request: %v", usecase.ErrCartServiceUnavailable, err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-User-ID", userID)
	request.Header.Set("X-User-Roles", "buyer")
	if requestID := strings.TrimSpace(input.RequestID); requestID != "" {
		request.Header.Set("X-Request-ID", requestID)
	}
	if idempotencyKey := strings.TrimSpace(input.IdempotencyKey); idempotencyKey != "" {
		request.Header.Set("X-Idempotency-Key", idempotencyKey)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", usecase.ErrCartServiceUnavailable, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, mapCartHTTPStatus(response.StatusCode)
	}

	cart, err := decodeCartResponse(response.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: decode cart response: %v", usecase.ErrCartServiceUnavailable, err)
	}
	return cart, nil
}

type cartAddItemRequest struct {
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

type cartEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *apiError       `json:"error"`
}

type cartResponse struct {
	CartID   string             `json:"cart_id"`
	ID       string             `json:"id"`
	MongoID  string             `json:"_id"`
	UserID   string             `json:"user_id"`
	Items    []cartItemResponse `json:"items"`
	Subtotal *cartMoneyResponse `json:"subtotal"`
	Discount *cartMoneyResponse `json:"discount"`
	Total    *cartMoneyResponse `json:"total"`
}

type cartItemResponse struct {
	ItemID    string `json:"item_id"`
	ID        string `json:"id"`
	MongoID   string `json:"_id"`
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

type cartMoneyResponse struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func decodeCartResponse(body io.Reader) (*usecase.Cart, error) {
	payload, err := io.ReadAll(io.LimitReader(body, cartResponseBodyLimit))
	if err != nil {
		return nil, err
	}
	var envelope cartEnvelope
	if err := json.Unmarshal(payload, &envelope); err == nil && len(envelope.Data) > 0 {
		payload = envelope.Data
	}

	var cart cartResponse
	if err := json.Unmarshal(payload, &cart); err != nil {
		return nil, err
	}
	return cart.toUsecase(), nil
}

func (c cartResponse) toUsecase() *usecase.Cart {
	items := make([]usecase.CartItem, 0, len(c.Items))
	for _, item := range c.Items {
		items = append(items, usecase.CartItem{
			ItemID:    firstNonEmpty(item.ItemID, item.ID, item.MongoID),
			ProductID: strings.TrimSpace(item.ProductID),
			VariantID: strings.TrimSpace(item.VariantID),
			Quantity:  item.Quantity,
		})
	}
	return &usecase.Cart{
		CartID:   firstNonEmpty(c.CartID, c.ID, c.MongoID),
		UserID:   strings.TrimSpace(c.UserID),
		Items:    items,
		Subtotal: c.Subtotal.toDomain(),
		Discount: c.Discount.toDomain(),
		Total:    c.Total.toDomain(),
	}
}

func (m *cartMoneyResponse) toDomain() *domain.Money {
	if m == nil {
		return nil
	}
	money := &domain.Money{
		Amount:   m.Amount,
		Currency: strings.ToUpper(strings.TrimSpace(m.Currency)),
	}
	if err := money.Validate(); err != nil {
		return nil
	}
	return money
}

func mapCartHTTPStatus(statusCode int) error {
	switch {
	case statusCode == http.StatusBadRequest:
		return fmt.Errorf("%w: status %d", usecase.ErrCartValidation, statusCode)
	case statusCode == http.StatusUnauthorized:
		return fmt.Errorf("%w: status %d", usecase.ErrCartUnauthenticated, statusCode)
	case statusCode == http.StatusForbidden:
		return fmt.Errorf("%w: status %d", usecase.ErrCartForbidden, statusCode)
	case statusCode == http.StatusNotFound:
		return fmt.Errorf("%w: status %d", usecase.ErrCartProductNotFound, statusCode)
	case statusCode == http.StatusConflict || statusCode == http.StatusGone:
		return fmt.Errorf("%w: status %d", usecase.ErrCartItemUnavailable, statusCode)
	case statusCode == http.StatusRequestTimeout || statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError:
		return fmt.Errorf("%w: status %d", usecase.ErrCartServiceUnavailable, statusCode)
	default:
		return fmt.Errorf("%w: unexpected status %d", usecase.ErrCartServiceUnavailable, statusCode)
	}
}
