package client

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

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/usecase"
)

const couponValidationResponseLimitBytes = 1 << 20

type HTTPCouponValidator struct {
	baseURL      *url.URL
	validatePath string
	client       *http.Client
}

func NewHTTPCouponValidator(baseURL string, validatePath string, timeout time.Duration) (*HTTPCouponValidator, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, errors.New("cms base url is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse cms base url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("cms base url must use http or https")
	}
	if parsed.Host == "" {
		return nil, errors.New("cms base url host is required")
	}
	validatePath = strings.TrimSpace(validatePath)
	if validatePath == "" {
		return nil, errors.New("cms validate coupon path is required")
	}
	if !strings.HasPrefix(validatePath, "/") {
		validatePath = "/" + validatePath
	}
	if timeout <= 0 {
		return nil, errors.New("cms coupon validator timeout must be greater than zero")
	}
	return &HTTPCouponValidator{
		baseURL:      parsed,
		validatePath: validatePath,
		client:       &http.Client{Timeout: timeout},
	}, nil
}

func (c *HTTPCouponValidator) ValidateCoupon(ctx context.Context, req usecase.CouponValidationRequest) (*usecase.CouponValidationResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	code, err := domain.NormalizeCouponCode(req.Code)
	if err != nil {
		return nil, err
	}
	payload := couponValidationRequest{
		CouponCode:     code,
		Code:           code,
		UserID:         strings.TrimSpace(req.UserID),
		GuestSessionID: strings.TrimSpace(req.GuestSessionID),
		CartID:         strings.TrimSpace(req.CartID),
		Currency:       strings.ToUpper(strings.TrimSpace(req.Currency)),
		Subtotal:       moneyPayloadFromDomain(req.Subtotal),
		Items:          couponItemsFromUsecase(req.Items),
		RequestedAt:    req.RequestedAt.UTC(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: marshal coupon validation request: %v", domain.ErrCouponPreviewUnavailable, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.validationURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: build coupon validation request: %v", domain.ErrCouponPreviewUnavailable, err)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Internal-Service", "cart-service")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrCouponPreviewUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: cms status=%d", domain.ErrCouponPreviewUnavailable, resp.StatusCode)
	}

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, couponValidationResponseLimitBytes))
	if err != nil {
		return nil, fmt.Errorf("%w: read coupon validation response: %v", domain.ErrCouponPreviewUnavailable, err)
	}
	result, err := decodeCouponValidationResult(responseBody)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *HTTPCouponValidator) validationURL() string {
	endpoint := *c.baseURL
	basePath := strings.TrimRight(endpoint.Path, "/")
	endpoint.Path = basePath + c.validatePath
	endpoint.RawQuery = ""
	endpoint.Fragment = ""
	return endpoint.String()
}

type couponValidationRequest struct {
	CouponCode     string              `json:"coupon_code"`
	Code           string              `json:"code"`
	UserID         string              `json:"user_id,omitempty"`
	GuestSessionID string              `json:"guest_session_id,omitempty"`
	CartID         string              `json:"cart_id"`
	Currency       string              `json:"currency"`
	Subtotal       moneyPayload        `json:"subtotal"`
	Items          []couponItemPayload `json:"items"`
	RequestedAt    time.Time           `json:"requested_at"`
}

type couponItemPayload struct {
	ProductID    string       `json:"product_id"`
	VariantID    string       `json:"variant_id"`
	SellerID     string       `json:"seller_id"`
	UnitPrice    moneyPayload `json:"unit_price"`
	Quantity     int          `json:"quantity"`
	LineSubtotal moneyPayload `json:"line_subtotal"`
}

type moneyPayload struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type couponValidationEnvelope struct {
	Valid    *bool                     `json:"valid"`
	CouponID string                    `json:"coupon_id"`
	Code     string                    `json:"code"`
	Discount moneyPayload              `json:"discount"`
	Reason   string                    `json:"reason"`
	Data     *couponValidationEnvelope `json:"data"`
	Result   *couponValidationEnvelope `json:"result"`
	Preview  *couponValidationEnvelope `json:"preview"`
}

func decodeCouponValidationResult(body []byte) (*usecase.CouponValidationResult, error) {
	var envelope couponValidationEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("%w: decode coupon validation response: %v", domain.ErrCouponPreviewUnavailable, err)
	}
	source := &envelope
	switch {
	case envelope.Valid != nil:
	case envelope.Data != nil:
		source = envelope.Data
	case envelope.Result != nil:
		source = envelope.Result
	case envelope.Preview != nil:
		source = envelope.Preview
	default:
		return nil, fmt.Errorf("%w: coupon validation response missing valid flag", domain.ErrCouponPreviewUnavailable)
	}
	if source.Valid == nil {
		return nil, fmt.Errorf("%w: coupon validation response missing valid flag", domain.ErrCouponPreviewUnavailable)
	}
	return &usecase.CouponValidationResult{
		Valid:    *source.Valid,
		CouponID: strings.TrimSpace(source.CouponID),
		Code:     strings.TrimSpace(source.Code),
		Discount: domain.NewMoney(source.Discount.Amount, source.Discount.Currency),
		Reason:   strings.TrimSpace(source.Reason),
	}, nil
}

func couponItemsFromUsecase(items []usecase.CouponCartItem) []couponItemPayload {
	output := make([]couponItemPayload, 0, len(items))
	for _, item := range items {
		output = append(output, couponItemPayload{
			ProductID:    strings.TrimSpace(item.ProductID),
			VariantID:    strings.TrimSpace(item.VariantID),
			SellerID:     strings.TrimSpace(item.SellerID),
			UnitPrice:    moneyPayloadFromDomain(item.UnitPrice),
			Quantity:     item.Quantity,
			LineSubtotal: moneyPayloadFromDomain(item.LineSubtotal),
		})
	}
	return output
}

func moneyPayloadFromDomain(money domain.Money) moneyPayload {
	normalized := domain.NewMoney(money.Amount, money.Currency)
	return moneyPayload{
		Amount:   normalized.Amount,
		Currency: normalized.Currency,
	}
}
