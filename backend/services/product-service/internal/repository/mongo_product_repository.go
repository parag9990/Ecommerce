package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"product-service/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoProductRepository struct {
	products   *mongo.Collection
	categories *mongo.Collection
	logger     *slog.Logger
}

func NewMongoProductRepository(database *mongo.Database, logger *slog.Logger) (*MongoProductRepository, error) {
	if database == nil {
		return nil, fmt.Errorf("mongo database is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &MongoProductRepository{
		products:   database.Collection(CollectionProducts),
		categories: database.Collection(CollectionCategories),
		logger:     logger,
	}, nil
}

func (r *MongoProductRepository) InsertProduct(ctx context.Context, product *domain.Product) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if product == nil {
		return fmt.Errorf("product is required")
	}
	_, err := r.products.InsertOne(ctx, productDocumentFromDomain(*product))
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateKey
	}
	if err != nil {
		return fmt.Errorf("insert product %q: %w", product.ID, err)
	}
	r.logger.Debug("product document inserted", "product_id", product.ID, "seller_id", product.SellerID)
	return nil
}

func (r *MongoProductRepository) FindProductByID(ctx context.Context, productID string) (*domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var doc productDocument
	err := r.products.FindOne(ctx, bson.D{e("_id", strings.TrimSpace(productID))}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find product %q: %w", productID, err)
	}
	product := doc.toDomain()
	return &product, nil
}

func (r *MongoProductRepository) UpdateProduct(ctx context.Context, product *domain.Product, expectedStatuses ...domain.ProductStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if product == nil {
		return fmt.Errorf("product is required")
	}

	filter := bson.D{
		e("_id", product.ID),
		e("seller_id", product.SellerID),
	}
	if len(expectedStatuses) > 0 {
		statuses := make(bson.A, 0, len(expectedStatuses))
		for _, status := range expectedStatuses {
			statuses = append(statuses, string(status))
		}
		filter = append(filter, e("status", bson.D{e("$in", statuses)}))
	}

	result, err := r.products.ReplaceOne(ctx, filter, productDocumentFromDomain(*product))
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateKey
	}
	if err != nil {
		return fmt.Errorf("replace product %q: %w", product.ID, err)
	}
	if result.MatchedCount == 0 {
		return ErrWriteConflict
	}
	r.logger.Debug("product document updated", "product_id", product.ID, "seller_id", product.SellerID)
	return nil
}

func (r *MongoProductRepository) GetCategoryByID(ctx context.Context, categoryID string) (*domain.Category, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var doc categoryDocument
	err := r.categories.FindOne(ctx, bson.D{e("_id", strings.TrimSpace(categoryID))}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find category %q: %w", categoryID, err)
	}
	category := doc.toDomain()
	return &category, nil
}

func (r *MongoProductRepository) IsSKUUnique(ctx context.Context, sku string, excludeProductID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	filter := bson.D{e("variants.sku", normalizeRepositorySKU(sku))}
	if strings.TrimSpace(excludeProductID) != "" {
		filter = append(filter, e("_id", bson.D{e("$ne", strings.TrimSpace(excludeProductID))}))
	}
	err := r.products.FindOne(
		ctx,
		filter,
		options.FindOne().SetProjection(bson.D{e("_id", 1)}),
	).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("check sku uniqueness %q: %w", sku, err)
	}
	return false, nil
}

func (r *MongoProductRepository) FindProduct(ctx context.Context, filter ProductReadFilter) (*domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(filter.ProductID) == "" {
		return nil, fmt.Errorf("product id is required")
	}

	query := productReadBSONFilter(filter)
	var doc productDocument
	err := r.products.FindOne(ctx, query, options.FindOne().SetProjection(productDetailProjection())).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find readable product %q: %w", filter.ProductID, err)
	}
	product := doc.toDomain()
	return &product, nil
}

