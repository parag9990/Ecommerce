package clients

import (
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
	defaultProductSearchExportPath    = "/internal/v1/products/search-export"
	defaultProductSearchExportTimeout = 2 * time.Second
	maxProductExportResponseBytes     = 16 << 20
)

type HTTPProductExportClientConfig struct {
	BaseURL          string
	SearchExportPath string
	Timeout          time.Duration
	ServiceToken     string
}

type HTTPProductExportClient struct {
	baseURL          *url.URL
	searchExportPath string
	httpClient       *http.Client
	serviceToken     string
}

func NewHTTPProductExportClient(cfg HTTPProductExportClientConfig, httpClient *http.Client) (*HTTPProductExportClient, error) {
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
	if strings.TrimSpace(cfg.SearchExportPath) == "" {
		cfg.SearchExportPath = defaultProductSearchExportPath
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultProductSearchExportTimeout
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.Timeout}
	}
	return &HTTPProductExportClient{
		baseURL:          baseURL,
		searchExportPath: cfg.SearchExportPath,
		httpClient:       httpClient,
		serviceToken:     strings.TrimSpace(cfg.ServiceToken),
	}, nil
}

func (c *HTTPProductExportClient) ListSearchableProducts(ctx context.Context, req domain.ProductExportRequest) (domain.ProductExportPage, error) {
	if req.Limit <= 0 {
		return domain.ProductExportPage{}, fmt.Errorf("%w: limit must be greater than zero", domain.ErrInvalidReindexRequest)
	}

	endpoint := c.endpoint()
	values := endpoint.Query()
	values.Set("limit", fmt.Sprintf("%d", req.Limit))
	if cursor := strings.TrimSpace(req.Cursor); cursor != "" {
		values.Set("cursor", cursor)
	}
	endpoint.RawQuery = values.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return domain.ProductExportPage{}, fmt.Errorf("%w: build request: %v", domain.ErrProductCatalogExportUnavailable, err)
	}
	httpReq.Header.Set("Accept", "application/json")
	if c.serviceToken != "" {
		httpReq.Header.Set("X-Service-Token", c.serviceToken)
	}
	if requestID := requestctx.RequestID(ctx); requestID != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return domain.ProductExportPage{}, fmt.Errorf("%w: call product service: %v", domain.ErrProductCatalogExportUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.ProductExportPage{}, fmt.Errorf("%w: product service returned status %d", domain.ErrProductCatalogExportUnavailable, resp.StatusCode)
	}

	var decoded productSearchExportResponse
	decoder := json.NewDecoder(io.LimitReader(resp.Body, maxProductExportResponseBytes))
	if err := decoder.Decode(&decoded); err != nil {
		return domain.ProductExportPage{}, fmt.Errorf("%w: decode response: %v", domain.ErrProductCatalogExportUnavailable, err)
	}
	return decoded.toDomain(), nil
}

func (c *HTTPProductExportClient) endpoint() url.URL {
	endpoint := *c.baseURL
	basePath := strings.TrimRight(endpoint.Path, "/")
	exportPath := "/" + strings.TrimLeft(c.searchExportPath, "/")
	endpoint.Path = basePath + exportPath
	return endpoint
}

type productSearchExportResponse struct {
	Items      []productSearchExportDTO `json:"items"`
	Products   []productSearchExportDTO `json:"products,omitempty"`
	NextCursor string                   `json:"next_cursor"`
	HasMore    bool                     `json:"has_more"`
	Total      int                      `json:"total,omitempty"`
}

func (r productSearchExportResponse) toDomain() domain.ProductExportPage {
	items := r.Items
	if len(items) == 0 && len(r.Products) > 0 {
		items = r.Products
	}
	payloads := make([]domain.ProductIndexPayload, 0, len(items))
	for _, item := range items {
		payload := item.toDomain().Normalized()
		if strings.TrimSpace(payload.ProductID) == "" {
			continue
		}
		payloads = append(payloads, payload)
	}
	return domain.ProductExportPage{
		Items:      payloads,
		NextCursor: strings.TrimSpace(r.NextCursor),
		HasMore:    r.HasMore,
		Total:      r.Total,
	}
}

type productSearchExportDTO struct {
	ID              string     `json:"id,omitempty"`
	ProductID       string     `json:"product_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Brand           string     `json:"brand"`
	CategoryID      string     `json:"category_id,omitempty"`
	CategoryIDs     []string   `json:"category_ids"`
	SellerID        string     `json:"seller_id"`
	Price           *float64   `json:"price"`
	Rating          *float64   `json:"rating"`
	PopularityScore *int32     `json:"popularity_score"`
	InStock         *bool      `json:"in_stock"`
	Status          string     `json:"status"`
	IsDeleted       bool       `json:"is_deleted"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (p productSearchExportDTO) toDomain() domain.ProductIndexPayload {
	productID := strings.TrimSpace(p.ProductID)
	if productID == "" {
		productID = strings.TrimSpace(p.ID)
	}
	categoryIDs := append([]string(nil), p.CategoryIDs...)
	if len(categoryIDs) == 0 && strings.TrimSpace(p.CategoryID) != "" {
		categoryIDs = []string{p.CategoryID}
	}
	return domain.ProductIndexPayload{
		ProductID:       productID,
		Title:           p.Title,
		Description:     p.Description,
		Brand:           p.Brand,
		CategoryIDs:     categoryIDs,
		SellerID:        p.SellerID,
		Price:           p.Price,
		Rating:          p.Rating,
		PopularityScore: p.PopularityScore,
		InStock:         p.InStock,
		Status:          p.Status,
		IsDeleted:       p.IsDeleted,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}
