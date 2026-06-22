package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	WishlistEventVersion = 1
	WishlistEventSource  = "wishlist"
)

type WishlistEventType string

const (
	WishlistEventItemAdded   WishlistEventType = "wishlist_item_added"
	WishlistEventItemRemoved WishlistEventType = "wishlist_item_removed"
)

func (t WishlistEventType) IsValid() bool {
	switch t {
	case WishlistEventItemAdded, WishlistEventItemRemoved:
		return true
	default:
		return false
	}
}

type WishlistEventStatus string

const (
	WishlistEventPending    WishlistEventStatus = "pending"
	WishlistEventPublishing WishlistEventStatus = "publishing"
	WishlistEventPublished  WishlistEventStatus = "published"
	WishlistEventFailed     WishlistEventStatus = "failed"
)

func (s WishlistEventStatus) IsValid() bool {
	switch s {
	case WishlistEventPending, WishlistEventPublishing, WishlistEventPublished, WishlistEventFailed:
		return true
	default:
		return false
	}
}

type WishlistEventAction string

const (
	WishlistEventActionAdd    WishlistEventAction = "add"
	WishlistEventActionRemove WishlistEventAction = "remove"
)

func (a WishlistEventAction) IsValid() bool {
	switch a {
	case WishlistEventActionAdd, WishlistEventActionRemove:
		return true
	default:
		return false
	}
}

type WishlistAnalyticsPayload struct {
	UserID         string
	ProductID      string
	VariantID      string
	Action         WishlistEventAction
	Source         string
	Availability   Availability
	LastKnownPrice *Money
}

func (p WishlistAnalyticsPayload) Normalized() WishlistAnalyticsPayload {
	p.UserID = normalizeID(p.UserID)
	p.ProductID = normalizeID(p.ProductID)
	p.VariantID = normalizeID(p.VariantID)
	p.Source = strings.TrimSpace(p.Source)
	p.Availability = p.Availability.Normalized()
	p.LastKnownPrice = copyMoney(p.LastKnownPrice)
	return p
}

func (p WishlistAnalyticsPayload) Validate() error {
	p = p.Normalized()
	if p.UserID == "" {
		return invalidField("payload.user_id", "is required")
	}
	if p.ProductID == "" {
		return invalidField("payload.product_id", "is required")
	}
	if !p.Action.IsValid() {
		return invalidField("payload.action", "must be add or remove")
	}
	if p.Source != WishlistEventSource {
		return invalidField("payload.source", "must be wishlist")
	}
	if !p.Availability.IsValid() {
		return invalidField("payload.availability", "must be in_stock, out_of_stock, unknown, or deleted")
	}
	if p.LastKnownPrice != nil {
		if err := p.LastKnownPrice.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type WishlistAnalyticsEvent struct {
	EventID     string
	EventType   WishlistEventType
	Version     int
	Topic       string
	Payload     WishlistAnalyticsPayload
	TraceID     string
	Status      WishlistEventStatus
	Attempts    int
	NextRetryAt time.Time
	LastError   string
	OccurredAt  time.Time
	PublishedAt *time.Time
	LockedBy    string
	LockedUntil *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (e WishlistAnalyticsEvent) Normalized() WishlistAnalyticsEvent {
	e.EventID = normalizeID(e.EventID)
	e.Topic = strings.TrimSpace(e.Topic)
	e.Payload = e.Payload.Normalized()
	e.TraceID = normalizeID(e.TraceID)
	e.LastError = strings.TrimSpace(e.LastError)
	e.LockedBy = normalizeID(e.LockedBy)
	e.NextRetryAt = normalizeTime(e.NextRetryAt)
	e.OccurredAt = normalizeTime(e.OccurredAt)
	if e.PublishedAt != nil {
		publishedAt := normalizeTime(*e.PublishedAt)
		e.PublishedAt = &publishedAt
	}
	if e.LockedUntil != nil {
		lockedUntil := normalizeTime(*e.LockedUntil)
		e.LockedUntil = &lockedUntil
	}
	e.CreatedAt = normalizeTime(e.CreatedAt)
	e.UpdatedAt = normalizeTime(e.UpdatedAt)
	return e
}

func (e WishlistAnalyticsEvent) Validate() error {
	e = e.Normalized()
	if e.EventID == "" {
		return invalidField("event_id", "is required")
	}
	if !e.EventType.IsValid() {
		return invalidField("event_type", "must be wishlist_item_added or wishlist_item_removed")
	}
	if e.Version <= 0 {
		return invalidField("version", "must be positive")
	}
	if e.Topic == "" {
		return invalidField("topic", "is required")
	}
	if err := e.Payload.Validate(); err != nil {
		return err
	}
	if !eventActionMatchesType(e.EventType, e.Payload.Action) {
		return invalidField("payload.action", fmt.Sprintf("does not match event_type %s", e.EventType))
	}
	if !e.Status.IsValid() {
		return invalidField("status", "must be pending, publishing, published, or failed")
	}
	if e.Attempts < 0 {
		return invalidField("attempts", "must be greater than or equal to zero")
	}
	if e.NextRetryAt.IsZero() {
		return invalidField("next_retry_at", "is required")
	}
	if e.OccurredAt.IsZero() {
		return invalidField("occurred_at", "is required")
	}
	if e.CreatedAt.IsZero() {
		return invalidField("created_at", "is required")
	}
	if e.UpdatedAt.IsZero() {
		return invalidField("updated_at", "is required")
	}
	if e.UpdatedAt.Before(e.CreatedAt) {
		return invalidField("updated_at", "must not be before created_at")
	}
	if e.Status == WishlistEventPublished && e.PublishedAt == nil {
		return invalidField("published_at", "is required when status is published")
	}
	if e.Status == WishlistEventPublishing && (e.LockedBy == "" || e.LockedUntil == nil || e.LockedUntil.IsZero()) {
		return invalidField("locked_until", "and locked_by are required when status is publishing")
	}
	return nil
}

func eventActionMatchesType(eventType WishlistEventType, action WishlistEventAction) bool {
	switch eventType {
	case WishlistEventItemAdded:
		return action == WishlistEventActionAdd
	case WishlistEventItemRemoved:
		return action == WishlistEventActionRemove
	default:
		return false
	}
}
