package domain

const (
	ProductEventPublished        = "ProductPublished"
	ProductEventUpdated          = "ProductUpdated"
	ProductEventPriceChanged     = "ProductPriceChanged"
	ProductEventInventoryChanged = "ProductInventoryChanged"
	ProductEventUnpublished      = "ProductUnpublished"
	ProductEventDeleted          = "ProductDeleted"
	ProductEventBlocked          = "ProductBlocked"

	ProductStatusPublished = "published"
)

func IsSupportedProductEvent(eventType string) bool {
	switch eventType {
	case ProductEventPublished,
		ProductEventUpdated,
		ProductEventPriceChanged,
		ProductEventInventoryChanged,
		ProductEventUnpublished,
		ProductEventDeleted,
		ProductEventBlocked:
		return true
	default:
		return false
	}
}

func IsProductDeleteEvent(eventType string) bool {
	switch eventType {
	case ProductEventDeleted, ProductEventUnpublished, ProductEventBlocked:
		return true
	default:
		return false
	}
}
