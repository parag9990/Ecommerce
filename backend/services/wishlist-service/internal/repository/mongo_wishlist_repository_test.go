package repository

import (
	"testing"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestWishlistCollectionValidatorMatchesTask3Contract(t *testing.T) {
	validator := WishlistCollectionValidator()
	schema := mustDocument(t, documentValue(t, validator, "$jsonSchema"))
	required := mustArray(t, documentValue(t, schema, "required"))
	assertArrayContains(t, required, "_id")
	assertArrayContains(t, required, "user_id")
	assertArrayContains(t, required, "visibility")
	assertArrayContains(t, required, "items")
	assertArrayContains(t, required, "created_at")
	assertArrayContains(t, required, "updated_at")

	properties := mustDocument(t, documentValue(t, schema, "properties"))
	visibility := mustDocument(t, documentValue(t, properties, "visibility"))
	visibilityEnum := mustArray(t, documentValue(t, visibility, "enum"))
	if len(visibilityEnum) != 1 || visibilityEnum[0] != "private" {
		t.Fatalf("visibility enum = %#v, want only private", visibilityEnum)
	}

	items := mustDocument(t, documentValue(t, properties, "items"))
	itemSchema := mustDocument(t, documentValue(t, items, "items"))
	itemRequired := mustArray(t, documentValue(t, itemSchema, "required"))
	assertArrayContains(t, itemRequired, "product_id")
	assertArrayContains(t, itemRequired, "added_at")

	itemProperties := mustDocument(t, documentValue(t, itemSchema, "properties"))
	price := mustDocument(t, documentValue(t, itemProperties, "last_known_price"))
	priceProperties := mustDocument(t, documentValue(t, price, "properties"))
	amount := mustDocument(t, documentValue(t, priceProperties, "amount"))
	if got := documentValue(t, amount, "bsonType"); got != "long" {
		t.Fatalf("last_known_price.amount bsonType = %v, want long", got)
	}

	availability := mustDocument(t, documentValue(t, itemProperties, "availability"))
	availabilityEnum := mustArray(t, documentValue(t, availability, "enum"))
	assertArrayContains(t, availabilityEnum, "unknown")
	assertArrayContains(t, availabilityEnum, "in_stock")
	assertArrayContains(t, availabilityEnum, "out_of_stock")
	assertArrayContains(t, availabilityEnum, "deleted")
}

func TestWishlistIndexModelsMatchTask3Contract(t *testing.T) {
	models := WishlistIndexModels()
	if len(models) != 4 {
		t.Fatalf("len(WishlistIndexModels()) = %d, want 4", len(models))
	}

	userIndexKeys := mustDocument(t, models[0].Keys)
	if got := documentValue(t, userIndexKeys, "user_id"); got != 1 {
		t.Fatalf("user index key = %v, want 1", got)
	}
	userIndexOptions := materializeIndexOptions(t, models[0].Options)
	if userIndexOptions.Name == nil || *userIndexOptions.Name != UniqueUserIDIndexName {
		t.Fatalf("user index name = %v, want %s", userIndexOptions.Name, UniqueUserIDIndexName)
	}
	if userIndexOptions.Unique == nil || !*userIndexOptions.Unique {
		t.Fatal("user index unique option is not true")
	}

	productIndexKeys := mustDocument(t, models[1].Keys)
	if got := documentValue(t, productIndexKeys, "items.product_id"); got != 1 {
		t.Fatalf("product index key = %v, want 1", got)
	}
	productIndexOptions := materializeIndexOptions(t, models[1].Options)
	if productIndexOptions.Name == nil || *productIndexOptions.Name != ItemsProductIDIndexName {
		t.Fatalf("product index name = %v, want %s", productIndexOptions.Name, ItemsProductIDIndexName)
	}

	variantIndexKeys := mustDocument(t, models[2].Keys)
	if got := documentValue(t, variantIndexKeys, "items.product_id"); got != 1 {
		t.Fatalf("variant index product key = %v, want 1", got)
	}
	if got := documentValue(t, variantIndexKeys, "items.variant_id"); got != 1 {
		t.Fatalf("variant index variant key = %v, want 1", got)
	}
	variantIndexOptions := materializeIndexOptions(t, models[2].Options)
	if variantIndexOptions.Name == nil || *variantIndexOptions.Name != ItemsProductVariantIndexName {
		t.Fatalf("variant index name = %v, want %s", variantIndexOptions.Name, ItemsProductVariantIndexName)
	}

	priceIndexKeys := mustDocument(t, models[3].Keys)
	if got := documentValue(t, priceIndexKeys, "items.product_id"); got != 1 {
		t.Fatalf("price index product key = %v, want 1", got)
	}
	if got := documentValue(t, priceIndexKeys, "items.variant_id"); got != 1 {
		t.Fatalf("price index variant key = %v, want 1", got)
	}
	if got := documentValue(t, priceIndexKeys, "items.last_known_price.currency"); got != 1 {
		t.Fatalf("price index currency key = %v, want 1", got)
	}
	if got := documentValue(t, priceIndexKeys, "items.last_known_price.amount"); got != 1 {
		t.Fatalf("price index amount key = %v, want 1", got)
	}
	if got := documentValue(t, priceIndexKeys, "items.availability"); got != 1 {
		t.Fatalf("price index availability key = %v, want 1", got)
	}
	priceIndexOptions := materializeIndexOptions(t, models[3].Options)
	if priceIndexOptions.Name == nil || *priceIndexOptions.Name != PriceDropCandidateIndexName {
		t.Fatalf("price index name = %v, want %s", priceIndexOptions.Name, PriceDropCandidateIndexName)
	}
}

func TestAvailabilityUpdateDocumentsBuildsProductLevelUpdate(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 30, 0, 0, time.UTC)
	filter, update, arrayFilters, err := availabilityUpdateDocuments(" prod_123 ", "", domain.AvailabilityDeleted, now)
	if err != nil {
		t.Fatalf("availabilityUpdateDocuments returned error: %v", err)
	}

	items := mustMap(t, filter["items"])
	elemMatch := mustMap(t, items["$elemMatch"])
	if got := elemMatch["product_id"]; got != "prod_123" {
		t.Fatalf("elemMatch product_id = %v, want prod_123", got)
	}
	if got := mustMap(t, elemMatch["added_at"])["$lte"]; got != now {
		t.Fatalf("elemMatch added_at <= %v, want %v", got, now)
	}
	set := mustMap(t, update["$set"])
	if got := set["items.$[item].availability"]; got != domain.AvailabilityDeleted {
		t.Fatalf("availability update = %v, want deleted", got)
	}
	if got := set["updated_at"]; got != now {
		t.Fatalf("updated_at = %v, want %v", got, now)
	}
	if len(arrayFilters) != 1 {
		t.Fatalf("len(arrayFilters) = %d, want 1", len(arrayFilters))
	}
	itemFilter := mustMap(t, arrayFilters[0])
	if got := itemFilter["item.product_id"]; got != "prod_123" {
		t.Fatalf("item filter product = %v, want prod_123", got)
	}
	if got := mustMap(t, itemFilter["item.added_at"])["$lte"]; got != now {
		t.Fatalf("item filter added_at <= %v, want %v", got, now)
	}
	if _, exists := itemFilter["item.variant_id"]; exists {
		t.Fatal("product-level array filter unexpectedly included variant_id")
	}
}

