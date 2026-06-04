package domain

import (
	"strings"
	"time"
)

const (
	ProductEventDefaultTopic  = "product.events"
	ProductEventDefaultSource = "product-service"
	ProductEventSchemaVersion = 1
	OutboxStatusPending       = ProductOutboxStatus("pending")
	OutboxStatusPublishing    = ProductOutboxStatus("publishing")
	OutboxStatusPublished     = ProductOutboxStatus("published")
	OutboxStatusDeadLettered  = ProductOutboxStatus("dead_lettered")
)

type ProductEventType string

const (
	ProductEventCreated          ProductEventType = "ProductCreated"
	ProductEventUpdated          ProductEventType = "ProductUpdated"
	ProductEventPublished        ProductEventType = "ProductPublished"
	ProductEventUnpublished      ProductEventType = "ProductUnpublished"
	ProductEventInventoryChanged ProductEventType = "ProductInventoryChanged"
)

type SearchAction string

const (
	SearchActionUpsert SearchAction = "UPSERT"
	SearchActionDelete SearchAction = "DELETE"
)

type ProductOutboxStatus string

type ProductEventMoney struct {
	Amount   int64  `json:"amount" bson:"amount"`
	Currency string `json:"currency" bson:"currency"`
}

type ProductSearchEventPayload struct {
	ProductID       string            `json:"product_id" bson:"product_id"`
	SellerID        string            `json:"seller_id" bson:"seller_id"`
	Title           string            `json:"title" bson:"title"`
	Description     string            `json:"description" bson:"description"`
	Brand           string            `json:"brand" bson:"brand"`
	CategoryID      string            `json:"category_id" bson:"category_id"`
	CategoryPath    []string          `json:"category_path" bson:"category_path"`
	Status          ProductStatus     `json:"status" bson:"status"`
	SearchAction    SearchAction      `json:"search_action" bson:"search_action"`
	Price           ProductEventMoney `json:"price" bson:"price"`
	Rating          float64           `json:"rating" bson:"rating"`
	PopularityScore int32             `json:"popularity_score" bson:"popularity_score"`
	InStock         bool              `json:"in_stock" bson:"in_stock"`
	ImageURL        string            `json:"image_url" bson:"image_url"`
	Attributes      map[string]any    `json:"attributes" bson:"attributes"`
	UpdatedAt       time.Time         `json:"updated_at" bson:"updated_at"`
}

type EventEnvelope struct {
	EventID    string         `json:"event_id" bson:"event_id"`
	EventType  string         `json:"event_type" bson:"event_type"`
	Version    int            `json:"version" bson:"version"`
	Source     string         `json:"source" bson:"source"`
	RequestID  string         `json:"request_id" bson:"request_id"`
	TraceID    string         `json:"trace_id" bson:"trace_id"`
	OccurredAt time.Time      `json:"occurred_at" bson:"occurred_at"`
	Payload    map[string]any `json:"payload" bson:"payload"`
}

type ProductOutboxEvent struct {
	ID            string
	Topic         string
	EventType     string
	Version       int
	Source        string
	RequestID     string
	TraceID       string
	Payload       map[string]any
	Status        ProductOutboxStatus
	Attempts      int
	NextAttemptAt time.Time
	OccurredAt    time.Time
	PublishedAt   *time.Time
	LastError     string
}

func (t ProductEventType) Valid() bool {
	switch t {
	case ProductEventCreated,
		ProductEventUpdated,
		ProductEventPublished,
		ProductEventUnpublished,
		ProductEventInventoryChanged:
		return true
	default:
		return false
	}
}

func (a SearchAction) Valid() bool {
	switch a {
	case SearchActionUpsert, SearchActionDelete:
		return true
	default:
		return false
	}
}

func (s ProductOutboxStatus) Valid() bool {
	switch s {
	case OutboxStatusPending, OutboxStatusPublishing, OutboxStatusPublished, OutboxStatusDeadLettered:
		return true
	default:
		return false
	}
}

