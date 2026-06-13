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
	"net/url"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

const defaultProductClientTimeout = 5 * time.Second

type ProductClientConfig struct {
	BaseURL            string
	InternalAuthHeader string
	InternalAuthToken  string
	Timeout            time.Duration
}

type HTTPProductClient struct {
	baseURL            *url.URL
	internalAuthHeader string
	internalAuthToken  string
	httpClient         *http.Client
	logger             *slog.Logger
}

func NewHTTPProductClient(cfg ProductClientConfig, logger *slog.Logger) (*HTTPProductClient, error) {
	rawBaseURL := strings.TrimSpace(cfg.BaseURL)
	if rawBaseURL == "" {
		return nil, errors.New("product service base url is required")
	}
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse product service base url: %w", err)
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, errors.New("product service base url must use http or https")
	}
	if baseURL.Host == "" {
		return nil, errors.New("product service base url host is required")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultProductClientTimeout
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &HTTPProductClient{
		baseURL:            baseURL,
		internalAuthHeader: strings.TrimSpace(cfg.InternalAuthHeader),
		internalAuthToken:  cfg.InternalAuthToken,
		httpClient:         &http.Client{Timeout: timeout},
		logger:             logger,
	}, nil
}

func (c *HTTPProductClient) GetProduct(ctx context.Context, productID string) (domain.Product, error) {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return domain.Product{}, domain.ErrProductNotFound
	}
	req, err := c.newRequest(ctx, http.MethodGet, "/internal/v1/products/"+url.PathEscape(productID), nil, "")
	if err != nil {
		return domain.Product{}, err
	}
	return c.doProduct(req)
}

func (c *HTTPProductClient) UpdateProductStatus(ctx context.Context, update domain.ProductStatusUpdate) (domain.Product, error) {
	update = normalizeStatusUpdate(update)
	body := statusUpdateRequest{
		ExpectedStatus: update.FromStatus.Normalized(),
		Status:         update.ToStatus.Normalized(),
		ReviewID:       update.ReviewID,
		Reason:         update.Reason,
		ActorUserID:    update.ActorUserID,
		SellerID:       update.SellerID,
		RequestID:      update.RequestID,
	}
	req, err := c.newJSONRequest(ctx, http.MethodPatch, "/internal/v1/products/"+url.PathEscape(update.ProductID)+"/status", body, update.RequestID)
	if err != nil {
		return domain.Product{}, err
	}
	return c.doProduct(req)
}

func (c *HTTPProductClient) PublishProduct(ctx context.Context, update domain.ProductStatusUpdate) (domain.Product, error) {
	update = normalizeStatusUpdate(update)
	body := statusUpdateRequest{
		ExpectedStatus: update.FromStatus.Normalized(),
		Status:         domain.ProductStatusPublished,
		ActorUserID:    update.ActorUserID,
		SellerID:       update.SellerID,
		RequestID:      update.RequestID,
	}
	req, err := c.newJSONRequest(ctx, http.MethodPost, "/internal/v1/products/"+url.PathEscape(update.ProductID)+"/publish", body, update.RequestID)
	if err != nil {
		return domain.Product{}, err
	}
	return c.doProduct(req)
}

func (c *HTTPProductClient) UnpublishProduct(ctx context.Context, update domain.ProductStatusUpdate) (domain.Product, error) {
	update = normalizeStatusUpdate(update)
	body := statusUpdateRequest{
		ExpectedStatus: update.FromStatus.Normalized(),
		Status:         domain.ProductStatusUnpublished,
		Reason:         update.Reason,
		ActorUserID:    update.ActorUserID,
		SellerID:       update.SellerID,
		RequestID:      update.RequestID,
		Force:          update.ForceUnpublish,
	}
	req, err := c.newJSONRequest(ctx, http.MethodPost, "/internal/v1/products/"+url.PathEscape(update.ProductID)+"/unpublish", body, update.RequestID)
	if err != nil {
		return domain.Product{}, err
	}
	return c.doProduct(req)
}

func (c *HTTPProductClient) newJSONRequest(ctx context.Context, method string, path string, body any, requestID string) (*http.Request, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	req, err := c.newRequest(ctx, method, path, &buf, requestID)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *HTTPProductClient) newRequest(ctx context.Context, method string, path string, body io.Reader, requestID string) (*http.Request, error) {
	u := *c.baseURL
	u.Path = strings.TrimRight(c.baseURL.Path, "/") + path
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if c.internalAuthHeader != "" && c.internalAuthToken != "" {
		req.Header.Set(c.internalAuthHeader, c.internalAuthToken)
	}
	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}
	return req, nil
}

