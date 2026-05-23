package repository

import "context"

const (
	ProductDatabaseName          = "product_db"
	CollectionProducts           = "products"
	CollectionCategories         = "categories"
	CollectionBrands             = "brands"
	CollectionInventorySnapshots = "inventory_snapshots"
	CollectionPriceBooks         = "price_books"
)

type CollectionSchemaManager interface {
	EnsureCollections(ctx context.Context) (CollectionSetupResult, error)
	DescribeCollections() []CollectionDescription
}

type CollectionSetupResult struct {
	Database    string
	Collections []CollectionSetupItem
}

type CollectionSetupItem struct {
	Name             string
	Created          bool
	ValidatorApplied bool
	IndexNames       []string
}

type CollectionDescription struct {
	Name         string
	IndexNames   []string
	HasValidator bool
}
