package events

type ProductInteractionPayload struct {
	UserID      string         `json:"user_id"`
	AnonymousID string         `json:"anonymous_id"`
	SessionID   string         `json:"session_id"`
	ProductID   string         `json:"product_id"`
	VariantID   string         `json:"variant_id"`
	CategoryID  string         `json:"category_id"`
	SellerID    string         `json:"seller_id"`
	BrandID     string         `json:"brand_id"`
	Quantity    int            `json:"quantity"`
	Page        string         `json:"page"`
	CartID      string         `json:"cart_id"`
	WishlistID  string         `json:"wishlist_id"`
	Metadata    map[string]any `json:"metadata"`
}

type PurchasePayload struct {
	UserID      string         `json:"user_id"`
	AnonymousID string         `json:"anonymous_id"`
	SessionID   string         `json:"session_id"`
	OrderID     string         `json:"order_id"`
	Items       []PurchaseItem `json:"items"`
	Metadata    map[string]any `json:"metadata"`
}

type PurchaseItem struct {
	ProductID       string         `json:"product_id"`
	VariantID       string         `json:"variant_id"`
	CategoryID      string         `json:"category_id"`
	SellerID        string         `json:"seller_id"`
	BrandID         string         `json:"brand_id"`
	Quantity        int            `json:"quantity"`
	UnitPriceAmount int64          `json:"unit_price_amount"`
	Currency        string         `json:"currency"`
	Metadata        map[string]any `json:"metadata"`
}
