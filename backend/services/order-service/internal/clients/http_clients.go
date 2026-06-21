package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/usecase"
)

const maxResponseBytes = 2 << 20

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("downstream returned HTTP %d: %s", e.StatusCode, e.Body)
}

type CartHTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewCartHTTPClient(baseURL string, client *http.Client) (*CartHTTPClient, error) {
	normalized, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("cart base URL: %w", err)
	}
	return &CartHTTPClient{baseURL: normalized, client: requiredHTTPClient(client)}, nil
}

func (c *CartHTTPClient) GetCart(ctx context.Context, req usecase.GetCartRequest) (*domain.CartSnapshot, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/internal/v1/carts/"+url.PathEscape(strings.TrimSpace(req.CartID)), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-User-ID", strings.TrimSpace(req.UserID))
	var response cartResponse
	if err := doJSON(c.client, request, &response); err != nil {
		var httpErr *HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return nil, domain.ErrCartNotFound
		}
		return nil, fmt.Errorf("cart service request: %w", err)
	}
	items := make([]domain.CartItemSnapshot, 0, len(response.Items))
	for _, item := range response.Items {
		items = append(items, domain.CartItemSnapshot{ProductID: item.ProductID, VariantID: item.VariantID, Quantity: int32(item.Quantity)})
	}
	return &domain.CartSnapshot{
		CartID: response.CartID,
		UserID: stringValue(response.UserID),
		Currency: response.Totals.Currency,
		Items: items,
	}, nil
}

type ProductHTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewProductHTTPClient(baseURL string, client *http.Client) (*ProductHTTPClient, error) {
	normalized, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("product base URL: %w", err)
	}
	return &ProductHTTPClient{baseURL: normalized, client: requiredHTTPClient(client)}, nil
}

func (c *ProductHTTPClient) BatchGetProducts(ctx context.Context, req usecase.BatchGetProductsRequest) (*usecase.BatchGetProductsResponse, error) {
	productIDs := make([]string, 0, len(req.Items))
	seen := make(map[string]struct{}, len(req.Items))
	for _, item := range req.Items {
		id := strings.TrimSpace(item.ProductID)
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			productIDs = append(productIDs, id)
		}
	}
	var response productListResponse
	if err := c.post(ctx, "/internal/v1/products/batch", map[string]any{"product_ids": productIDs}, &response, ""); err != nil {
		return nil, fmt.Errorf("batch product lookup: %w", err)
	}

	requested := make(map[string]struct{}, len(req.Items))
	for _, item := range req.Items {
		requested[productKey(item.ProductID, item.VariantID)] = struct{}{}
	}
	result := make([]domain.ProductSnapshot, 0, len(req.Items))
	for _, product := range response.Products {
		for _, variant := range product.Variants {
			if _, ok := requested[productKey(product.ProductID, variant.VariantID)]; !ok {
				continue
			}
			result = append(result, domain.ProductSnapshot{
				ProductID: product.ProductID, VariantID: variant.VariantID, SellerID: product.SellerID,
				SKU: variant.SKU, Title: product.Title, ImageURL: primaryImageURL(product.Images),
				Currency: variant.Price.Currency, UnitAmount: variant.Price.Amount,
				Published: strings.EqualFold(product.Status, "published"),
				VariantActive: strings.EqualFold(variant.Status, "active"),
				InStock: variant.AvailableQuantity > 0,
			})
		}
	}
	return &usecase.BatchGetProductsResponse{Items: result}, nil
}

func (c *ProductHTTPClient) ReserveInventory(ctx context.Context, req usecase.ReserveInventoryRequest) (*usecase.ReserveInventoryResponse, error) {
	items := make([]map[string]any, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, map[string]any{"product_id": item.ProductID, "variant_id": item.VariantID, "quantity": item.Quantity})
	}
	input := map[string]any{
		"order_id": req.OrderID, "items": items, "ttl_seconds": int(req.TTL.Seconds()), "idempotency_key": req.IdempotencyKey,
	}
	var response inventoryReservationResponse
	if err := c.post(ctx, "/internal/v1/inventory/reservations", input, &response, req.IdempotencyKey); err != nil {
		return nil, fmt.Errorf("reserve product inventory: %w", err)
	}
	return &usecase.ReserveInventoryResponse{ReservationID: response.ReservationID, ExpiresAt: response.ExpiresAt}, nil
}

func (c *ProductHTTPClient) ReleaseInventory(ctx context.Context, req usecase.ReleaseInventoryRequest) error {
	path := "/internal/v1/inventory/reservations/" + url.PathEscape(strings.TrimSpace(req.ReservationID)) + "/release"
	return c.post(ctx, path, map[string]string{"reservation_id": req.ReservationID, "reason": req.Reason}, nil, req.IdempotencyKey)
}

func (c *ProductHTTPClient) CommitInventory(ctx context.Context, req usecase.CommitInventoryRequest) error {
	path := "/internal/v1/inventory/reservations/" + url.PathEscape(strings.TrimSpace(req.ReservationID)) + "/commit"
	return c.post(ctx, path, map[string]string{"reservation_id": req.ReservationID}, nil, req.IdempotencyKey)
}