func TestAvailabilityUpdateDocumentsBuildsVariantLevelUpdate(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 30, 0, 0, time.UTC)
	filter, _, arrayFilters, err := availabilityUpdateDocuments("prod_123", " var_1 ", domain.AvailabilityOutOfStock, now)
	if err != nil {
		t.Fatalf("availabilityUpdateDocuments returned error: %v", err)
	}

	items := mustMap(t, filter["items"])
	elemMatch := mustMap(t, items["$elemMatch"])
	if got := elemMatch["product_id"]; got != "prod_123" {
		t.Fatalf("elemMatch product_id = %v, want prod_123", got)
	}
	if got := elemMatch["variant_id"]; got != "var_1" {
		t.Fatalf("elemMatch variant_id = %v, want var_1", got)
	}
	if got := mustMap(t, elemMatch["added_at"])["$lte"]; got != now {
		t.Fatalf("elemMatch added_at <= %v, want %v", got, now)
	}
	itemFilter := mustMap(t, arrayFilters[0])
	if got := itemFilter["item.variant_id"]; got != "var_1" {
		t.Fatalf("item filter variant_id = %v, want var_1", got)
	}
	if got := mustMap(t, itemFilter["item.added_at"])["$lte"]; got != now {
		t.Fatalf("item filter added_at <= %v, want %v", got, now)
	}
}

