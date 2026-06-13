package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	CurrentJourneySummarySchemaVersion = 1

	DefaultMaxJourneyMilestones = 20
	DefaultMaxJourneyTopPaths   = 20
)

type JourneyMilestoneName string

const (
	MilestoneFirstPageView    JourneyMilestoneName = "first_page_view"
	MilestoneFirstSearch      JourneyMilestoneName = "first_search"
	MilestoneFirstProductView JourneyMilestoneName = "first_product_view"
	MilestoneFirstAddToCart   JourneyMilestoneName = "first_add_to_cart"
	MilestoneCheckoutStarted  JourneyMilestoneName = "checkout_started"
	MilestonePaymentCompleted JourneyMilestoneName = "payment_completed"
)

type Journey struct {
	Session Session        `json:"session" bson:"session"`
	Events  []SessionEvent `json:"events" bson:"events"`
	Summary JourneySummary `json:"summary" bson:"summary"`
}

type JourneySummary struct {
	SessionID        string             `json:"session_id" bson:"session_id"`
	AnonymousID      string             `json:"anonymous_id" bson:"anonymous_id"`
	UserID           *string            `json:"user_id,omitempty" bson:"user_id,omitempty"`
	EntryPage        string             `json:"entry_page,omitempty" bson:"entry_page,omitempty"`
	ExitPage         string             `json:"exit_page,omitempty" bson:"exit_page,omitempty"`
	FirstEventAt     time.Time          `json:"first_event_at,omitempty" bson:"first_event_at,omitempty"`
	LastEventAt      time.Time          `json:"last_event_at,omitempty" bson:"last_event_at,omitempty"`
	DurationSeconds  int64              `json:"duration_seconds" bson:"duration_seconds"`
	TotalEvents      int                `json:"total_events" bson:"total_events"`
	ProductsViewed   int                `json:"products_viewed" bson:"products_viewed"`
	Searches         int                `json:"searches" bson:"searches"`
	CartActions      int                `json:"cart_actions" bson:"cart_actions"`
	CheckoutStarted  bool               `json:"checkout_started" bson:"checkout_started"`
	PaymentCompleted bool               `json:"payment_completed" bson:"payment_completed"`
	Milestones       []JourneyMilestone `json:"milestones" bson:"milestones"`
	TopPaths         []JourneyPathCount `json:"top_paths" bson:"top_paths"`
	SchemaVersion    int                `json:"schema_version" bson:"schema_version"`
	CalculatedAt     time.Time          `json:"calculated_at" bson:"calculated_at"`
	CreatedAt        time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
	RetainUntil      *time.Time         `json:"retain_until,omitempty" bson:"retain_until,omitempty"`
	AnonymizedAt     *time.Time         `json:"anonymized_at,omitempty" bson:"anonymized_at,omitempty"`
	LegalHold        bool               `json:"legal_hold,omitempty" bson:"legal_hold,omitempty"`
	LegalHoldReason  *string            `json:"legal_hold_reason,omitempty" bson:"legal_hold_reason,omitempty"`
	LegalHoldSetBy   *string            `json:"legal_hold_set_by,omitempty" bson:"legal_hold_set_by,omitempty"`
	LegalHoldSetAt   *time.Time         `json:"legal_hold_set_at,omitempty" bson:"legal_hold_set_at,omitempty"`
	RetentionClass   RetentionClass     `json:"retention_class,omitempty" bson:"retention_class,omitempty"`
}

type JourneyMilestone struct {
	Name       JourneyMilestoneName `json:"name" bson:"name"`
	EventID    string               `json:"event_id" bson:"event_id"`
	Path       string               `json:"path,omitempty" bson:"path,omitempty"`
	OccurredAt time.Time            `json:"occurred_at" bson:"occurred_at"`
}

type JourneyPathCount struct {
	Path  string `json:"path" bson:"path"`
	Count int    `json:"count" bson:"count"`
}

type JourneySummaryValidationConfig struct {
	Session       ValidationConfig
	MaxMilestones int
	MaxTopPaths   int
}

type JourneyValidationError struct {
	Violations []FieldViolation
}