func (c *HTTPProductClient) doProduct(req *http.Request) (domain.Product, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WarnContext(req.Context(), "cms.product_client.request_failed",
			slog.String("method", req.Method),
			slog.String("url", req.URL.Redacted()),
			slog.String("error", err.Error()),
		)
		return domain.Product{}, fmt.Errorf("%w: %v", domain.ErrProductServiceUnavailable, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return domain.Product{}, fmt.Errorf("%w: %v", domain.ErrProductServiceUnavailable, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.Product{}, mapProductHTTPError(resp.StatusCode, body)
	}
	product, err := decodeProduct(body)
	if err != nil {
		return domain.Product{}, fmt.Errorf("decode product response: %w", err)
	}
	return product.Normalized(), nil
}

type statusUpdateRequest struct {
	ExpectedStatus domain.ProductStatus `json:"expected_status,omitempty"`
	Status         domain.ProductStatus `json:"status"`
	ReviewID       string               `json:"review_id,omitempty"`
	Reason         string               `json:"reason,omitempty"`
	ActorUserID    string               `json:"actor_user_id,omitempty"`
	SellerID       string               `json:"seller_id,omitempty"`
	RequestID      string               `json:"request_id,omitempty"`
	Force          bool                 `json:"force,omitempty"`
}

type productEnvelope struct {
	Product productDTO `json:"product"`
	Data    productDTO `json:"data"`
}

type productDTO struct {
	ID           string               `json:"id"`
	ProductID    string               `json:"product_id"`
	SellerID     string               `json:"seller_id"`
	Title        string               `json:"title"`
	Description  string               `json:"description"`
	CategoryID   string               `json:"category_id"`
	Brand        string               `json:"brand"`
	GenericBrand bool                 `json:"generic_brand"`
	IsGeneric    bool                 `json:"is_generic_brand"`
	Status       domain.ProductStatus `json:"status"`
	Images       json.RawMessage      `json:"images"`
	Variants     []productVariantDTO  `json:"variants"`
	UpdatedAt    time.Time            `json:"updated_at"`
	Moderation   productModerationDTO `json:"moderation"`
}

type productModerationDTO struct {
	Status domain.ProductStatus `json:"status"`
}

type productVariantDTO struct {
	SKU           string   `json:"sku"`
	Price         moneyDTO `json:"price"`
	MRP           moneyDTO `json:"mrp"`
	StockQuantity int64    `json:"stock_quantity"`
}

type moneyDTO struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func decodeProduct(body []byte) (domain.Product, error) {
	var direct productDTO
	if err := json.Unmarshal(body, &direct); err == nil && direct.identity() != "" {
		return direct.domainProduct()
	}

	var envelope productEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return domain.Product{}, err
	}
	switch {
	case envelope.Product.identity() != "":
		return envelope.Product.domainProduct()
	case envelope.Data.identity() != "":
		return envelope.Data.domainProduct()
	default:
		return domain.Product{}, errors.New("product payload missing product id")
	}
}

func (p productDTO) identity() string {
	if strings.TrimSpace(p.ID) != "" {
		return strings.TrimSpace(p.ID)
	}
	return strings.TrimSpace(p.ProductID)
}

func (p productDTO) domainProduct() (domain.Product, error) {
	images, err := decodeImages(p.Images)
	if err != nil {
		return domain.Product{}, err
	}
	status := p.Status.Normalized()
	if status == "" {
		status = p.Moderation.Status.Normalized()
	}
	variants := make([]domain.ProductVariant, 0, len(p.Variants))
	for _, variant := range p.Variants {
		variants = append(variants, domain.ProductVariant{
			SKU: variant.SKU,
			Price: domain.Money{
				Amount:   variant.Price.Amount,
				Currency: variant.Price.Currency,
			},
			MRP: domain.Money{
				Amount:   variant.MRP.Amount,
				Currency: variant.MRP.Currency,
			},
			StockQuantity: variant.StockQuantity,
		})
	}
	return domain.Product{
		ID:           p.identity(),
		SellerID:     p.SellerID,
		Title:        p.Title,
		Description:  p.Description,
		CategoryID:   p.CategoryID,
		Brand:        p.Brand,
		GenericBrand: p.GenericBrand || p.IsGeneric,
		Status:       status,
		Images:       images,
		Variants:     variants,
		UpdatedAt:    p.UpdatedAt,
	}, nil
}

func decodeImages(raw json.RawMessage) ([]domain.ProductImage, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var stringImages []string
	if err := json.Unmarshal(raw, &stringImages); err == nil {
		images := make([]domain.ProductImage, 0, len(stringImages))
		for _, image := range stringImages {
			images = append(images, domain.ProductImage{URL: image})
		}
		return images, nil
	}
	var objectImages []struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(raw, &objectImages); err != nil {
		return nil, err
	}
	images := make([]domain.ProductImage, 0, len(objectImages))
	for _, image := range objectImages {
		images = append(images, domain.ProductImage{URL: image.URL})
	}
	return images, nil
}

func normalizeStatusUpdate(update domain.ProductStatusUpdate) domain.ProductStatusUpdate {
	update.ProductID = strings.TrimSpace(update.ProductID)
	update.SellerID = strings.TrimSpace(update.SellerID)
	update.ActorUserID = strings.TrimSpace(update.ActorUserID)
	update.FromStatus = update.FromStatus.Normalized()
	update.ToStatus = update.ToStatus.Normalized()
	update.ReviewID = strings.TrimSpace(update.ReviewID)
	update.Reason = domain.NormalizeModerationReason(update.Reason)
	update.RequestID = strings.TrimSpace(update.RequestID)
	return update
}

func mapProductHTTPError(statusCode int, body []byte) error {
	message := extractErrorMessage(body)
	switch statusCode {
	case http.StatusNotFound:
		return domain.ErrProductNotFound
	case http.StatusConflict, http.StatusPreconditionFailed:
		if message == "" {
			return domain.ErrInvalidProductTransition
		}
		return fmt.Errorf("%w: %s", domain.ErrInvalidProductTransition, message)
	case http.StatusUnprocessableEntity, http.StatusBadRequest:
		if message == "" {
			return domain.ErrValidationFailed
		}
		return fmt.Errorf("%w: %s", domain.ErrValidationFailed, message)
	case http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return domain.ErrProductServiceUnavailable
	default:
		if statusCode >= 500 {
			return domain.ErrProductServiceUnavailable
		}
		if message != "" {
			return errors.New(message)
		}
		return fmt.Errorf("product service returned status %d", statusCode)
	}
}

func extractErrorMessage(body []byte) string {
	var response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &response); err == nil {
		if strings.TrimSpace(response.Error.Message) != "" {
			return strings.TrimSpace(response.Error.Message)
		}
		if strings.TrimSpace(response.Message) != "" {
			return strings.TrimSpace(response.Message)
		}
	}
	return strings.TrimSpace(string(body))
}
