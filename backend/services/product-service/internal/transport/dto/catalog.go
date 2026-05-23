package dto

import (
	"time"

	"product-service/internal/domain"
	"product-service/internal/usecase"
)

type MoneyDTO struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type ProductDTO struct {
	ID            string            `json:"product_id"`
	SellerID      string            `json:"seller_id"`
	Title         string            `json:"title"`
	Slug          string            `json:"slug,omitempty"`
	Description   string            `json:"description,omitempty"`
	Brand         string            `json:"brand,omitempty"`
	CategoryID    string            `json:"category_id"`
	CategoryPath  []string          `json:"category_path,omitempty"`
	Status        string            `json:"status"`
	Attributes    map[string]any    `json:"attributes,omitempty"`
	Images        []ProductImageDTO `json:"images,omitempty"`
	Variants      []VariantDTO      `json:"variants"`
	RatingSummary RatingSummaryDTO  `json:"rating_summary"`
	CreatedBy     string            `json:"created_by,omitempty"`
	UpdatedBy     string            `json:"updated_by,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	PublishedAt   *time.Time        `json:"published_at,omitempty"`
}

type VariantDTO struct {
	ID               string         `json:"variant_id"`
	SKU              string         `json:"sku"`
	Title            string         `json:"title,omitempty"`
	Attributes       map[string]any `json:"attributes"`
	Price            MoneyDTO       `json:"price"`
	MRP              *MoneyDTO      `json:"mrp,omitempty"`
	StockQuantity    int64          `json:"stock_quantity"`
	ReservedQuantity int64          `json:"reserved_quantity"`
	SafetyStock      int64          `json:"safety_stock"`
	Status           string         `json:"status"`
	Barcode          string         `json:"barcode,omitempty"`
}

type ProductImageDTO struct {
	ID         string   `json:"image_id"`
	URL        string   `json:"url"`
	AltText    string   `json:"alt_text,omitempty"`
	Position   int      `json:"position"`
	IsPrimary  bool     `json:"is_primary"`
	VariantIDs []string `json:"variant_ids,omitempty"`
	Width      int      `json:"width,omitempty"`
	Height     int      `json:"height,omitempty"`
	Status     string   `json:"status"`
}

type RatingSummaryDTO struct {
	Average float64 `json:"average"`
	Count   int64   `json:"count"`
}

type CategoryDTO struct {
	ID              string                   `json:"category_id"`
	Name            string                   `json:"name"`
	Slug            string                   `json:"slug"`
	ParentID        *string                  `json:"parent_id,omitempty"`
	Path            []string                 `json:"path"`
	Level           int                      `json:"level"`
	SortOrder       int                      `json:"sort_order"`
	IsActive        bool                     `json:"is_active"`
	AttributeSchema []AttributeDefinitionDTO `json:"attribute_schema,omitempty"`
}

type AttributeDefinitionDTO struct {
	Key        string   `json:"key"`
	Label      string   `json:"label"`
	Type       string   `json:"type"`
	Scope      string   `json:"scope"`
	Required   bool     `json:"required"`
	Filterable bool     `json:"filterable"`
	Searchable bool     `json:"searchable"`
	Values     []string `json:"values,omitempty"`
	Unit       string   `json:"unit,omitempty"`
}

type ValidateProductRequestDTO struct {
	Product            ProductDTO   `json:"product"`
	Category           *CategoryDTO `json:"category,omitempty"`
	ExcludeProductID   string       `json:"exclude_product_id,omitempty"`
	CheckSKUUniqueness bool         `json:"check_sku_uniqueness"`
}

type ValidateProductResponseDTO struct {
	Report ValidationReportDTO `json:"report"`
}

type ValidationReportDTO struct {
	Valid  bool                 `json:"valid"`
	Issues []ValidationIssueDTO `json:"issues"`
}

type ValidationIssueDTO struct {
	Code     string `json:"code"`
	Field    string `json:"field,omitempty"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

func (d MoneyDTO) ToDomain() domain.Money {
	return domain.NewMoney(d.Amount, d.Currency)
}

func MoneyFromDomain(m domain.Money) MoneyDTO {
	return MoneyDTO{Amount: m.Amount, Currency: m.Currency}
}

func (d ProductDTO) ToDomain() domain.Product {
	images := make([]domain.ProductImage, 0, len(d.Images))
	for _, image := range d.Images {
		images = append(images, image.ToDomain())
	}
	variants := make([]domain.Variant, 0, len(d.Variants))
	for _, variant := range d.Variants {
		variants = append(variants, variant.ToDomain())
	}
	return domain.Product{
		ID:            d.ID,
		SellerID:      d.SellerID,
		Title:         d.Title,
		Slug:          d.Slug,
		Description:   d.Description,
		Brand:         d.Brand,
		CategoryID:    d.CategoryID,
		CategoryPath:  d.CategoryPath,
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

func ProductFromDomain(p domain.Product) ProductDTO {
	images := make([]ProductImageDTO, 0, len(p.Images))
	for _, image := range p.Images {
		images = append(images, ProductImageFromDomain(image))
	}
	variants := make([]VariantDTO, 0, len(p.Variants))
	for _, variant := range p.Variants {
		variants = append(variants, VariantFromDomain(variant))
	}
	return ProductDTO{
		ID:            p.ID,
		SellerID:      p.SellerID,
		Title:         p.Title,
		Slug:          p.Slug,
		Description:   p.Description,
		Brand:         p.Brand,
		CategoryID:    p.CategoryID,
		CategoryPath:  p.CategoryPath,
		Status:        string(p.Status),
		Attributes:    map[string]any(p.Attributes),
		Images:        images,
		Variants:      variants,
		RatingSummary: RatingSummaryDTO{Average: p.RatingSummary.Average, Count: p.RatingSummary.Count},
		CreatedBy:     p.CreatedBy,
		UpdatedBy:     p.UpdatedBy,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		PublishedAt:   p.PublishedAt,
	}
}

func (d VariantDTO) ToDomain() domain.Variant {
	var mrp *domain.Money
	if d.MRP != nil {
		value := d.MRP.ToDomain()
		mrp = &value
	}
	return domain.Variant{
		ID:               d.ID,
		SKU:              d.SKU,
		Title:            d.Title,
		Attributes:       domain.Attributes(d.Attributes),
		Price:            d.Price.ToDomain(),
		MRP:              mrp,
		StockQuantity:    d.StockQuantity,
		ReservedQuantity: d.ReservedQuantity,
		SafetyStock:      d.SafetyStock,
		Status:           domain.VariantStatus(d.Status),
		Barcode:          d.Barcode,
	}
}

func VariantFromDomain(v domain.Variant) VariantDTO {
	var mrp *MoneyDTO
	if v.MRP != nil {
		value := MoneyFromDomain(*v.MRP)
		mrp = &value
	}
	return VariantDTO{
		ID:               v.ID,
		SKU:              v.SKU,
		Title:            v.Title,
		Attributes:       map[string]any(v.Attributes),
		Price:            MoneyFromDomain(v.Price),
		MRP:              mrp,
		StockQuantity:    v.StockQuantity,
		ReservedQuantity: v.ReservedQuantity,
		SafetyStock:      v.SafetyStock,
		Status:           string(v.Status),
		Barcode:          v.Barcode,
	}
}

func (d ProductImageDTO) ToDomain() domain.ProductImage {
	return domain.ProductImage{
		ID:         d.ID,
		URL:        d.URL,
		AltText:    d.AltText,
		Position:   d.Position,
		IsPrimary:  d.IsPrimary,
		VariantIDs: d.VariantIDs,
		Width:      d.Width,
		Height:     d.Height,
		Status:     domain.ImageStatus(d.Status),
	}
}

func ProductImageFromDomain(i domain.ProductImage) ProductImageDTO {
	return ProductImageDTO{
		ID:         i.ID,
		URL:        i.URL,
		AltText:    i.AltText,
		Position:   i.Position,
		IsPrimary:  i.IsPrimary,
		VariantIDs: i.VariantIDs,
		Width:      i.Width,
		Height:     i.Height,
		Status:     string(i.Status),
	}
}

func (d CategoryDTO) ToDomain() domain.Category {
	schema := make([]domain.AttributeDefinition, 0, len(d.AttributeSchema))
	for _, definition := range d.AttributeSchema {
		schema = append(schema, definition.ToDomain())
	}
	return domain.Category{
		ID:              d.ID,
		Name:            d.Name,
		Slug:            d.Slug,
		ParentID:        d.ParentID,
		Path:            d.Path,
		Level:           d.Level,
		SortOrder:       d.SortOrder,
		IsActive:        d.IsActive,
		AttributeSchema: schema,
	}
}

func CategoryFromDomain(c domain.Category) CategoryDTO {
	schema := make([]AttributeDefinitionDTO, 0, len(c.AttributeSchema))
	for _, definition := range c.AttributeSchema {
		schema = append(schema, AttributeDefinitionFromDomain(definition))
	}
	return CategoryDTO{
		ID:              c.ID,
		Name:            c.Name,
		Slug:            c.Slug,
		ParentID:        c.ParentID,
		Path:            c.Path,
		Level:           c.Level,
		SortOrder:       c.SortOrder,
		IsActive:        c.IsActive,
		AttributeSchema: schema,
	}
}

func (d AttributeDefinitionDTO) ToDomain() domain.AttributeDefinition {
	return domain.AttributeDefinition{
		Key:        d.Key,
		Label:      d.Label,
		Type:       domain.AttributeType(d.Type),
		Scope:      domain.AttributeScope(d.Scope),
		Required:   d.Required,
		Filterable: d.Filterable,
		Searchable: d.Searchable,
		Values:     d.Values,
		Unit:       d.Unit,
	}
}

func AttributeDefinitionFromDomain(a domain.AttributeDefinition) AttributeDefinitionDTO {
	return AttributeDefinitionDTO{
		Key:        a.Key,
		Label:      a.Label,
		Type:       string(a.Type),
		Scope:      string(a.Scope),
		Required:   a.Required,
		Filterable: a.Filterable,
		Searchable: a.Searchable,
		Values:     a.Values,
		Unit:       a.Unit,
	}
}

func (d ValidateProductRequestDTO) ToUseCase() usecase.ValidateProductRequest {
	var category *domain.Category
	if d.Category != nil {
		value := d.Category.ToDomain()
		category = &value
	}
	return usecase.ValidateProductRequest{
		Product:            d.Product.ToDomain(),
		Category:           category,
		ExcludeProductID:   d.ExcludeProductID,
		CheckSKUUniqueness: d.CheckSKUUniqueness,
	}
}

func ValidationReportFromDomain(report domain.ValidationReport) ValidationReportDTO {
	issues := make([]ValidationIssueDTO, 0, len(report.Issues))
	for _, issue := range report.Issues {
		issues = append(issues, ValidationIssueDTO{
			Code:     issue.Code,
			Field:    issue.Field,
			Message:  issue.Message,
			Severity: string(issue.Severity),
		})
	}
	return ValidationReportDTO{
		Valid:  report.IsValid(),
		Issues: issues,
	}
}