func (e JourneyValidationError) Error() string {
	if len(e.Violations) == 0 {
		return ErrInvalidJourney.Error()
	}
	parts := make([]string, 0, len(e.Violations))
	for _, violation := range e.Violations {
		parts = append(parts, fmt.Sprintf("%s: %s", violation.Field, violation.Message))
	}
	return ErrInvalidJourney.Error() + ": " + strings.Join(parts, "; ")
}

func (e JourneyValidationError) Is(target error) bool {
	return target == ErrInvalidJourney
}

func DefaultJourneySummaryValidationConfig() JourneySummaryValidationConfig {
	return JourneySummaryValidationConfig{
		Session:       DefaultValidationConfig(),
		MaxMilestones: DefaultMaxJourneyMilestones,
		MaxTopPaths:   DefaultMaxJourneyTopPaths,
	}
}

func (c JourneySummaryValidationConfig) WithDefaults() JourneySummaryValidationConfig {
	defaults := DefaultJourneySummaryValidationConfig()
	c.Session = c.Session.WithDefaults()
	if c.MaxMilestones <= 0 {
		c.MaxMilestones = defaults.MaxMilestones
	}
	if c.MaxTopPaths <= 0 {
		c.MaxTopPaths = defaults.MaxTopPaths
	}
	return c
}

func BuildJourneySummary(session Session, events []SessionEvent, calculatedAt time.Time, topPathLimit int) JourneySummary {
	normalizedSession := session.Normalize()
	if topPathLimit <= 0 {
		topPathLimit = 10
	}
	calculatedAt = normalizeTime(calculatedAt)
	if calculatedAt.IsZero() {
		calculatedAt = time.Now().UTC()
	}

	summary := JourneySummary{
		SessionID:     normalizedSession.SessionID,
		AnonymousID:   normalizedSession.AnonymousID,
		UserID:        normalizedSession.UserID,
		TotalEvents:   len(events),
		SchemaVersion: CurrentJourneySummarySchemaVersion,
		CalculatedAt:  calculatedAt,
		UpdatedAt:     calculatedAt,
	}
	if len(events) == 0 {
		return summary.Normalize()
	}

	normalizedEvents := make([]SessionEvent, 0, len(events))
	for _, event := range events {
		normalizedEvents = append(normalizedEvents, event.Normalize())
	}
	sortSessionEvents(normalizedEvents)

	summary.FirstEventAt = normalizedEvents[0].OccurredAt
	summary.LastEventAt = normalizedEvents[len(normalizedEvents)-1].OccurredAt
	if !summary.FirstEventAt.IsZero() && !summary.LastEventAt.Before(summary.FirstEventAt) {
		summary.DurationSeconds = int64(summary.LastEventAt.Sub(summary.FirstEventAt).Seconds())
	}

	milestoneSeen := make(map[JourneyMilestoneName]struct{})
	pathCounts := make(map[string]int)
	for _, event := range normalizedEvents {
		path := eventPath(event)
		if path != "" {
			if summary.EntryPage == "" {
				summary.EntryPage = path
			}
			summary.ExitPage = path
			pathCounts[path]++
		}

		switch event.EventType {
		case EventPageView:
			addMilestone(&summary, milestoneSeen, MilestoneFirstPageView, event)
		case EventProductView:
			summary.ProductsViewed++
			addMilestone(&summary, milestoneSeen, MilestoneFirstProductView, event)
		case EventSearch:
			summary.Searches++
			addMilestone(&summary, milestoneSeen, MilestoneFirstSearch, event)
		case EventAddToCart:
			summary.CartActions++
			addMilestone(&summary, milestoneSeen, MilestoneFirstAddToCart, event)
		case EventCheckoutStep:
			summary.CheckoutStarted = true
			addMilestone(&summary, milestoneSeen, MilestoneCheckoutStarted, event)
		case EventPaymentResult:
			if paymentCompleted(event.Properties["status"]) {
				summary.PaymentCompleted = true
				addMilestone(&summary, milestoneSeen, MilestonePaymentCompleted, event)
			}
		}
	}
	summary.TopPaths = topPaths(pathCounts, topPathLimit)
	return summary.Normalize()
}