func TestAvailabilityUpdateDocumentsRejectsInvalidInput(t *testing.T) {
	if _, _, _, err := availabilityUpdateDocuments("", "", domain.AvailabilityDeleted, time.Now()); err == nil {
		t.Fatal("availabilityUpdateDocuments empty productID error = nil, want error")
	}
	if _, _, _, err := availabilityUpdateDocuments("prod_123", "", domain.Availability("bad"), time.Now()); err == nil {
		t.Fatal("availabilityUpdateDocuments invalid availability error = nil, want error")
	}
}

func TestPriceDropCandidateDocumentsBuildsProductLevelQuery(t *testing.T) {
	filter, projection, err := priceDropCandidateDocuments(" prod_123 ", "", domain.Money{Amount: 299900, Currency: "INR"})
	if err != nil {
		t.Fatalf("priceDropCandidateDocuments returned error: %v", err)
	}

	items := mustMap(t, filter["items"])
	elemMatch := mustMap(t, items["$elemMatch"])
	if got := elemMatch["product_id"]; got != "prod_123" {
		t.Fatalf("elemMatch product_id = %v, want prod_123", got)
	}
	if got := elemMatch["last_known_price.currency"]; got != "INR" {
		t.Fatalf("elemMatch currency = %v, want INR", got)
	}
	if got := mustMap(t, elemMatch["last_known_price.amount"])["$gt"]; got != int64(299900) {
		t.Fatalf("elemMatch amount gt = %v, want 299900", got)
	}
	if got := mustMap(t, elemMatch["availability"])["$ne"]; got != domain.AvailabilityDeleted {
		t.Fatalf("elemMatch availability ne = %v, want deleted", got)
	}
	if _, exists := elemMatch["variant_id"]; exists {
		t.Fatal("product-level candidate query unexpectedly included variant_id")
	}
	projectedItems := mustMap(t, projection["items"])
	if _, ok := projectedItems["$elemMatch"]; !ok {
		t.Fatalf("projection items = %#v, want $elemMatch", projectedItems)
	}
}

func TestPriceDropCandidateDocumentsBuildsVariantLevelQuery(t *testing.T) {
	filter, _, err := priceDropCandidateDocuments("prod_123", " var_2 ", domain.Money{Amount: 299900, Currency: "INR"})
	if err != nil {
		t.Fatalf("priceDropCandidateDocuments returned error: %v", err)
	}

	elemMatch := mustMap(t, mustMap(t, filter["items"])["$elemMatch"])
	if got := elemMatch["variant_id"]; got != "var_2" {
		t.Fatalf("elemMatch variant_id = %v, want var_2", got)
	}
}

func TestPriceSnapshotUpdateDocumentsBuildsUpdate(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 30, 0, 0, time.UTC)
	filter, update, arrayFilters, err := priceSnapshotUpdateDocuments("prod_123", "var_1", domain.Money{Amount: 299900, Currency: "inr"}, now)
	if err != nil {
		t.Fatalf("priceSnapshotUpdateDocuments returned error: %v", err)
	}

	elemMatch := mustMap(t, mustMap(t, filter["items"])["$elemMatch"])
	if got := elemMatch["product_id"]; got != "prod_123" {
		t.Fatalf("elemMatch product_id = %v, want prod_123", got)
	}
	if got := elemMatch["variant_id"]; got != "var_1" {
		t.Fatalf("elemMatch variant_id = %v, want var_1", got)
	}
	if _, exists := elemMatch["availability"]; exists {
		t.Fatal("price snapshot update unexpectedly filtered by availability")
	}
	if _, exists := elemMatch["added_at"]; exists {
		t.Fatal("price snapshot update unexpectedly filtered by added_at")
	}

	set := mustMap(t, update["$set"])
	price := mustMap(t, set["items.$[item].last_known_price"])
	if got := price["amount"]; got != int64(299900) {
		t.Fatalf("updated amount = %v, want 299900", got)
	}
	if got := price["currency"]; got != "INR" {
		t.Fatalf("updated currency = %v, want INR", got)
	}
	itemFilter := mustMap(t, arrayFilters[0])
	if got := itemFilter["item.variant_id"]; got != "var_1" {
		t.Fatalf("item filter variant_id = %v, want var_1", got)
	}
	if _, exists := itemFilter["item.availability"]; exists {
		t.Fatal("price snapshot item filter unexpectedly included availability")
	}
	if _, exists := itemFilter["item.added_at"]; exists {
		t.Fatal("price snapshot item filter unexpectedly included added_at")
	}
}

