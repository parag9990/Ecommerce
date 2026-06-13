package domain

import (
	"fmt"
	"strings"
	"time"
)

type ProductIndexPayload struct {
	ProductID       string     `json:"product_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Brand           string     `json:"brand"`
	CategoryIDs     []string   `json:"category_ids"`
	SellerID        string     `json:"seller_id"`
	Price           *float64   `json:"price"`
	Rating          *float64   `json:"rating"`
	PopularityScore *int32     `json:"popularity_score"`
	InStock         *bool      `json:"in_stock"`
	Status          string     `json:"status"`
	IsDeleted       bool       `json:"is_deleted"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (p ProductIndexPayload) Normalized() ProductIndexPayload {
	p.ProductID = strings.TrimSpace(p.ProductID)
	p.Title = strings.TrimSpace(p.Title)
	p.Description = strings.TrimSpace(p.Description)
	p.Brand = strings.TrimSpace(p.Brand)
	p.SellerID = strings.TrimSpace(p.SellerID)
	p.Status = strings.ToLower(strings.TrimSpace(p.Status))
	p.CategoryIDs = cleanStringList(p.CategoryIDs)
	return p
}

func (p ProductIndexPayload) ValidateForRouting() error {
	p = p.Normalized()
	if strings.TrimSpace(p.ProductID) == "" {
		return fmt.Errorf("%w: product_id is required", ErrInvalidProductEvent)
	}
	if p.UpdatedAt.IsZero() {
		return fmt.Errorf("%w: updated_at is required", ErrInvalidProductEvent)
	}
	return nil
}

func (p ProductIndexPayload) ValidateForUpsert() error {
	p = p.Normalized()
	if err := p.ValidateForRouting(); err != nil {
		return err
	}
	if strings.TrimSpace(p.Title) == "" {
		return fmt.Errorf("%w: title is required for searchable products", ErrInvalidProductEvent)
	}
	if len(cleanStringList(p.CategoryIDs)) == 0 {
		return fmt.Errorf("%w: category_ids are required for searchable products", ErrInvalidProductEvent)
	}
	if strings.TrimSpace(p.SellerID) == "" {
		return fmt.Errorf("%w: seller_id is required for searchable products", ErrInvalidProductEvent)
	}
	if p.Price == nil {
		return fmt.Errorf("%w: price is required for searchable products", ErrInvalidProductEvent)
	}
	if *p.Price < 0 {
		return fmt.Errorf("%w: price cannot be negative", ErrInvalidProductEvent)
	}
	if p.Rating != nil && (*p.Rating < 0 || *p.Rating > 5) {
		return fmt.Errorf("%w: rating must be between 0 and 5", ErrInvalidProductEvent)
	}
	if p.PopularityScore == nil {
		return fmt.Errorf("%w: popularity_score is required for searchable products", ErrInvalidProductEvent)
	}
	if *p.PopularityScore < 0 {
		return fmt.Errorf("%w: popularity_score cannot be negative", ErrInvalidProductEvent)
	}
	if p.InStock == nil {
		return fmt.Errorf("%w: in_stock is required for searchable products", ErrInvalidProductEvent)
	}
	return nil
}

func (p ProductIndexPayload) IsSearchable() bool {
	p = p.Normalized()
	if p.ProductID == "" {
		return false
	}
	if p.IsDeleted {
		return false
	}
	if p.Status != ProductStatusPublished {
		return false
	}
	if p.Title == "" || p.SellerID == "" || len(p.CategoryIDs) == 0 {
		return false
	}
	return true
}

func cleanStringList(values []string) []string {
	cleaned := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		cleaned = append(cleaned, item)
	}
	return cleaned
}