func (r *MongoProductRepository) ListProducts(ctx context.Context, filter ProductReadFilter) ([]domain.Product, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	query := productReadBSONFilter(filter)
	total, err := r.products.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("count readable products: %w", err)
	}

	cursor, err := r.products.Find(
		ctx,
		query,
		options.Find().
			SetSort(productReadSortBSON(filter.Sort)).
			SetSkip(int64((page-1)*pageSize)).
			SetLimit(int64(pageSize)).
			SetProjection(productListProjection()),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list readable products: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []productDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("decode readable products: %w", err)
	}

	products := make([]domain.Product, 0, len(docs))
	for _, doc := range docs {
		products = append(products, doc.toDomain())
	}
	return products, total, nil
}

func (r *MongoProductRepository) BatchGetProducts(ctx context.Context, filter ProductReadFilter) ([]domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(filter.ProductIDs) == 0 {
		return []domain.Product{}, nil
	}

	query := productReadBSONFilter(filter)
	cursor, err := r.products.Find(
		ctx,
		query,
		options.Find().SetProjection(productDetailProjection()),
	)
	if err != nil {
		return nil, fmt.Errorf("batch get readable products: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []productDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode batch products: %w", err)
	}

	byID := make(map[string]domain.Product, len(docs))
	for _, doc := range docs {
		product := doc.toDomain()
		byID[product.ID] = product
	}

	products := make([]domain.Product, 0, len(filter.ProductIDs))
	for _, id := range filter.ProductIDs {
		if product, ok := byID[strings.TrimSpace(id)]; ok {
			products = append(products, product)
		}
	}
	return products, nil
}

func (r *MongoProductRepository) ListCategories(ctx context.Context, filter CategoryReadFilter) ([]domain.Category, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	query := bson.D{}
	if filter.ActiveOnly {
		query = append(query, e("is_active", true))
	}
	if filter.ParentID != nil {
		query = append(query, e("parent_id", strings.TrimSpace(*filter.ParentID)))
	}

	cursor, err := r.categories.Find(
		ctx,
		query,
		options.Find().
			SetSort(bson.D{e("sort_order", 1), e("name", 1), e("_id", 1)}).
			SetProjection(categoryBrowseProjection()),
	)
	if err != nil {
		return nil, fmt.Errorf("list readable categories: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []categoryDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode readable categories: %w", err)
	}

	categories := make([]domain.Category, 0, len(docs))
	for _, doc := range docs {
		categories = append(categories, doc.toDomain())
	}
	return categories, nil
}

func productReadBSONFilter(filter ProductReadFilter) bson.D {
	query := bson.D{}
	if productID := strings.TrimSpace(filter.ProductID); productID != "" {
		query = append(query, e("_id", productID))
	}
	if len(filter.ProductIDs) > 0 {
		ids := make(bson.A, 0, len(filter.ProductIDs))
		seen := make(map[string]struct{}, len(filter.ProductIDs))
		for _, id := range filter.ProductIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
		query = append(query, e("_id", bson.D{e("$in", ids)}))
	}
	if sellerID := strings.TrimSpace(filter.SellerID); sellerID != "" {
		query = append(query, e("seller_id", sellerID))
	}
	if len(filter.Statuses) > 0 {
		statuses := make(bson.A, 0, len(filter.Statuses))
		for _, status := range filter.Statuses {
			if strings.TrimSpace(string(status)) != "" {
				statuses = append(statuses, string(status))
			}
		}
		if len(statuses) == 1 {
			query = append(query, e("status", statuses[0]))
		} else if len(statuses) > 1 {
			query = append(query, e("status", bson.D{e("$in", statuses)}))
		}
	}
	if categoryID := strings.TrimSpace(filter.CategoryID); categoryID != "" {
		query = append(query, e("$or", bson.A{
			bson.D{e("category_id", categoryID)},
			bson.D{e("category_path.category_id", categoryID)},
			bson.D{e("category_path", categoryID)},
		}))
	}
	return query
}

func productReadSortBSON(sort ProductReadSort) bson.D {
	switch sort {
	case ProductReadSortNewest:
		return bson.D{e("published_at", -1), e("_id", 1)}
	case ProductReadSortPriceAsc:
		return bson.D{e("variants.price.amount", 1), e("_id", 1)}
	case ProductReadSortPriceDesc:
		return bson.D{e("variants.price.amount", -1), e("_id", 1)}
	case ProductReadSortRatingDesc:
		return bson.D{e("rating_summary.average", -1), e("rating_summary.count", -1), e("_id", 1)}
	default:
		return bson.D{e("updated_at", -1), e("_id", 1)}
	}
}

