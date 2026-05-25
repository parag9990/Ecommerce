package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/usecase"
)

const productResponseLimitBytes = 2 << 20

type HTTPProductClient struct {
	baseURL *url.URL
	client  *http.Client
}

func NewHTTPProductClient(baseURL string, timeout time.Duration) (*HTTPProductClient, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, errors.New("product base url is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse product base url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("product base url must use http or https")
	}
	if parsed.Host == "" {
		return nil, errors.New("product base url host is required")
	}
	if timeout <= 0 {
		return nil, errors.New("product client timeout must be greater than zero")
	}
	return &HTTPProductClient{
		baseURL: parsed,
		client:  &http.Client{Timeout: timeout},
	}, nil
}

func (c *HTTPProductClient) GetProductForCart(ctx context.Context, productID string) (*usecase.ProductForCart, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, domain.ErrProductIDRequired
	}

	endpoint := c.productURL(productID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build product request: %v", domain.ErrProductUnavailable, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrProductUnavailable, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, domain.ErrProductNotFound
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		return nil, fmt.Errorf("%w: status=%d", domain.ErrProductUnavailable, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, productResponseLimitBytes))
	if err != nil {
		return nil, fmt.Errorf("%w: read product response: %v", domain.ErrProductUnavailable, err)
	}
	product, err := decodeProductForCart(body)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (c *HTTPProductClient) productURL(productID string) string {
	endpoint := *c.baseURL
	basePath := strings.TrimRight(endpoint.Path, "/")
	endpoint.Path = basePath + "/api/v1/products/" + productID
	endpoint.RawQuery = ""
	endpoint.Fragment = ""
	return endpoint.String()
}

type productEnvelope struct {
	Product *productResponse `json:"product"`
	Data    *productResponse `json:"data"`
	Item    *productResponse `json:"item"`
}

type productResponse struct {
	ID        string            `json:"id"`
	ProductID string            `json:"product_id"`
	SellerID  string            `json:"seller_id"`
	Title     string            `json:"title"`
	Name      string            `json:"name"`
	ImageURL  string            `json:"image_url"`
	Images    []imageResponse   `json:"images"`
	Status    string            `json:"status"`
	Variants  []variantResponse `json:"variants"`
}

type imageResponse struct {
	URL string `json:"url"`
}

type variantResponse struct {
	ID                string         `json:"id"`
	VariantID         string         `json:"variant_id"`
	SKU               string         `json:"sku"`
	Attributes        map[string]any `json:"attributes"`
	Price             moneyResponse  `json:"price"`
	UnitPrice         moneyResponse  `json:"unit_price"`
	PriceAmount       *int64         `json:"price_amount"`
	Currency          string         `json:"currency"`
	StockQuantity     *int           `json:"stock_quantity"`
	InventoryQuantity *int           `json:"inventory_quantity"`
	Available         *bool          `json:"available"`
	Enabled           *bool          `json:"enabled"`
	Status            string         `json:"status"`
}

type moneyResponse struct {
	Amount   *int64 `json:"amount"`
	Currency string `json:"currency"`
}

func decodeProductForCart(body []byte) (*usecase.ProductForCart, error) {
	var root productResponse
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("%w: decode product response: %v", domain.ErrProductUnavailable, err)
	}
	if root.identity() == "" {
		var envelope productEnvelope
		if err := json.Unmarshal(body, &envelope); err != nil {
			return nil, fmt.Errorf("%w: decode product envelope: %v", domain.ErrProductUnavailable, err)
		}
		switch {
		case envelope.Product != nil:
			root = *envelope.Product
		case envelope.Data != nil:
			root = *envelope.Data
		case envelope.Item != nil:
			root = *envelope.Item
		}
	}
	if root.identity() == "" {
		return nil, domain.ErrProductNotFound
	}
	product, err := root.toUsecase()
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (p productResponse) identity() string {
	if strings.TrimSpace(p.ProductID) != "" {
		return strings.TrimSpace(p.ProductID)
	}
	return strings.TrimSpace(p.ID)
}

func (p productResponse) toUsecase() (*usecase.ProductForCart, error) {
	imageURL := strings.TrimSpace(p.ImageURL)
	if imageURL == "" && len(p.Images) > 0 {
		imageURL = strings.TrimSpace(p.Images[0].URL)
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		title = strings.TrimSpace(p.Name)
	}
	variants := make([]usecase.ProductVariantForCart, 0, len(p.Variants))
	for _, variant := range p.Variants {
		mapped, err := variant.toUsecase()
		if err != nil {
			return nil, err
		}
		variants = append(variants, mapped)
	}
	return &usecase.ProductForCart{
		ID:       p.identity(),
		SellerID: strings.TrimSpace(p.SellerID),
		Title:    title,
		ImageURL: imageURL,
		Status:   strings.TrimSpace(p.Status),
		Variants: variants,
	}, nil
}

func (v variantResponse) toUsecase() (usecase.ProductVariantForCart, error) {
	amount, currency, err := v.money()
	if err != nil {
		return usecase.ProductVariantForCart{}, err
	}
	stockQuantity := 0
	if v.StockQuantity != nil {
		stockQuantity = *v.StockQuantity
	} else if v.InventoryQuantity != nil {
		stockQuantity = *v.InventoryQuantity
	}
	available := stockQuantity > 0
	if v.Available != nil {
		available = *v.Available
	}
	if v.Enabled != nil {
		available = available && *v.Enabled
	}
	return usecase.ProductVariantForCart{
		ID:            v.identity(),
		SKU:           strings.TrimSpace(v.SKU),
		Attributes:    stringifyAttributes(v.Attributes),
		PriceAmount:   amount,
		Currency:      currency,
		StockQuantity: stockQuantity,
		Available:     available,
		Status:        strings.TrimSpace(v.Status),
	}, nil
}

func (v variantResponse) identity() string {
	if strings.TrimSpace(v.VariantID) != "" {
		return strings.TrimSpace(v.VariantID)
	}
	return strings.TrimSpace(v.ID)
}

func (v variantResponse) money() (int64, string, error) {
	if v.PriceAmount != nil {
		return *v.PriceAmount, strings.TrimSpace(v.Currency), nil
	}
	if v.Price.Amount != nil {
		currency := v.Price.Currency
		if currency == "" {
			currency = v.Currency
		}
		return *v.Price.Amount, strings.TrimSpace(currency), nil
	}
	if v.UnitPrice.Amount != nil {
		currency := v.UnitPrice.Currency
		if currency == "" {
			currency = v.Currency
		}
		return *v.UnitPrice.Amount, strings.TrimSpace(currency), nil
	}
	return 0, "", domain.ErrProductPriceInvalid
}

func stringifyAttributes(input map[string]any) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		key = strings.TrimSpace(key)
		if key == "" || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			output[key] = strings.TrimSpace(typed)
		case float64:
			output[key] = strconv.FormatFloat(typed, 'f', -1, 64)
		case bool:
			output[key] = strconv.FormatBool(typed)
		default:
			output[key] = fmt.Sprint(typed)
		}
	}
	if len(output) == 0 {
		return nil
	}
	return output
}
