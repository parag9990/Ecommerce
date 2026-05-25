package domain

const (
	CartDatabaseName   = "cart_db"
	CartCollectionName = "carts"
)

type CollectionSpec struct {
	DatabaseName   string
	CollectionName string
	MaxItems       int
	Statuses       []CartStatus
	Indexes        []IndexSpec
}

type IndexSpec struct {
	Name    string
	Purpose string
	Unique  bool
	TTL     bool
}

type CollectionReport struct {
	DatabaseName      string
	CollectionName    string
	CollectionCreated bool
	ValidatorUpdated  bool
	IndexesEnsured    []string
}

func CartCollectionSpec() CollectionSpec {
	return CollectionSpec{
		DatabaseName:   CartDatabaseName,
		CollectionName: CartCollectionName,
		MaxItems:       MaxUniqueItemsPerCart,
		Statuses: []CartStatus{
			CartStatusActive,
			CartStatusMerged,
			CartStatusCheckedOut,
			CartStatusExpired,
			CartStatusAbandoned,
		},
		Indexes: []IndexSpec{
			{Name: "idx_carts_user_status", Purpose: "Logged-in active cart fast lookup"},
			{Name: "idx_carts_guest_status", Purpose: "Guest active cart fast lookup"},
			{Name: "idx_carts_expires_at_ttl", Purpose: "Expired cart cleanup support", TTL: true},
			{Name: "uniq_active_cart_per_user", Purpose: "Block duplicate active carts per user", Unique: true},
			{Name: "uniq_active_cart_per_guest_session", Purpose: "Block duplicate active carts per guest session", Unique: true},
			{Name: "idx_carts_item_product_variant", Purpose: "Product and variant item lookup/debugging"},
			{Name: "idx_carts_updated_at", Purpose: "Abandoned cart and cleanup scans"},
		},
	}
}