func productListProjection() bson.D {
	return bson.D{
		e("_id", 1),
		e("seller_id", 1),
		e("title", 1),
		e("slug", 1),
		e("description", 1),
		e("brand_name", 1),
		e("category_id", 1),
		e("category_path", 1),
		e("status", 1),
		e("images", 1),
		e("variants", 1),
		e("rating_summary", 1),
		e("updated_at", 1),
		e("published_at", 1),
	}
}

func productDetailProjection() bson.D {
	projection := productListProjection()
	projection = append(projection,
		e("attributes", 1),
		e("created_at", 1),
	)
	return projection
}

func categoryBrowseProjection() bson.D {
	return bson.D{
		e("_id", 1),
		e("name", 1),
		e("slug", 1),
		e("parent_id", 1),
		e("path", 1),
		e("level", 1),
		e("depth", 1),
		e("sort_order", 1),
		e("is_active", 1),
	}
}

type productDocument struct {
	ID            string            `bson:"_id"`
	SellerID      string            `bson:"seller_id"`
	Title         string            `bson:"title"`
	Slug          string            `bson:"slug,omitempty"`
	Description   string            `bson:"description,omitempty"`
	BrandName     string            `bson:"brand_name,omitempty"`
	CategoryID    string            `bson:"category_id"`
	CategoryPath  []any             `bson:"category_path,omitempty"`
	Status        string            `bson:"status"`
	Attributes    map[string]any    `bson:"attributes,omitempty"`
	Images        []productImageDoc `bson:"images,omitempty"`
	Variants      []variantDocument `bson:"variants"`
	RatingSummary ratingSummaryDoc  `bson:"rating_summary,omitempty"`
	CreatedBy     string            `bson:"created_by,omitempty"`
	UpdatedBy     string            `bson:"updated_by,omitempty"`
	CreatedAt     time.Time         `bson:"created_at"`
	UpdatedAt     time.Time         `bson:"updated_at"`
	PublishedAt   *time.Time        `bson:"published_at,omitempty"`
}

type ratingSummaryDoc struct {
	Average float64 `bson:"average"`
	Count   int64   `bson:"count"`
}

type productImageDoc struct {
	ID         string   `bson:"image_id,omitempty"`
	URL        string   `bson:"url"`
	AltText    string   `bson:"alt_text,omitempty"`
	Alt        string   `bson:"alt,omitempty"`
	Position   int      `bson:"position"`
	IsPrimary  bool     `bson:"is_primary"`
	VariantIDs []string `bson:"variant_ids,omitempty"`
	Width      int      `bson:"width,omitempty"`
	Height     int      `bson:"height,omitempty"`
	Status     string   `bson:"status"`
}

type variantDocument struct {
	ID               string         `bson:"variant_id"`
	SKU              string         `bson:"sku"`
	Title            string         `bson:"title,omitempty"`
	Attributes       map[string]any `bson:"attributes,omitempty"`
	Price            moneyDocument  `bson:"price"`
	MRP              *moneyDocument `bson:"mrp,omitempty"`
	StockQuantity    int64          `bson:"stock_quantity"`
	ReservedQuantity int64          `bson:"reserved_quantity"`
	SafetyStock      int64          `bson:"safety_stock,omitempty"`
	Status           string         `bson:"status"`
	Barcode          string         `bson:"barcode,omitempty"`
}

type moneyDocument struct {
	Amount   int64  `bson:"amount"`
	Currency string `bson:"currency"`
}