func (c *ProductHTTPClient) post(ctx context.Context, path string, input any, output any, idempotencyKey string) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		request.Header.Set("Idempotency-Key", idempotencyKey)
	}
	return doJSON(c.client, request, output)
}

type PaymentHTTPClient struct {
	baseURL     string
	token       string
	actionTTL   time.Duration
	client      *http.Client
}

func NewPaymentHTTPClient(baseURL string, token string, actionTTL time.Duration, client *http.Client) (*PaymentHTTPClient, error) {
	normalized, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("payment base URL: %w", err)
	}
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("payment internal token is required")
	}
	if actionTTL <= 0 {
		return nil, errors.New("payment action TTL must be greater than zero")
	}
	return &PaymentHTTPClient{baseURL: normalized, token: strings.TrimSpace(token), actionTTL: actionTTL, client: requiredHTTPClient(client)}, nil
}

func (c *PaymentHTTPClient) CreatePaymentIntent(ctx context.Context, req usecase.CreatePaymentIntentRequest) (*usecase.CreatePaymentIntentResponse, error) {
	body, err := json.Marshal(map[string]any{
		"order_id": req.OrderID, "user_id": req.UserID, "amount": req.Amount, "currency": req.Currency,
		"idempotency_key": req.IdempotencyKey, "metadata": map[string]string{"return_url": req.ReturnURL},
	})
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/payment-intents", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Idempotency-Key", req.IdempotencyKey)
	var response paymentIntentResponse
	if err := doJSON(c.client, request, &response); err != nil {
		var httpErr *HTTPError
		if !errors.As(err, &httpErr) || httpErr.StatusCode >= http.StatusInternalServerError {
			return nil, fmt.Errorf("%w: %v", domain.ErrPaymentIntentPendingResolution, err)
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrPaymentIntentCreationFailed, err)
	}
	return &usecase.CreatePaymentIntentResponse{
		PaymentID: response.PaymentID, Status: response.Status, Provider: response.Provider,
		ProviderIntentRef: response.ProviderIntentID, ClientActionToken: response.ClientPayload.actionToken(),
		ExpiresAt: time.Now().UTC().Add(c.actionTTL),
	}, nil
}

func doJSON(client *http.Client, request *http.Request, output any) error {
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return err
	}
	if len(body) > maxResponseBytes {
		return errors.New("downstream response exceeds size limit")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &HTTPError{StatusCode: response.StatusCode, Body: strings.TrimSpace(string(body))}
	}
	if output == nil || len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, output); err != nil {
		return fmt.Errorf("decode downstream response: %w", err)
	}
	return nil
}

func normalizeBaseURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", errors.New("must be an absolute HTTP(S) URL")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func requiredHTTPClient(client *http.Client) *http.Client {
	if client == nil {
		return &http.Client{Timeout: 5 * time.Second}
	}
	return client
}

func productKey(productID string, variantID string) string {
	return strings.TrimSpace(productID) + ":" + strings.TrimSpace(variantID)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func primaryImageURL(images []productImageResponse) string {
	for _, image := range images {
		if image.IsPrimary {
			return image.URL
		}
	}
	if len(images) > 0 {
		return images[0].URL
	}
	return ""
}

type cartResponse struct {
	CartID string `json:"cart_id"`
	UserID *string `json:"user_id"`
	Items []struct {
		ProductID string `json:"product_id"`
		VariantID string `json:"variant_id"`
		Quantity int `json:"quantity"`
	} `json:"items"`
	Totals struct { Currency string `json:"currency"` } `json:"totals"`
}

type productListResponse struct { Products []productResponse `json:"products"` }

type productResponse struct {
	ProductID string `json:"product_id"`
	SellerID string `json:"seller_id"`
	Title string `json:"title"`
	Status string `json:"status"`
	Images []productImageResponse `json:"images"`
	Variants []productVariantResponse `json:"variants"`
}

type productImageResponse struct {
	URL string `json:"url"`
	IsPrimary bool `json:"is_primary"`
}

type productVariantResponse struct {
	VariantID string `json:"variant_id"`
	SKU string `json:"sku"`
	Price struct {
		Amount int64 `json:"amount"`
		Currency string `json:"currency"`
	} `json:"price"`
	AvailableQuantity int64 `json:"available_quantity"`
	Status string `json:"status"`
}

type inventoryReservationResponse struct {
	ReservationID string `json:"reservation_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type paymentIntentResponse struct {
	PaymentID string `json:"payment_id"`
	Provider string `json:"provider"`
	ProviderIntentID string `json:"provider_intent_id"`
	Status string `json:"status"`
	ClientPayload paymentClientPayloadResponse `json:"client_payload"`
}

type paymentClientPayloadResponse struct {
	ClientSecret string `json:"client_secret"`
	ProviderOrderID string `json:"provider_order_id"`
	CheckoutSessionID string `json:"checkout_session_id"`
	RedirectURL string `json:"redirect_url"`
}

func (p paymentClientPayloadResponse) actionToken() string {
	for _, value := range []string{p.ClientSecret, p.CheckoutSessionID, p.ProviderOrderID, p.RedirectURL} {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
