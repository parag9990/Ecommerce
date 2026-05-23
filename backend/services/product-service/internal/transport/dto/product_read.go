package dto

import (
	"strings"
	"time"

	"product-service/internal/domain"
	"product-service/internal/usecase"
)

type ListProductsRequestDTO struct {
	CategoryID string `json:"category_id,omitempty"`
	SellerID   string `json:"seller_id,omitempty"`
	Status     string `json:"status,omitempty"`
	Page       int    `json:"page,omitempty"`
	PageSize   int    `json:"page_size,omitempty"`
	Sort       string `json:"sort,omitempty"`
}

type GetProductRequestDTO struct {
	ProductID string `json:"product_id"`
}

type BatchGetProductsRequestDTO struct {
	ProductIDs []string `json:"product_ids"`
}

type ListCategoriesRequestDTO struct {
	ParentID string `json:"parent_id,omitempty"`
}

type ListSellerProductsRequestDTO struct {
	Actor      ActorContextDTO `json:"actor"`
	CategoryID string          `json:"category_id,omitempty"`
	Status     string          `json:"status,omitempty"`
	Page       int             `json:"page,omitempty"`
	PageSize   int             `json:"page_size,omitempty"`
	Sort       string          `json:"sort,omitempty"`
}

type GetSellerProductRequestDTO struct {
	Actor     ActorContextDTO `json:"actor"`
	ProductID string          `json:"product_id"`
}

type ProductListResponseDTO struct {
	Products []ProductReadDTO `json:"products"`
	Total    int64            `json:"total"`
}

type CategoryListResponseDTO struct {
	Categories []CategoryBrowseDTO `json:"categories"`
}

type ProductReadDTO struct {
	ID            string                  `json:"product_id"`
	SellerID      string                  `json:"seller_id"`
	Title         string                  `json:"title"`
	Slug          string                  `json:"slug,omitempty"`
	Description   string                  `json:"description,omitempty"`
	Brand         string                  `json:"brand,omitempty"`
	CategoryID    string                  `json:"category_id"`
	CategoryPath  []string                `json:"category_path,omitempty"`
	Status        string                  `json:"status"`
	Attributes    map[string]any          `json:"attributes,omitempty"`
	Images        []ProductImageReadDTO   `json:"images,omitempty"`
	Variants      []ProductVariantReadDTO `json:"variants"`
	RatingSummary RatingSummaryDTO        `json:"rating_summary"`
	UpdatedAt     time.Time               `json:"updated_at"`
	PublishedAt   *time.Time              `json:"published_at,omitempty"`
}

type ProductVariantReadDTO struct {
	ID                string         `json:"variant_id"`
	SKU               string         `json:"sku"`
	Title             string         `json:"title,omitempty"`
	Attributes        map[string]any `json:"attributes,omitempty"`
	Price             MoneyDTO       `json:"price"`
	MRP               *MoneyDTO      `json:"mrp,omitempty"`
	StockQuantity     int64          `json:"stock_quantity"`
	AvailableQuantity int64          `json:"available_quantity"`
	Status            string         `json:"status"`
}