func productDocumentFromDomain(product domain.Product) productDocument {
	images := make([]productImageDoc, 0, len(product.Images))
	for _, image := range product.Images {
		images = append(images, productImageDoc{
			ID:         image.ID,
			URL:        image.URL,
			AltText:    image.AltText,
			Position:   image.Position,
			IsPrimary:  image.IsPrimary,
			VariantIDs: image.VariantIDs,
			Width:      image.Width,
			Height:     image.Height,
			Status:     string(image.Status),
		})
	}

	variants := make([]variantDocument, 0, len(product.Variants))
	for _, variant := range product.Variants {
		doc := variantDocument{
			ID:               variant.ID,
			SKU:              normalizeRepositorySKU(variant.SKU),
			Title:            variant.Title,
			Attributes:       map[string]any(variant.Attributes),
			Price:            moneyDocumentFromDomain(variant.Price),
			StockQuantity:    variant.StockQuantity,
			ReservedQuantity: variant.ReservedQuantity,
			SafetyStock:      variant.SafetyStock,
			Status:           string(variant.Status),
			Barcode:          variant.Barcode,
		}
		if variant.MRP != nil {
			mrp := moneyDocumentFromDomain(*variant.MRP)
			doc.MRP = &mrp
		}
		variants = append(variants, doc)
	}

	return productDocument{
		ID:            product.ID,
		SellerID:      product.SellerID,
		Title:         product.Title,
		Slug:          product.Slug,
		Description:   product.Description,
		BrandName:     product.Brand,
		CategoryID:    product.CategoryID,
		CategoryPath:  categoryPathDocumentFromDomain(product.CategoryPath),
		Status:        string(product.Status),
		Attributes:    map[string]any(product.Attributes),
		Images:        images,
		Variants:      variants,
		RatingSummary: ratingSummaryDoc{Average: product.RatingSummary.Average, Count: product.RatingSummary.Count},
		CreatedBy:     product.CreatedBy,
		UpdatedBy:     product.UpdatedBy,
		CreatedAt:     product.CreatedAt,
		UpdatedAt:     product.UpdatedAt,
		PublishedAt:   product.PublishedAt,
	}
}

func (d productDocument) toDomain() domain.Product {
	images := make([]domain.ProductImage, 0, len(d.Images))
	for _, image := range d.Images {
		altText := image.AltText
		if altText == "" {
			altText = image.Alt
		}
		images = append(images, domain.ProductImage{
			ID:         image.ID,
			URL:        image.URL,
			AltText:    altText,
			Position:   image.Position,
			IsPrimary:  image.IsPrimary,
			VariantIDs: image.VariantIDs,
			Width:      image.Width,
			Height:     image.Height,
			Status:     domain.ImageStatus(image.Status),
		})
	}

	variants := make([]domain.Variant, 0, len(d.Variants))
	for _, variant := range d.Variants {
		domainVariant := domain.Variant{
			ID:               variant.ID,
			SKU:              variant.SKU,
			Title:            variant.Title,
			Attributes:       domain.Attributes(variant.Attributes),
			Price:            variant.Price.toDomain(),
			StockQuantity:    variant.StockQuantity,
			ReservedQuantity: variant.ReservedQuantity,
			SafetyStock:      variant.SafetyStock,
			Status:           domain.VariantStatus(variant.Status),
			Barcode:          variant.Barcode,
		}
		if variant.MRP != nil {
			mrp := variant.MRP.toDomain()
			domainVariant.MRP = &mrp
		}
		variants = append(variants, domainVariant)
	}

	return domain.Product{
		ID:            d.ID,
		SellerID:      d.SellerID,
		Title:         d.Title,
		Slug:          d.Slug,
		Description:   d.Description,
		Brand:         d.BrandName,
		CategoryID:    d.CategoryID,
		CategoryPath:  categoryPathToDomain(d.CategoryPath),
		Status:        domain.ProductStatus(d.Status),
		Attributes:    domain.Attributes(d.Attributes),
		Images:        images,
		Variants:      variants,
		RatingSummary: domain.RatingSummary{Average: d.RatingSummary.Average, Count: d.RatingSummary.Count},
		CreatedBy:     d.CreatedBy,
		UpdatedBy:     d.UpdatedBy,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
		PublishedAt:   d.PublishedAt,
	}
}

func moneyDocumentFromDomain(money domain.Money) moneyDocument {
	normalized := domain.NewMoney(money.Amount, money.Currency)
	return moneyDocument{Amount: normalized.Amount, Currency: normalized.Currency}
}

