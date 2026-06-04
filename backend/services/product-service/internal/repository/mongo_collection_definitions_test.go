package repository

import "testing"

func TestProductMongoCollectionDefinitionsCoverTask3Collections(t *testing.T) {
	definitions := ProductMongoCollectionDefinitions()
	if len(definitions) != 7 {
		t.Fatalf("definition count = %d, want 7", len(definitions))
	}

	want := []string{
		CollectionProducts,
		CollectionCategories,
		CollectionBrands,
		CollectionInventoryReservations,
		CollectionInventorySnapshots,
		CollectionPriceBooks,
		CollectionProductEventOutbox,
	}
	for _, name := range want {
		definition := findDefinition(definitions, name)
		if definition == nil {
			t.Fatalf("missing collection definition %q", name)
		}
		if len(definition.Validator) == 0 {
			t.Fatalf("%s validator is empty", name)
		}
		if len(definition.Indexes) == 0 {
			t.Fatalf("%s indexes are empty", name)
		}
	}
}

func TestProductMongoCollectionDefinitionsExposeRequiredIndexes(t *testing.T) {
	definitions := ProductMongoCollectionDefinitions()
	assertIndexes(t, definitions, CollectionProducts,
		"idx_products_seller_status_updated",
		"idx_products_status_published",
		"idx_products_category_status_updated",
		"idx_products_category_status_published",
		"idx_products_category_path_status_updated",
		"idx_products_category_path_ids_status_updated",
		"idx_products_brand_status_updated",
		"idx_products_status_price_amount",
		"idx_products_status_rating",
		"uniq_products_variant_sku",
		"uniq_products_slug",
		"idx_products_dev_text_search",
	)
	assertIndexes(t, definitions, CollectionCategories,
		"idx_categories_parent_sort",
		"uniq_categories_slug",
		"idx_categories_path",
		"idx_categories_active_sort",
		"idx_categories_active_parent_sort_name",
	)
	assertIndexes(t, definitions, CollectionBrands,
		"uniq_brands_slug",
		"idx_brands_status_name",
		"idx_brands_name_text",
	)
	assertIndexes(t, definitions, CollectionInventoryReservations,
		"uq_inventory_reservations_order",
		"uq_inventory_reservations_idempotency",
		"idx_inventory_reservations_status_expires",
		"ttl_inventory_reservations_cleanup",
	)
	assertIndexes(t, definitions, CollectionInventorySnapshots,
		"idx_inventory_product_variant_created",
		"idx_inventory_seller_created",
		"idx_inventory_sku_created",
		"idx_inventory_type_created",
	)
	assertIndexes(t, definitions, CollectionPriceBooks,
		"idx_price_books_seller_status_starts",
		"idx_price_books_active_window_priority",
		"idx_price_books_entry_product_variant_status",
		"idx_price_books_entry_sku_status",
	)
	assertIndexes(t, definitions, CollectionProductEventOutbox,
		"idx_product_event_outbox_status_next_occurred",
		"idx_product_event_outbox_type_occurred",
		"ttl_product_event_outbox_published",
	)
}

func assertIndexes(t *testing.T, definitions []MongoCollectionDefinition, collection string, want ...string) {
	t.Helper()
	definition := findDefinition(definitions, collection)
	if definition == nil {
		t.Fatalf("missing collection definition %q", collection)
	}
	got := make(map[string]struct{}, len(definition.Indexes))
	for _, index := range definition.Indexes {
		got[index.Name] = struct{}{}
		if index.Model.Keys == nil {
			t.Fatalf("%s index %s has nil keys", collection, index.Name)
		}
		if index.Model.Options == nil {
			t.Fatalf("%s index %s has nil options", collection, index.Name)
		}
	}
	for _, name := range want {
		if _, ok := got[name]; !ok {
			t.Fatalf("%s missing index %s", collection, name)
		}
	}
}

func findDefinition(definitions []MongoCollectionDefinition, name string) *MongoCollectionDefinition {
	for index := range definitions {
		if definitions[index].Name == name {
			return &definitions[index]
		}
	}
	return nil
}
