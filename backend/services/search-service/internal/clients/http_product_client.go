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

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/requestctx"
)

const (
	defaultProductBatchGetPath  = "/internal/v1/products:batchGet"
	defaultProductClientTimeout = 300 * time.Millisecond
	maxProductResponseBytes     = 4 << 20
)

type HTTPProductClientConfig struct {
	BaseURL      string
	BatchGetPath string
	Timeout      time.Duration
}

type HTTPProductClient struct {
	baseURL      *url.URL
	batchGetPath string
	httpClient   *http.Client
}

func NewHTTPProductClient(cfg HTTPProductClientConfig, httpClient *http.Client) (*HTTPProductClient, error) {
	baseURL, err := url.Parse(strings.TrimSpace(cfg.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("parse product service url: %w", err)
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, errors.New("PRODUCT_SERVICE_URL must use http or https")
	}
	if strings.TrimSpace(baseURL.Host) == "" {
		return nil, errors.New("PRODUCT_SERVICE_URL host is required")
	}
	if strings.TrimSpace(cfg.BatchGetPath) == "" {
		cfg.BatchGetPath = defaultProductBatchGetPath
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultProductClientTimeout
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.Timeout}
	}
	return &HTTPProductClient{
		baseURL:      baseURL,
		batchGetPath: cfg.BatchGetPath,
		httpClient:   httpClient,
	}, nil
}

func (c *HTTPProductClient) BatchGetProducts(ctx context.Context, ids []string) ([]domain.ProductSummary, error) {
	ids = cleanIDs(ids)
	if len(ids) == 0 {
		return []domain.ProductSummary{}, nil
	}

	body, err := json.Marshal(batchGetProductsRequest{IDs: ids})
	if err != nil {
		return nil, fmt.Errorf("%w: encode request: %v", domain.ErrProductHydrationUnavailable, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", domain.ErrProductHydrationUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if requestID := requestctx.RequestID(ctx); requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: call product service: %v", domain.ErrProductHydrationUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%w: product service returned status %d", domain.ErrProductHydrationUnavailable, resp.StatusCode)
	}

	var decoded productListResponse
	decoder := json.NewDecoder(io.LimitReader(resp.Body, maxProductResponseBytes))
	if err := decoder.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("%w: decode response: %v", domain.ErrProductHydrationUnavailable, err)
	}

	products := make([]domain.ProductSummary, 0, len(decoded.Products))
	for _, product := range decoded.Products {
		summary := product.toDomain()
		if strings.TrimSpace(summary.ProductID) == "" {
			continue
		}
		products = append(products, summary)
	}
	return products, nil
}

func (c *HTTPProductClient) endpoint() string {
	endpoint := *c.baseURL
	basePath := strings.TrimRight(endpoint.Path, "/")
	batchPath := "/" + strings.TrimLeft(c.batchGetPath, "/")
	endpoint.Path = basePath + batchPath
	return endpoint.String()
}

type batchGetProductsRequest struct {
	IDs []string `json:"ids"`
}

type productListResponse struct {
	Products []productDTO `json:"products"`
}

type productDTO struct {
	ID          string              `json:"id,omitempty"`
	ProductID   string              `json:"product_id"`
	SellerID    string              `json:"seller_id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Brand       string              `json:"brand"`
	CategoryID  string              `json:"category_id"`
	Status      string              `json:"status"`
	Variants    []productVariantDTO `json:"variants"`
}

type productVariantDTO struct {
	SKU           string         `json:"sku"`
	Attributes    map[string]any `json:"attributes"`
	Price         *moneyDTO      `json:"price"`
	StockQuantity int            `json:"stock_quantity"`
}

type moneyDTO struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func (p productDTO) toDomain() domain.ProductSummary {
	productID := strings.TrimSpace(p.ProductID)
	if productID == "" {
		productID = strings.TrimSpace(p.ID)
	}
	variants := make([]domain.ProductVariant, 0, len(p.Variants))
	for _, variant := range p.Variants {
		var price *domain.Money
		if variant.Price != nil {
			price = &domain.Money{
				Amount:   variant.Price.Amount,
				Currency: variant.Price.Currency,
			}
		}
		variants = append(variants, domain.ProductVariant{
			SKU:           variant.SKU,
			Attributes:    variant.Attributes,
			Price:         price,
			StockQuantity: variant.StockQuantity,
		})
	}
	return domain.ProductSummary{
		ProductID:   productID,
		SellerID:    p.SellerID,
		Title:       p.Title,
		Description: p.Description,
		Brand:       p.Brand,
		CategoryID:  p.CategoryID,
		Status:      p.Status,
		Variants:    variants,
	}
}

func cleanIDs(ids []string) []string {
	cleaned := make([]string, 0, len(ids))
	seen := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		cleaned = append(cleaned, id)
	}
	return cleaned
}