func (d moneyDocument) toDomain() domain.Money {
	return domain.NewMoney(d.Amount, d.Currency)
}

func categoryPathDocumentFromDomain(path []string) []any {
	if len(path) == 0 {
		return nil
	}
	values := make([]any, 0, len(path))
	for _, id := range path {
		id = strings.TrimSpace(id)
		if id != "" {
			values = append(values, id)
		}
	}
	return values
}

func categoryPathToDomain(values []any) []string {
	path := make([]string, 0, len(values))
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			if id := strings.TrimSpace(typed); id != "" {
				path = append(path, id)
			}
		case bson.D:
			if id := categoryIDFromBSOND(typed); id != "" {
				path = append(path, id)
			}
		case bson.M:
			if id, _ := typed["category_id"].(string); strings.TrimSpace(id) != "" {
				path = append(path, strings.TrimSpace(id))
			}
		case map[string]any:
			if id, _ := typed["category_id"].(string); strings.TrimSpace(id) != "" {
				path = append(path, strings.TrimSpace(id))
			}
		}
	}
	return path
}

func categoryIDFromBSOND(value bson.D) string {
	for _, element := range value {
		if element.Key != "category_id" {
			continue
		}
		if id, ok := element.Value.(string); ok {
			return strings.TrimSpace(id)
		}
	}
	return ""
}

type categoryDocument struct {
	ID              string                        `bson:"_id"`
	Name            string                        `bson:"name"`
	Slug            string                        `bson:"slug"`
	ParentID        *string                       `bson:"parent_id,omitempty"`
	Path            []string                      `bson:"path"`
	Level           int                           `bson:"level,omitempty"`
	Depth           int                           `bson:"depth,omitempty"`
	SortOrder       int                           `bson:"sort_order"`
	IsActive        bool                          `bson:"is_active"`
	AttributeSchema []attributeDefinitionDocument `bson:"attribute_schema,omitempty"`
}

type attributeDefinitionDocument struct {
	Key           string   `bson:"key"`
	Label         string   `bson:"label"`
	Type          string   `bson:"type"`
	Scope         string   `bson:"scope,omitempty"`
	Required      bool     `bson:"required"`
	Filterable    bool     `bson:"filterable"`
	Searchable    bool     `bson:"searchable,omitempty"`
	VariantAxis   bool     `bson:"variant_axis,omitempty"`
	Values        []string `bson:"values,omitempty"`
	AllowedValues []string `bson:"allowed_values,omitempty"`
	Unit          string   `bson:"unit,omitempty"`
}

func (d categoryDocument) toDomain() domain.Category {
	schema := make([]domain.AttributeDefinition, 0, len(d.AttributeSchema))
	for _, definition := range d.AttributeSchema {
		schema = append(schema, definition.toDomain())
	}
	level := d.Level
	if level == 0 {
		level = d.Depth
	}
	return domain.Category{
		ID:              d.ID,
		Name:            d.Name,
		Slug:            d.Slug,
		ParentID:        d.ParentID,
		Path:            d.Path,
		Level:           level,
		SortOrder:       d.SortOrder,
		IsActive:        d.IsActive,
		AttributeSchema: schema,
	}
}

func (d attributeDefinitionDocument) toDomain() domain.AttributeDefinition {
	scope := domain.AttributeScope(d.Scope)
	if !scope.Valid() {
		scope = domain.AttributeScopeProduct
		if d.VariantAxis {
			scope = domain.AttributeScopeVariant
		}
	}
	values := d.Values
	if len(values) == 0 {
		values = d.AllowedValues
	}
	return domain.AttributeDefinition{
		Key:        d.Key,
		Label:      d.Label,
		Type:       domain.AttributeType(d.Type),
		Scope:      scope,
		Required:   d.Required,
		Filterable: d.Filterable,
		Searchable: d.Searchable,
		Values:     values,
		Unit:       d.Unit,
	}
}

func normalizeRepositorySKU(sku string) string {
	return strings.ToUpper(strings.TrimSpace(sku))
}