func (s JourneySummary) Normalize() JourneySummary {
	out := s
	out.SessionID = strings.TrimSpace(out.SessionID)
	out.AnonymousID = strings.TrimSpace(out.AnonymousID)
	out.UserID = trimStringPtr(out.UserID)
	out.EntryPage = strings.TrimSpace(out.EntryPage)
	out.ExitPage = strings.TrimSpace(out.ExitPage)
	out.FirstEventAt = normalizeTime(out.FirstEventAt)
	out.LastEventAt = normalizeTime(out.LastEventAt)
	out.Milestones = normalizeJourneyMilestones(out.Milestones)
	out.TopPaths = normalizeJourneyTopPaths(out.TopPaths)
	out.CalculatedAt = normalizeTime(out.CalculatedAt)
	out.CreatedAt = normalizeTime(out.CreatedAt)
	out.UpdatedAt = normalizeTime(out.UpdatedAt)
	out.RetainUntil = normalizeTimePtr(out.RetainUntil)
	out.AnonymizedAt = normalizeTimePtr(out.AnonymizedAt)
	out.LegalHoldReason = trimStringPtr(out.LegalHoldReason)
	out.LegalHoldSetBy = trimStringPtr(out.LegalHoldSetBy)
	out.LegalHoldSetAt = normalizeTimePtr(out.LegalHoldSetAt)
	out.RetentionClass = RetentionClass(strings.TrimSpace(string(out.RetentionClass)))
	return out
}

func (s JourneySummary) Validate() error {
	return s.ValidateWithConfig(DefaultJourneySummaryValidationConfig())
}

func (s JourneySummary) ValidateWithConfig(cfg JourneySummaryValidationConfig) error {
	cfg = cfg.WithDefaults()
	summary := s.Normalize()
	violations := make([]FieldViolation, 0)

	validateID(&violations, "session_id", summary.SessionID, true, cfg.Session.MaxIDLength)
	validateID(&violations, "anonymous_id", summary.AnonymousID, true, cfg.Session.MaxIDLength)
	validateOptionalID(&violations, "user_id", summary.UserID, cfg.Session.MaxIDLength)
	validateOptionalJourneyPage(&violations, "entry_page", summary.EntryPage, cfg.Session.MaxPageLength)
	validateOptionalJourneyPage(&violations, "exit_page", summary.ExitPage, cfg.Session.MaxPageLength)

	if summary.SchemaVersion != CurrentJourneySummarySchemaVersion {
		addViolation(&violations, "schema_version", "must match current journey summary schema version")
	}
	if summary.DurationSeconds < 0 {
		addViolation(&violations, "duration_seconds", "cannot be negative")
	}
	validateNonNegativeInt(&violations, "total_events", summary.TotalEvents)
	validateNonNegativeInt(&violations, "products_viewed", summary.ProductsViewed)
	validateNonNegativeInt(&violations, "searches", summary.Searches)
	validateNonNegativeInt(&violations, "cart_actions", summary.CartActions)
	if !summary.FirstEventAt.IsZero() && !summary.LastEventAt.IsZero() && summary.LastEventAt.Before(summary.FirstEventAt) {
		addViolation(&violations, "last_event_at", "cannot be before first_event_at")
	}
	if len(summary.Milestones) > cfg.MaxMilestones {
		addViolation(&violations, "milestones", "has too many values")
	}
	for i, milestone := range summary.Milestones {
		validateJourneyMilestone(&violations, i, milestone, cfg)
	}
	if len(summary.TopPaths) > cfg.MaxTopPaths {
		addViolation(&violations, "top_paths", "has too many values")
	}
	for i, path := range summary.TopPaths {
		validateJourneyTopPath(&violations, i, path, cfg)
	}
	if summary.CalculatedAt.IsZero() {
		addViolation(&violations, "calculated_at", "is required")
	}
	if summary.UpdatedAt.IsZero() {
		addViolation(&violations, "updated_at", "is required")
	}
	validateOptionalRetentionMetadata(&violations, summary.RetentionClass, summary.LegalHold, summary.LegalHoldReason, summary.LegalHoldSetAt)

	if len(violations) > 0 {
		return JourneyValidationError{Violations: violations}
	}
	return nil
}