func (p ProductSearchEventPayload) Validate() ValidationReport {
	var report ValidationReport
	if !requiredString(p.ProductID) {
		report.AddError(CodeProductIDRequired, "payload.product_id", "product id is required")
	}
	if !requiredString(p.SellerID) {
		report.AddError(CodeSellerRequired, "payload.seller_id", "seller id is required")
	}
	if !requiredString(p.Title) && p.SearchAction == SearchActionUpsert {
		report.AddError(CodeTitleRequired, "payload.title", "title is required for search upsert events")
	}
	if !requiredString(p.CategoryID) && p.SearchAction == SearchActionUpsert {
		report.AddError(CodeCategoryRequired, "payload.category_id", "category id is required for search upsert events")
	}
	if !p.Status.Valid() {
		report.AddError(CodeInvalidProductStatus, "payload.status", "product status is not supported")
	}
	if !p.SearchAction.Valid() {
		report.AddError(CodeInvalidSearchAction, "payload.search_action", "search action must be UPSERT or DELETE")
	}
	if p.SearchAction == SearchActionUpsert {
		report.Merge(NewMoney(p.Price.Amount, p.Price.Currency).Validate("payload.price"))
	}
	if p.Rating < 0 || p.Rating > 5 {
		report.AddError(CodeInvalidRating, "payload.rating", "rating must be between 0 and 5")
	}
	if p.UpdatedAt.IsZero() {
		report.AddError(CodeInvalidEventPayload, "payload.updated_at", "payload updated time is required")
	}
	return report
}

func (p ProductSearchEventPayload) ToMap() map[string]any {
	attributes := make(map[string]any, len(p.Attributes))
	for key, value := range p.Attributes {
		attributes[key] = value
	}
	categoryPath := append([]string(nil), p.CategoryPath...)
	return map[string]any{
		"product_id":    p.ProductID,
		"seller_id":     p.SellerID,
		"title":         p.Title,
		"description":   p.Description,
		"brand":         p.Brand,
		"category_id":   p.CategoryID,
		"category_path": categoryPath,
		"status":        string(p.Status),
		"search_action": string(p.SearchAction),
		"price": map[string]any{
			"amount":   p.Price.Amount,
			"currency": strings.ToUpper(p.Price.Currency),
		},
		"rating":           p.Rating,
		"popularity_score": p.PopularityScore,
		"in_stock":         p.InStock,
		"image_url":        p.ImageURL,
		"attributes":       attributes,
		"updated_at":       p.UpdatedAt,
	}
}

func (e ProductOutboxEvent) Envelope() EventEnvelope {
	payload := make(map[string]any, len(e.Payload))
	for key, value := range e.Payload {
		payload[key] = value
	}
	return EventEnvelope{
		EventID:    e.ID,
		EventType:  e.EventType,
		Version:    e.Version,
		Source:     e.Source,
		RequestID:  e.RequestID,
		TraceID:    e.TraceID,
		OccurredAt: e.OccurredAt,
		Payload:    payload,
	}
}

func (e ProductOutboxEvent) Validate() ValidationReport {
	var report ValidationReport
	if !requiredString(e.ID) {
		report.AddError(CodeEventIDRequired, "event_id", "event id is required")
	}
	if strings.TrimSpace(e.Topic) == "" {
		report.AddError(CodeInvalidEventEnvelope, "topic", "event topic is required")
	}
	if !ProductEventType(e.EventType).Valid() {
		report.AddError(CodeInvalidEventType, "event_type", "product event type is not supported")
	}
	if e.Version <= 0 {
		report.AddError(CodeInvalidEventEnvelope, "version", "event version must be greater than zero")
	}
	if strings.TrimSpace(e.Source) == "" {
		report.AddError(CodeInvalidEventEnvelope, "source", "event source is required")
	}
	if strings.TrimSpace(e.RequestID) == "" {
		report.AddError(CodeInvalidEventEnvelope, "request_id", "request id is required")
	}
	if strings.TrimSpace(e.TraceID) == "" {
		report.AddError(CodeInvalidEventEnvelope, "trace_id", "trace id is required")
	}
	if len(e.Payload) == 0 {
		report.AddError(CodeInvalidEventPayload, "payload", "event payload is required")
	}
	if !e.Status.Valid() {
		report.AddError(CodeInvalidOutboxStatus, "status", "outbox status is not supported")
	}
	if e.Attempts < 0 {
		report.AddError(CodeInvalidOutboxStatus, "attempts", "outbox attempts cannot be negative")
	}
	if e.NextAttemptAt.IsZero() {
		report.AddError(CodeInvalidOutboxStatus, "next_attempt_at", "next attempt time is required")
	}
	if e.OccurredAt.IsZero() {
		report.AddError(CodeInvalidEventEnvelope, "occurred_at", "event occurrence time is required")
	}
	return report
}