type ProductImageReadDTO struct {
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

type CategoryBrowseDTO struct {
	ID        string   `json:"category_id"`
	Name      string   `json:"name"`
	Slug      string   `json:"slug,omitempty"`
	ParentID  *string  `json:"parent_id,omitempty"`
	Path      []string `json:"path,omitempty"`
	Level     int      `json:"level,omitempty"`
	SortOrder int      `json:"sort_order,omitempty"`
}

func (d ListProductsRequestDTO) ToUseCase() usecase.ListProductsRequest {
	return usecase.ListProductsRequest{
		CategoryID: strings.TrimSpace(d.CategoryID),
		SellerID:   strings.TrimSpace(d.SellerID),
		Status:     strings.TrimSpace(d.Status),
		Page:       d.Page,
		PageSize:   d.PageSize,
		Sort:       strings.TrimSpace(d.Sort),
	}
}

func (d BatchGetProductsRequestDTO) ToUseCase() usecase.BatchGetProductsRequest {
	ids := make([]string, 0, len(d.ProductIDs))
	for _, id := range d.ProductIDs {
		ids = append(ids, strings.TrimSpace(id))
	}
	return usecase.BatchGetProductsRequest{ProductIDs: ids}
}

func (d ListCategoriesRequestDTO) ToUseCase() usecase.ListCategoriesRequest {
	return usecase.ListCategoriesRequest{ParentID: strings.TrimSpace(d.ParentID)}
}

func (d ListSellerProductsRequestDTO) ToUseCase() usecase.SellerListProductsRequest {
	return usecase.SellerListProductsRequest{
		Actor:      d.Actor.ToDomain(),
		CategoryID: strings.TrimSpace(d.CategoryID),
		Status:     strings.TrimSpace(d.Status),
		Page:       d.Page,
		PageSize:   d.PageSize,
		Sort:       strings.TrimSpace(d.Sort),
	}
}

func (d GetSellerProductRequestDTO) ToUseCase() usecase.SellerGetProductRequest {
	return usecase.SellerGetProductRequest{
		Actor:     d.Actor.ToDomain(),
		ProductID: strings.TrimSpace(d.ProductID),
	}
}

func PublicProductListFromUseCase(result usecase.ProductListResult) ProductListResponseDTO {
	return ProductListResponseDTO{
		Products: PublicProductsFromDomain(result.Products),
		Total:    result.Total,
	}
}

func SellerProductListFromUseCase(result usecase.ProductListResult) ProductListResponseDTO {
	return ProductListResponseDTO{
		Products: SellerProductsFromDomain(result.Products),
		Total:    result.Total,
	}
}

func PublicProductsFromDomain(products []domain.Product) []ProductReadDTO {
	dtos := make([]ProductReadDTO, 0, len(products))
	for _, product := range products {
		dtos = append(dtos, PublicProductFromDomain(product))
	}
	return dtos
}

func SellerProductsFromDomain(products []domain.Product) []ProductReadDTO {
	dtos := make([]ProductReadDTO, 0, len(products))
	for _, product := range products {
		dtos = append(dtos, SellerProductFromDomain(product))
	}
	return dtos
}

func PublicProductFromDomain(product domain.Product) ProductReadDTO {
	return productReadFromDomain(product, false)
}

func SellerProductFromDomain(product domain.Product) ProductReadDTO {
	return productReadFromDomain(product, true)
}

func CategoriesFromDomain(categories []domain.Category) []CategoryBrowseDTO {
	dtos := make([]CategoryBrowseDTO, 0, len(categories))
	for _, category := range categories {
		dtos = append(dtos, CategoryBrowseFromDomain(category))
	}
	return dtos
}

func CategoryBrowseFromDomain(category domain.Category) CategoryBrowseDTO {
	return CategoryBrowseDTO{
		ID:        category.ID,
		Name:      category.Name,
		Slug:      category.Slug,
		ParentID:  category.ParentID,
		Path:      category.Path,
		Level:     category.Level,
		SortOrder: category.SortOrder,
	}
}

func productReadFromDomain(product domain.Product, includeInactive bool) ProductReadDTO {
	return ProductReadDTO{
		ID:            product.ID,
		SellerID:      product.SellerID,
		Title:         product.Title,
		Slug:          product.Slug,
		Description:   product.Description,
		Brand:         product.Brand,
		CategoryID:    product.CategoryID,
		CategoryPath:  product.CategoryPath,
		Status:        string(product.Status),
		Attributes:    map[string]any(product.Attributes),
		Images:        productReadImagesFromDomain(product.Images, includeInactive),
		Variants:      productReadVariantsFromDomain(product.Variants, includeInactive),
		RatingSummary: RatingSummaryDTO{Average: product.RatingSummary.Average, Count: product.RatingSummary.Count},
		UpdatedAt:     product.UpdatedAt,
		PublishedAt:   product.PublishedAt,
	}
}

func productReadVariantsFromDomain(variants []domain.Variant, includeInactive bool) []ProductVariantReadDTO {
	dtos := make([]ProductVariantReadDTO, 0, len(variants))
	for _, variant := range variants {
		if !includeInactive && !variant.Active() {
			continue
		}
		var mrp *MoneyDTO
		if variant.MRP != nil {
			value := MoneyFromDomain(*variant.MRP)
			mrp = &value
		}
		dtos = append(dtos, ProductVariantReadDTO{
			ID:                variant.ID,
			SKU:               variant.SKU,
			Title:             variant.Title,
			Attributes:        map[string]any(variant.Attributes),
			Price:             MoneyFromDomain(variant.Price),
			MRP:               mrp,
			StockQuantity:     variant.StockQuantity,
			AvailableQuantity: variant.AvailableQuantity(),
			Status:            string(variant.Status),
		})
	}
	return dtos
}

func productReadImagesFromDomain(images []domain.ProductImage, includeInactive bool) []ProductImageReadDTO {
	dtos := make([]ProductImageReadDTO, 0, len(images))
	for _, image := range images {
		if !includeInactive && !image.Active() {
			continue
		}
		dtos = append(dtos, ProductImageReadDTO{
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
	return dtos
}
