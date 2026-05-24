package domain

import (
	"fmt"
	"strings"
)

const (
	ProductFieldID              = "id"
	ProductFieldTitle           = "title"
	ProductFieldDescription     = "description"
	ProductFieldBrand           = "brand"
	ProductFieldCategoryIDs     = "category_ids"
	ProductFieldSellerID        = "seller_id"
	ProductFieldPrice           = "price"
	ProductFieldRating          = "rating"
	ProductFieldPopularityScore = "popularity_score"
	ProductFieldInStock         = "in_stock"
	ProductFieldCreatedAt       = "created_at"
)

type ProductDocument struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Description     *string  `json:"description,omitempty"`
	Brand           *string  `json:"brand,omitempty"`
	CategoryIDs     []string `json:"category_ids"`
	SellerID        string   `json:"seller_id"`
	Price           float64  `json:"price"`
	Rating          *float64 `json:"rating,omitempty"`
	PopularityScore int32    `json:"popularity_score"`
	InStock         bool     `json:"in_stock"`
	CreatedAt       int64    `json:"created_at"`
}

func (d ProductDocument) Validate() error {
	if strings.TrimSpace(d.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidProductDocument)
	}
	if strings.TrimSpace(d.Title) == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidProductDocument)
	}
	if len(d.CategoryIDs) == 0 {
		return fmt.Errorf("%w: at least one category id is required", ErrInvalidProductDocument)
	}
	for i, categoryID := range d.CategoryIDs {
		if strings.TrimSpace(categoryID) == "" {
			return fmt.Errorf("%w: category_ids[%d] is required", ErrInvalidProductDocument, i)
		}
	}
	if strings.TrimSpace(d.SellerID) == "" {
		return fmt.Errorf("%w: seller_id is required", ErrInvalidProductDocument)
	}
	if d.Price < 0 {
		return fmt.Errorf("%w: price cannot be negative", ErrInvalidProductDocument)
	}
	if d.Rating != nil && (*d.Rating < 0 || *d.Rating > 5) {
		return fmt.Errorf("%w: rating must be between 0 and 5", ErrInvalidProductDocument)
	}
	if d.PopularityScore < 0 {
		return fmt.Errorf("%w: popularity_score cannot be negative", ErrInvalidProductDocument)
	}
	if d.CreatedAt <= 0 {
		return fmt.Errorf("%w: created_at must be a unix timestamp in seconds", ErrInvalidProductDocument)
	}
	return nil
}
