package httptransport

import (
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/usecase"
)

type addItemRequest struct {
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

type updateItemRequest struct {
	Quantity int `json:"quantity"`
}

type couponPreviewRequest struct {
	CouponCode string `json:"coupon_code"`
	CartID     string `json:"cart_id,omitempty"`
	OrderID    string `json:"order_id,omitempty"`
}

type mergeCartRequest struct {
	GuestCartID string `json:"guest_cart_id"`
}

type cartResponse struct {
	CartID         string              `json:"cart_id"`
	UserID         *string             `json:"user_id,omitempty"`
	GuestSessionID *string             `json:"guest_session_id,omitempty"`
	Status         string              `json:"status"`
	Items          []cartItemResponse  `json:"items"`
	CouponCode     *string             `json:"coupon_code,omitempty"`
	CouponPreview  *couponViewResponse `json:"coupon_preview,omitempty"`
	Subtotal       moneyResponse       `json:"subtotal"`
	Discount       moneyResponse       `json:"discount"`
	Total          moneyResponse       `json:"total"`
	Totals         cartTotalsResponse  `json:"totals"`
	Version        int64               `json:"version"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	ExpiresAt      time.Time           `json:"expires_at"`
}

type cartItemResponse struct {
	ItemID           string            `json:"item_id"`
	ProductID        string            `json:"product_id"`
	VariantID        string            `json:"variant_id"`
	SellerID         string            `json:"seller_id"`
	SKUSnapshot      *string           `json:"sku_snapshot,omitempty"`
	TitleSnapshot    string            `json:"title_snapshot"`
	ImageURLSnapshot *string           `json:"image_url_snapshot,omitempty"`
	VariantSnapshot  map[string]string `json:"variant_snapshot,omitempty"`
	UnitPrice        moneyResponse     `json:"unit_price"`
	Quantity         int               `json:"quantity"`
	LineSubtotal     moneyResponse     `json:"line_subtotal"`
	PriceSnapshotAt  time.Time         `json:"price_snapshot_at"`
	AddedAt          time.Time         `json:"added_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type cartTotalsResponse struct {
	Subtotal        moneyResponse `json:"subtotal"`
	Discount        moneyResponse `json:"discount"`
	Total           moneyResponse `json:"total"`
	Currency        string        `json:"currency"`
	ItemCount       int           `json:"item_count"`
	UniqueItemCount int           `json:"unique_item_count"`
}

type couponViewResponse struct {
	CouponID string        `json:"coupon_id"`
	Code     string        `json:"code"`
	Valid    bool          `json:"valid"`
	Discount moneyResponse `json:"discount"`
	Reason   string        `json:"reason,omitempty"`
}

type couponPreviewResponse struct {
	Valid    bool          `json:"valid"`
	CouponID string        `json:"coupon_id"`
	Discount moneyResponse `json:"discount"`
	Reason   string        `json:"reason"`
}

type moneyResponse struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type collectionSpecResponse struct {
	DatabaseName   string          `json:"database_name"`
	CollectionName string          `json:"collection_name"`
	MaxItems       int             `json:"max_items"`
	Statuses       []string        `json:"statuses"`
	Indexes        []indexResponse `json:"indexes"`
}

type indexResponse struct {
	Name    string `json:"name"`
	Purpose string `json:"purpose"`
	Unique  bool   `json:"unique"`
	TTL     bool   `json:"ttl"`
}

type collectionReportResponse struct {
	DatabaseName      string   `json:"database_name"`
	CollectionName    string   `json:"collection_name"`
	CollectionCreated bool     `json:"collection_created"`
	ValidatorUpdated  bool     `json:"validator_updated"`
	IndexesEnsured    []string `json:"indexes_ensured"`
}

type healthResponse struct {
	Status string `json:"status"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func newCartResponse(cart *domain.Cart) cartResponse {
	items := make([]cartItemResponse, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, cartItemResponse{
			ItemID:           item.ItemID,
			ProductID:        item.ProductID,
			VariantID:        item.VariantID,
			SellerID:         item.SellerID,
			SKUSnapshot:      item.SKUSnapshot,
			TitleSnapshot:    item.TitleSnapshot,
			ImageURLSnapshot: item.ImageURLSnapshot,
			VariantSnapshot:  item.VariantSnapshot,
			UnitPrice:        newMoneyResponse(item.UnitPrice),
			Quantity:         item.Quantity,
			LineSubtotal:     newMoneyResponse(item.LineSubtotal),
			PriceSnapshotAt:  item.PriceSnapshotAt,
			AddedAt:          item.AddedAt,
			UpdatedAt:        item.UpdatedAt,
		})
	}
	totals := cartTotalsResponse{
		Subtotal:        newMoneyResponse(cart.Totals.Subtotal),
		Discount:        newMoneyResponse(cart.Totals.Discount),
		Total:           newMoneyResponse(cart.Totals.Total),
		Currency:        cart.Totals.Currency,
		ItemCount:       cart.Totals.ItemCount,
		UniqueItemCount: cart.Totals.UniqueItemCount,
	}
	return cartResponse{
		CartID:         cart.ID,
		UserID:         cart.UserID,
		GuestSessionID: cart.GuestSessionID,
		Status:         string(cart.Status),
		Items:          items,
		CouponCode:     cart.CouponCode,
		CouponPreview:  newCouponViewResponse(cart.CouponPreview),
		Subtotal:       totals.Subtotal,
		Discount:       totals.Discount,
		Total:          totals.Total,
		Totals:         totals,
		Version:        cart.Version,
		CreatedAt:      cart.CreatedAt,
		UpdatedAt:      cart.UpdatedAt,
		ExpiresAt:      cart.ExpiresAt,
	}
}

func newEmptyCartResponse(owner domain.CartOwner) cartResponse {
	totals := cartTotalsResponse{
		Subtotal:        newMoneyResponse(domain.NewMoney(0, domain.CurrencyINR)),
		Discount:        newMoneyResponse(domain.NewMoney(0, domain.CurrencyINR)),
		Total:           newMoneyResponse(domain.NewMoney(0, domain.CurrencyINR)),
		Currency:        domain.CurrencyINR,
		ItemCount:       0,
		UniqueItemCount: 0,
	}
	response := cartResponse{
		Status:   string(domain.CartStatusActive),
		Items:    []cartItemResponse{},
		Subtotal: totals.Subtotal,
		Discount: totals.Discount,
		Total:    totals.Total,
		Totals:   totals,
	}
	if owner.IsUser() {
		userID := owner.UserID
		response.UserID = &userID
	} else {
		guestSessionID := owner.GuestSessionID
		response.GuestSessionID = &guestSessionID
	}
	return response
}

func newCouponViewResponse(preview *domain.CouponView) *couponViewResponse {
	if preview == nil {
		return nil
	}
	return &couponViewResponse{
		CouponID: preview.CouponID,
		Code:     preview.Code,
		Valid:    preview.Valid,
		Discount: newMoneyResponse(preview.Discount),
		Reason:   preview.Reason,
	}
}

func newCouponPreviewResponse(preview *usecase.CouponPreviewResult) couponPreviewResponse {
	if preview == nil {
		return couponPreviewResponse{
			Discount: newMoneyResponse(domain.NewMoney(0, domain.CurrencyINR)),
		}
	}
	return couponPreviewResponse{
		Valid:    preview.Valid,
		CouponID: preview.CouponID,
		Discount: newMoneyResponse(preview.Discount),
		Reason:   preview.Reason,
	}
}

func newMoneyResponse(money domain.Money) moneyResponse {
	return moneyResponse{
		Amount:   money.Amount,
		Currency: money.Currency,
	}
}