func sortSessionEvents(events []SessionEvent) {
	sort.SliceStable(events, func(i, j int) bool {
		left := events[i]
		right := events[j]
		if !left.OccurredAt.Equal(right.OccurredAt) {
			return left.OccurredAt.Before(right.OccurredAt)
		}
		if !left.ReceivedAt.Equal(right.ReceivedAt) {
			return left.ReceivedAt.Before(right.ReceivedAt)
		}
		return left.EventID < right.EventID
	})
}

func eventPath(event SessionEvent) string {
	if event.Path == nil {
		return ""
	}
	return strings.TrimSpace(*event.Path)
}

func addMilestone(summary *JourneySummary, seen map[JourneyMilestoneName]struct{}, name JourneyMilestoneName, event SessionEvent) {
	if _, exists := seen[name]; exists {
		return
	}
	seen[name] = struct{}{}
	summary.Milestones = append(summary.Milestones, JourneyMilestone{
		Name:       name,
		EventID:    event.EventID,
		Path:       eventPath(event),
		OccurredAt: event.OccurredAt,
	})
}

func paymentCompleted(value any) bool {
	status, ok := value.(string)
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "success", "paid", "completed":
		return true
	default:
		return false
	}
}

func topPaths(counts map[string]int, limit int) []JourneyPathCount {
	if len(counts) == 0 || limit <= 0 {
		return nil
	}
	paths := make([]JourneyPathCount, 0, len(counts))
	for path, count := range counts {
		paths = append(paths, JourneyPathCount{Path: path, Count: count})
	}
	sort.Slice(paths, func(i, j int) bool {
		if paths[i].Count != paths[j].Count {
			return paths[i].Count > paths[j].Count
		}
		return paths[i].Path < paths[j].Path
	})
	if len(paths) > limit {
		paths = paths[:limit]
	}
	return paths
}

func normalizeJourneyMilestones(in []JourneyMilestone) []JourneyMilestone {
	if len(in) == 0 {
		return nil
	}
	out := make([]JourneyMilestone, 0, len(in))
	for _, milestone := range in {
		milestone.Name = JourneyMilestoneName(strings.TrimSpace(string(milestone.Name)))
		milestone.EventID = strings.TrimSpace(milestone.EventID)
		milestone.Path = strings.TrimSpace(milestone.Path)
		milestone.OccurredAt = normalizeTime(milestone.OccurredAt)
		if milestone.Name == "" {
			continue
		}
		out = append(out, milestone)
	}
	return out
}

func normalizeJourneyTopPaths(in []JourneyPathCount) []JourneyPathCount {
	if len(in) == 0 {
		return nil
	}
	out := make([]JourneyPathCount, 0, len(in))
	for _, path := range in {
		path.Path = strings.TrimSpace(path.Path)
		if path.Path == "" {
			continue
		}
		out = append(out, path)
	}
	return out
}

func validateOptionalJourneyPage(violations *[]FieldViolation, field string, value string, maxLength int) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	validatePagePath(violations, field, value, false, maxLength)
}

func validateJourneyMilestone(violations *[]FieldViolation, index int, milestone JourneyMilestone, cfg JourneySummaryValidationConfig) {
	field := fmt.Sprintf("milestones[%d]", index)
	if strings.TrimSpace(string(milestone.Name)) == "" {
		addViolation(violations, field+".name", "is required")
	}
	validateID(violations, field+".event_id", milestone.EventID, true, cfg.Session.MaxIDLength)
	validateOptionalJourneyPage(violations, field+".path", milestone.Path, cfg.Session.MaxPageLength)
	if milestone.OccurredAt.IsZero() {
		addViolation(violations, field+".occurred_at", "is required")
	}
}

func validateJourneyTopPath(violations *[]FieldViolation, index int, path JourneyPathCount, cfg JourneySummaryValidationConfig) {
	field := fmt.Sprintf("top_paths[%d]", index)
	validatePagePath(violations, field+".path", path.Path, true, cfg.Session.MaxPageLength)
	if path.Count <= 0 {
		addViolation(violations, field+".count", "must be greater than zero")
	}
}

func validateNonNegativeInt(violations *[]FieldViolation, field string, value int) {
	if value < 0 {
		addViolation(violations, field, "cannot be negative")
	}
}
