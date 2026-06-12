package clients

import (
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

const productResponseBodyLimit = 1 << 20

type HTTPProductValidator struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	logger     *slog.Logger
}

type HTTPProductValidatorOption func(*HTTPProductValidator)

func WithHTTPClient(httpClient *http.Client) HTTPProductValidatorOption {
	return func(v *HTTPProductValidator) {
		if httpClient != nil {
			v.httpClient = httpClient
		}
	}
}

func NewHTTPProductValidator(baseURL string, timeout time.Duration, logger *slog.Logger, options ...HTTPProductValidatorOption) (*HTTPProductValidator, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("product service base URL is required")
	}
	parsed, err := neturl.ParseRequestURI(baseURL)
	if err != nil {
		return nil, fmt.Errorf("product service base URL is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("product service base URL must use http or https")
	}
	if parsed.Host == "" {
		return nil, errors.New("product service base URL host is required")
	}
	if timeout <= 0 {
		return nil, errors.New("product service timeout must be positive")
	}
	if logger == nil {
		logger = slog.Default()
	}

	validator := &HTTPProductValidator{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
		logger:  logger,
	}
	for _, option := range options {
		option(validator)
	}
	return validator, nil
}

func (v *HTTPProductValidator) ValidateWishlistProduct(ctx context.Context, productID, variantID string) (usecase.ProductSnapshot, error) {
	if v == nil || v.httpClient == nil {
		return usecase.ProductSnapshot{}, errors.New("product validator is not initialized")
	}
	productID = strings.TrimSpace(productID)
	variantID = strings.TrimSpace(variantID)
	if productID == "" {
		return usecase.ProductSnapshot{}, usecase.ValidationError{Field: "product_id", Message: "is required"}
	}
	ctx = contextOrBackground(ctx)
	ctx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()

	requestURL := v.baseURL + "/api/v1/products/" + neturl.PathEscape(productID)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return usecase.ProductSnapshot{}, fmt.Errorf("%w: build product request: %v", usecase.ErrProductServiceUnavailable, err)
	}
	request.Header.Set("Accept", "application/json")

	response, err := v.httpClient.Do(request)
	if err != nil {
		return usecase.ProductSnapshot{}, fmt.Errorf("%w: %v", usecase.ErrProductServiceUnavailable, err)
	}
	defer response.Body.Close()

	switch {
	case response.StatusCode == http.StatusNotFound:
		return usecase.ProductSnapshot{}, fmt.Errorf("%w: %s", usecase.ErrProductNotFound, productID)
	case response.StatusCode < 200 || response.StatusCode >= 300:
		return usecase.ProductSnapshot{}, mapProductHTTPStatus(response.StatusCode, productID)
	}

	product, err := decodeProductResponse(response.Body)
	if err != nil {
		return usecase.ProductSnapshot{}, fmt.Errorf("%w: decode product response: %v", usecase.ErrProductServiceUnavailable, err)
	}
	return product.toSnapshot(productID, variantID)
}

type productEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *apiError       `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type productResponse struct {
	ProductID string                   `json:"product_id"`
	ID        string                   `json:"id"`
	MongoID   string                   `json:"_id"`
	Status    string                   `json:"status"`
	Price     *moneyResponse           `json:"price"`
	Variants  []productVariantResponse `json:"variants"`
}

type productVariantResponse struct {
	VariantID        string         `json:"variant_id"`
	ID               string         `json:"id"`
	MongoID          string         `json:"_id"`
	Price            *moneyResponse `json:"price"`
	StockQuantity    *int64         `json:"stock_quantity"`
	ReservedQuantity *int64         `json:"reserved_quantity"`
}

type moneyResponse struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func decodeProductResponse(body io.Reader) (productResponse, error) {
	payload, err := io.ReadAll(io.LimitReader(body, productResponseBodyLimit))
	if err != nil {
		return productResponse{}, err
	}
	var envelope productEnvelope
	if err := json.Unmarshal(payload, &envelope); err == nil && len(envelope.Data) > 0 {
		payload = envelope.Data
	}

	var product productResponse
	if err := json.Unmarshal(payload, &product); err != nil {
		return productResponse{}, err
	}
	return product, nil
}

func (p productResponse) toSnapshot(requestedProductID, requestedVariantID string) (usecase.ProductSnapshot, error) {
	responseProductID := firstNonEmpty(p.ProductID, p.ID, p.MongoID)
	if responseProductID != "" && responseProductID != requestedProductID {
		return usecase.ProductSnapshot{}, fmt.Errorf("%w: product service returned %q for %q", usecase.ErrProductServiceUnavailable, responseProductID, requestedProductID)
	}
	status := strings.ToLower(strings.TrimSpace(p.Status))
	if status != "published" {
		return usecase.ProductSnapshot{}, fmt.Errorf("%w: %s", usecase.ErrProductNotAvailable, requestedProductID)
	}

	variant, foundVariant := p.selectVariant(requestedVariantID)
	if requestedVariantID != "" && !foundVariant {
		return usecase.ProductSnapshot{}, fmt.Errorf("%w: %s", usecase.ErrVariantNotFound, requestedVariantID)
	}

	price := moneyOrNil(p.Price)
	availability := domain.AvailabilityUnknown
	var snapshotVariantID string
	if foundVariant {
		snapshotVariantID = firstNonEmpty(variant.VariantID, variant.ID, variant.MongoID, requestedVariantID)
		if variantPrice := moneyOrNil(variant.Price); variantPrice != nil {
			price = variantPrice
		}
		availability = variant.availability()
	} else {
		price = firstVariantPriceOrNil(p.Variants, price)
		availability = aggregateAvailability(p.Variants)
	}

	return usecase.ProductSnapshot{
		ProductID:      requestedProductID,
		VariantID:      snapshotVariantID,
		LastKnownPrice: price,
		Availability:   availability,
	}, nil
}

func (p productResponse) selectVariant(variantID string) (productVariantResponse, bool) {
	if variantID == "" {
		return productVariantResponse{}, false
	}
	for _, variant := range p.Variants {
		if firstNonEmpty(variant.VariantID, variant.ID, variant.MongoID) == variantID {
			return variant, true
		}
	}
	return productVariantResponse{}, false
}

func firstVariantPriceOrNil(variants []productVariantResponse, fallback *domain.Money) *domain.Money {
	if fallback != nil {
		return fallback
	}
	for _, variant := range variants {
		if price := moneyOrNil(variant.Price); price != nil {
			return price
		}
	}
	return nil
}

func aggregateAvailability(variants []productVariantResponse) domain.Availability {
	if len(variants) == 0 {
		return domain.AvailabilityUnknown
	}
	known := false
	for _, variant := range variants {
		switch variant.availability() {
		case domain.AvailabilityInStock:
			return domain.AvailabilityInStock
		case domain.AvailabilityOutOfStock:
			known = true
		}
	}
	if known {
		return domain.AvailabilityOutOfStock
	}
	return domain.AvailabilityUnknown
}

func (v productVariantResponse) availability() domain.Availability {
	if v.StockQuantity == nil {
		return domain.AvailabilityUnknown
	}
	reserved := int64(0)
	if v.ReservedQuantity != nil {
		reserved = *v.ReservedQuantity
	}
	if *v.StockQuantity-reserved > 0 {
		return domain.AvailabilityInStock
	}
	return domain.AvailabilityOutOfStock
}

func moneyOrNil(money *moneyResponse) *domain.Money {
	if money == nil {
		return nil
	}
	candidate := &domain.Money{
		Amount:   money.Amount,
		Currency: strings.ToUpper(strings.TrimSpace(money.Currency)),
	}
	if err := candidate.Validate(); err != nil {
		return nil
	}
	return candidate
}

func mapProductHTTPStatus(statusCode int, productID string) error {
	switch {
	case statusCode == http.StatusConflict || statusCode == http.StatusGone:
		return fmt.Errorf("%w: %s", usecase.ErrProductNotAvailable, productID)
	case statusCode == http.StatusRequestTimeout || statusCode == http.StatusTooManyRequests || statusCode >= 500:
		return fmt.Errorf("%w: status %d", usecase.ErrProductServiceUnavailable, statusCode)
	default:
		return fmt.Errorf("%w: unexpected status %d", usecase.ErrProductServiceUnavailable, statusCode)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