func TestPriceSnapshotDocumentsRejectInvalidInput(t *testing.T) {
	if _, _, err := priceDropCandidateDocuments("", "", domain.Money{Amount: 1, Currency: "INR"}); err == nil {
		t.Fatal("priceDropCandidateDocuments empty productID error = nil, want error")
	}
	if _, _, _, err := priceSnapshotUpdateDocuments("prod_123", "", domain.Money{Amount: -1, Currency: "INR"}, time.Now()); err == nil {
		t.Fatal("priceSnapshotUpdateDocuments negative amount error = nil, want error")
	}
}

func TestWishlistDocumentRoundTrip(t *testing.T) {
	createdAt := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	wishlist, err := domain.NewWishlist("wish_123", "user_123", createdAt)
	if err != nil {
		t.Fatalf("NewWishlist returned error: %v", err)
	}
	if err := wishlist.AddItem(domain.AddWishlistItemInput{
		ProductID: "prod_123",
		VariantID: "var_1",
		LastKnownPrice: &domain.Money{
			Amount:   299900,
			Currency: "INR",
		},
		Availability: domain.AvailabilityDeleted,
	}, createdAt.Add(time.Minute)); err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}

	document, err := NewWishlistDocument(wishlist)
	if err != nil {
		t.Fatalf("NewWishlistDocument returned error: %v", err)
	}
	if document.ID != "wish_123" {
		t.Fatalf("document.ID = %q, want wish_123", document.ID)
	}
	if len(document.Items) != 1 {
		t.Fatalf("len(document.Items) = %d, want 1", len(document.Items))
	}
	if document.Items[0].Availability != domain.AvailabilityDeleted {
		t.Fatalf("document item availability = %q, want deleted", document.Items[0].Availability)
	}

	roundTrip, err := document.ToDomain()
	if err != nil {
		t.Fatalf("ToDomain returned error: %v", err)
	}
	item, ok := roundTrip.FindItem("prod_123")
	if !ok {
		t.Fatal("round-trip wishlist missing prod_123")
	}
	if item.LastKnownPrice == nil || item.LastKnownPrice.Amount != 299900 {
		t.Fatalf("round-trip price = %#v, want amount 299900", item.LastKnownPrice)
	}
}

func documentValue(t *testing.T, document bson.D, key string) any {
	t.Helper()
	for _, element := range document {
		if element.Key == key {
			return element.Value
		}
	}
	t.Fatalf("key %q not found in document %#v", key, document)
	return nil
}

func mustDocument(t *testing.T, value any) bson.D {
	t.Helper()
	document, ok := value.(bson.D)
	if !ok {
		t.Fatalf("value %#v has type %T, want bson.D", value, value)
	}
	return document
}

func mustMap(t *testing.T, value any) bson.M {
	t.Helper()
	document, ok := value.(bson.M)
	if !ok {
		t.Fatalf("value %#v has type %T, want bson.M", value, value)
	}
	return document
}

func mustArray(t *testing.T, value any) bson.A {
	t.Helper()
	array, ok := value.(bson.A)
	if !ok {
		t.Fatalf("value %#v has type %T, want bson.A", value, value)
	}
	return array
}

func assertArrayContains(t *testing.T, array bson.A, want any) {
	t.Helper()
	for _, value := range array {
		if value == want {
			return
		}
	}
	t.Fatalf("array %#v does not contain %#v", array, want)
}

func materializeIndexOptions(t *testing.T, builder *options.IndexOptionsBuilder) options.IndexOptions {
	t.Helper()
	var indexOptions options.IndexOptions
	for _, setter := range builder.List() {
		if err := setter(&indexOptions); err != nil {
			t.Fatalf("index option setter returned error: %v", err)
		}
	}
	return indexOptions
}
