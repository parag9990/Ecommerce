package dto

import (
	"strings"
	"time"

	"product-service/internal/domain"
	"product-service/internal/mapper"
	"product-service/internal/usecase"
)

type SearchProductExportRequestDTO struct {
	Cursor string
	Limit  int
}

type SearchProductExportResponseDTO struct {
	Items      []SearchProductExportItemDTO `json:"items"`
	NextCursor string                       `json:"next_cursor"`
	HasMore    bool                         `json:"has_more"`
	Total      int64                        `json:"total"`
}

type SearchProductExportItemDTO struct {
	ProductID       string     `json:"product_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Brand           string     `json:"brand"`
	CategoryIDs     []string   `json:"category_ids"`
	SellerID        string     `json:"seller_id"`
	Price           float64    `json:"price"`
	Rating          float64    `json:"rating"`
	PopularityScore int32      `json:"popularity_score"`
	InStock         bool       `json:"in_stock"`
	Status          string     `json:"status"`
	IsDeleted       bool       `json:"is_deleted"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (d SearchProductExportRequestDTO) ToUseCase() usecase.SearchProductExportRequest {
	return usecase.SearchProductExportRequest{Cursor: strings.TrimSpace(d.Cursor), Limit: d.Limit}
}

func SearchProductExportFromUseCase(result usecase.SearchProductExportResult) SearchProductExportResponseDTO {
	items := make([]SearchProductExportItemDTO, 0, len(result.Products))
	for _, product := range result.Products {
		items = append(items, searchProductExportItem(product))
	}
	return SearchProductExportResponseDTO{
		Items:      items,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
		Total:      result.Total,
	}
}

func searchProductExportItem(product domain.Product) SearchProductExportItemDTO {
	payload := mapper.ToProductSearchEventPayload(product)
	categoryIDs := append([]string(nil), payload.CategoryPath...)
	if payload.CategoryID != "" && !contains(categoryIDs, payload.CategoryID) {
		categoryIDs = append(categoryIDs, payload.CategoryID)
	}
	createdAt := product.CreatedAt.UTC()
	return SearchProductExportItemDTO{
		ProductID:       payload.ProductID,
		Title:           payload.Title,
		Description:     payload.Description,
		Brand:           payload.Brand,
		CategoryIDs:     categoryIDs,
		SellerID:        payload.SellerID,
		Price:           float64(payload.Price.Amount),
		Rating:          payload.Rating,
		PopularityScore: payload.PopularityScore,
		InStock:         payload.InStock,
		Status:          string(payload.Status),
		CreatedAt:       &createdAt,
		UpdatedAt:       payload.UpdatedAt.UTC(),
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
