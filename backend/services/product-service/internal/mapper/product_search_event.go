package mapper

import (
	"math"
	"strings"

	"product-service/internal/domain"
)

func ToProductSearchEventPayload(product domain.Product) domain.ProductSearchEventPayload {
	price := primarySearchPrice(product)
	return domain.ProductSearchEventPayload{
		ProductID:       strings.TrimSpace(product.ID),
		SellerID:        strings.TrimSpace(product.SellerID),
		Title:           strings.TrimSpace(product.Title),
		Description:     strings.TrimSpace(product.Description),
		Brand:           strings.TrimSpace(product.Brand),
		CategoryID:      strings.TrimSpace(product.CategoryID),
		CategoryPath:    append([]string(nil), product.CategoryPath...),
		Status:          product.Status,
		SearchAction:    searchAction(product),
		Price:           price,
		Rating:          product.RatingSummary.Average,
		PopularityScore: popularityScore(product.RatingSummary),
		InStock:         productInStock(product),
		ImageURL:        primaryImageURL(product),
		Attributes:      searchableAttributes(product),
		UpdatedAt:       product.UpdatedAt.UTC(),
	}
}

func searchAction(product domain.Product) domain.SearchAction {
	if product.Status == domain.ProductStatusPublished {
		return domain.SearchActionUpsert
	}
	return domain.SearchActionDelete
}

func productInStock(product domain.Product) bool {
	for _, variant := range product.Variants {
		if variant.Active() && variant.AvailableQuantity() > 0 {
			return true
		}
	}
	return false
}

func primarySearchPrice(product domain.Product) domain.ProductEventMoney {
	var selected *domain.Money
	for index := range product.Variants {
		variant := product.Variants[index]
		if !variant.Active() {
			continue
		}
		if selected == nil || variant.Price.Amount < selected.Amount {
			price := variant.Price
			selected = &price
		}
	}
	if selected == nil && len(product.Variants) > 0 {
		price := product.Variants[0].Price
		selected = &price
	}
	if selected == nil {
		return domain.ProductEventMoney{}
	}
	return domain.ProductEventMoney{
		Amount:   selected.Amount,
		Currency: strings.ToUpper(selected.Currency),
	}
}

func primaryImageURL(product domain.Product) string {
	if image := product.PrimaryImage(); image != nil {
		return strings.TrimSpace(image.URL)
	}
	for _, image := range product.Images {
		if image.Active() {
			return strings.TrimSpace(image.URL)
		}
	}
	return ""
}

func searchableAttributes(product domain.Product) map[string]any {
	attributes := make(map[string]any, len(product.Attributes))
	for key, value := range product.Attributes {
		attributes[strings.TrimSpace(key)] = value
	}
	for _, variant := range product.Variants {
		if !variant.Active() {
			continue
		}
		for key, value := range variant.Attributes {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			if _, exists := attributes[key]; !exists {
				attributes[key] = value
			}
		}
	}
	return attributes
}

func popularityScore(summary domain.RatingSummary) int32 {
	if summary.Count <= 0 {
		return 0
	}
	score := summary.Average * math.Log10(float64(summary.Count)+1) * 20
	maxInt32 := float64(1<<31 - 1)
	if score > maxInt32 {
		return int32(maxInt32)
	}
	return int32(math.Round(score))
}
