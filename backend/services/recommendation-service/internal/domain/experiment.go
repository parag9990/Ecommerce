package domain

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

type ExperimentStatus string

const (
	ExperimentDraft   ExperimentStatus = "draft"
	ExperimentActive  ExperimentStatus = "active"
	ExperimentPaused  ExperimentStatus = "paused"
	ExperimentStopped ExperimentStatus = "stopped"
)

type ExperimentVariant struct {
	VariantID  string     `json:"variant_id" bson:"variant_id"`
	StrategyID StrategyID `json:"strategy_id" bson:"strategy_id"`
	Weight     int        `json:"weight" bson:"weight"`
}

type ExperimentDefinition struct {
	ExperimentID string                `json:"experiment_id" bson:"experiment_id"`
	Context      RecommendationContext `json:"context" bson:"context"`
	Type         RecommendationType    `json:"type" bson:"recommendation_type"`
	Status       ExperimentStatus      `json:"status" bson:"status"`
	Salt         string                `json:"salt" bson:"salt"`
	Variants     []ExperimentVariant   `json:"variants" bson:"variants"`
	StartsAt     time.Time             `json:"starts_at,omitempty" bson:"starts_at,omitempty"`
	EndsAt       time.Time             `json:"ends_at,omitempty" bson:"ends_at,omitempty"`
}

type AssignmentIdentity struct {
	AssignmentKey string
	UserID        string
	AnonymousID   string
	SessionID     string
}

type ExperimentAssignment struct {
	ID                 string                `json:"id" bson:"_id"`
	ExperimentID       string                `json:"experiment_id" bson:"experiment_id"`
	AssignmentKey      string                `json:"assignment_key" bson:"assignment_key"`
	UserID             string                `json:"user_id,omitempty" bson:"user_id,omitempty"`
	AnonymousID        string                `json:"anonymous_id,omitempty" bson:"anonymous_id,omitempty"`
	SessionID          string                `json:"session_id,omitempty" bson:"session_id,omitempty"`
	Context            RecommendationContext `json:"context" bson:"context"`
	RecommendationType RecommendationType    `json:"recommendation_type" bson:"recommendation_type"`
	VariantID          string                `json:"variant_id" bson:"variant_id"`
	StrategyID         StrategyID            `json:"strategy_id" bson:"strategy_id"`
	AssignedAt         time.Time             `json:"assigned_at" bson:"assigned_at"`
	ExpiresAt          time.Time             `json:"expires_at" bson:"expires_at"`
}

type StrategyAssignment struct {
	ExperimentID string
	VariantID    string
	StrategyID   StrategyID
	Assigned     bool
}

func (s ExperimentStatus) IsValid() bool {
	switch s {
	case ExperimentDraft, ExperimentActive, ExperimentPaused, ExperimentStopped:
		return true
	default:
		return false
	}
}

func (d ExperimentDefinition) Normalize(defaultSalt string) ExperimentDefinition {
	d.ExperimentID = strings.TrimSpace(d.ExperimentID)
	d.Context = RecommendationContext(strings.TrimSpace(string(d.Context)))
	d.Type = RecommendationType(strings.TrimSpace(string(d.Type)))
	d.Status = ExperimentStatus(strings.ToLower(strings.TrimSpace(string(d.Status))))
	if d.Status == "" {
		d.Status = ExperimentDraft
	}
	d.Salt = strings.TrimSpace(d.Salt)
	if d.Salt == "" {
		d.Salt = strings.TrimSpace(defaultSalt)
	}
	d.Variants = append([]ExperimentVariant(nil), d.Variants...)
	for i := range d.Variants {
		d.Variants[i].VariantID = strings.TrimSpace(d.Variants[i].VariantID)
		d.Variants[i].StrategyID = StrategyID(strings.TrimSpace(string(d.Variants[i].StrategyID)))
	}
	if !d.StartsAt.IsZero() {
		d.StartsAt = d.StartsAt.UTC()
	}
	if !d.EndsAt.IsZero() {
		d.EndsAt = d.EndsAt.UTC()
	}
	return d
}

func (d ExperimentDefinition) Validate() error {
	d = d.Normalize(DefaultABTestSalt)
	if d.ExperimentID == "" {
		return fmt.Errorf("%w: experiment_id is required", ErrInvalidExperiment)
	}
	if hasWhitespaceOrControl(d.ExperimentID) {
		return fmt.Errorf("%w: experiment_id cannot contain whitespace or control characters", ErrInvalidExperiment)
	}
	if !d.Context.IsValid() {
		return fmt.Errorf("%w: context is required", ErrInvalidExperiment)
	}
	if !d.Type.IsValid() {
		return fmt.Errorf("%w: recommendation type is required", ErrInvalidExperiment)
	}
	if !d.Status.IsValid() {
		return fmt.Errorf("%w: unsupported status %q", ErrInvalidExperiment, d.Status)
	}
	if d.Salt == "" {
		return fmt.Errorf("%w: salt is required", ErrInvalidExperiment)
	}
	if len(d.Variants) == 0 {
		return fmt.Errorf("%w: variants are required", ErrInvalidExperiment)
	}
	if !d.StartsAt.IsZero() && !d.EndsAt.IsZero() && !d.EndsAt.After(d.StartsAt) {
		return fmt.Errorf("%w: ends_at must be after starts_at", ErrInvalidExperiment)
	}

	seen := make(map[string]struct{}, len(d.Variants))
	totalWeight := 0
	for _, variant := range d.Variants {
		if variant.VariantID == "" {
			return fmt.Errorf("%w: variant_id is required", ErrInvalidExperiment)
		}
		if hasWhitespaceOrControl(variant.VariantID) {
			return fmt.Errorf("%w: variant_id cannot contain whitespace or control characters", ErrInvalidExperiment)
		}
		if _, ok := seen[variant.VariantID]; ok {
			return fmt.Errorf("%w: duplicate variant_id %q", ErrInvalidExperiment, variant.VariantID)
		}
		seen[variant.VariantID] = struct{}{}
		if variant.Weight <= 0 {
			return fmt.Errorf("%w: variant %s weight must be greater than zero", ErrInvalidExperiment, variant.VariantID)
		}
		strategyType, err := RecommendationTypeForStrategy(variant.StrategyID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidExperiment, err)
		}
		if strategyType != d.Type && strategyType != RecommendationTypeTrending {
			return fmt.Errorf("%w: variant %s strategy type %s is incompatible with experiment type %s", ErrInvalidExperiment, variant.VariantID, strategyType, d.Type)
		}
		totalWeight += variant.Weight
	}
	if totalWeight != 100 {
		return fmt.Errorf("%w: variant weights must total 100", ErrInvalidExperiment)
	}
	return nil
}

func (d ExperimentDefinition) ActiveAt(now time.Time) bool {
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	d = d.Normalize(DefaultABTestSalt)
	if d.Status != ExperimentActive {
		return false
	}
	if !d.StartsAt.IsZero() && now.Before(d.StartsAt) {
		return false
	}
	if !d.EndsAt.IsZero() && !now.Before(d.EndsAt) {
		return false
	}
	return true
}

func (d ExperimentDefinition) Matches(req RecommendationRequest) bool {
	req = req.Normalize()
	d = d.Normalize(DefaultABTestSalt)
	return d.Context == req.Context && d.Type == req.Type
}

func NewAssignmentIdentity(userID string, anonymousID string, sessionID string) (AssignmentIdentity, bool) {
	userID = strings.TrimSpace(userID)
	anonymousID = strings.TrimSpace(anonymousID)
	sessionID = strings.TrimSpace(sessionID)
	switch {
	case userID != "":
		return AssignmentIdentity{AssignmentKey: "user:" + userID, UserID: userID, AnonymousID: anonymousID, SessionID: sessionID}, true
	case anonymousID != "":
		return AssignmentIdentity{AssignmentKey: "anon:" + anonymousID, AnonymousID: anonymousID, SessionID: sessionID}, true
	case sessionID != "":
		return AssignmentIdentity{AssignmentKey: "session:" + sessionID, SessionID: sessionID}, true
	default:
		return AssignmentIdentity{}, false
	}
}

func BucketPercent(experimentID string, assignmentKey string, salt string) int {
	raw := strings.TrimSpace(experimentID) + ":" + strings.TrimSpace(assignmentKey) + ":" + strings.TrimSpace(salt)
	sum := sha256.Sum256([]byte(raw))
	value := binary.BigEndian.Uint32(sum[:4])
	return int(value % 100)
}

func ChooseExperimentVariant(bucket int, variants []ExperimentVariant) (ExperimentVariant, error) {
	if bucket < 0 || bucket > 99 {
		return ExperimentVariant{}, fmt.Errorf("%w: bucket must be between 0 and 99", ErrInvalidExperiment)
	}
	cursor := 0
	for _, variant := range variants {
		cursor += variant.Weight
		if bucket < cursor {
			return variant, nil
		}
	}
	return ExperimentVariant{}, fmt.Errorf("%w: no variant selected", ErrInvalidExperiment)
}

func NewExperimentAssignment(definition ExperimentDefinition, identity AssignmentIdentity, variant ExperimentVariant, now time.Time, ttl time.Duration) ExperimentAssignment {
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if ttl <= 0 {
		ttl = DefaultABAssignmentTTL
	}
	definition = definition.Normalize(DefaultABTestSalt)
	assignmentKey := strings.TrimSpace(identity.AssignmentKey)
	experimentID := strings.TrimSpace(definition.ExperimentID)
	return ExperimentAssignment{
		ID:                 experimentID + ":" + assignmentKey,
		ExperimentID:       experimentID,
		AssignmentKey:      assignmentKey,
		UserID:             strings.TrimSpace(identity.UserID),
		AnonymousID:        strings.TrimSpace(identity.AnonymousID),
		SessionID:          strings.TrimSpace(identity.SessionID),
		Context:            definition.Context,
		RecommendationType: definition.Type,
		VariantID:          strings.TrimSpace(variant.VariantID),
		StrategyID:         StrategyID(strings.TrimSpace(string(variant.StrategyID))),
		AssignedAt:         now,
		ExpiresAt:          now.Add(ttl),
	}
}

func (a ExperimentAssignment) Validate() error {
	if strings.TrimSpace(a.ID) == "" {
		return fmt.Errorf("%w: assignment id is required", ErrInvalidExperiment)
	}
	if strings.TrimSpace(a.ExperimentID) == "" {
		return fmt.Errorf("%w: experiment_id is required", ErrInvalidExperiment)
	}
	if strings.TrimSpace(a.AssignmentKey) == "" {
		return fmt.Errorf("%w: assignment_key is required", ErrInvalidExperiment)
	}
	if strings.TrimSpace(a.VariantID) == "" {
		return fmt.Errorf("%w: variant_id is required", ErrInvalidExperiment)
	}
	if !a.Context.IsValid() {
		return fmt.Errorf("%w: context is required", ErrInvalidExperiment)
	}
	if !a.RecommendationType.IsValid() {
		return fmt.Errorf("%w: recommendation_type is required", ErrInvalidExperiment)
	}
	if err := ValidateStrategyID(a.StrategyID); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidExperiment, err)
	}
	if a.AssignedAt.IsZero() {
		return fmt.Errorf("%w: assigned_at is required", ErrInvalidExperiment)
	}
	if a.ExpiresAt.IsZero() || !a.ExpiresAt.After(a.AssignedAt) {
		return fmt.Errorf("%w: expires_at must be after assigned_at", ErrInvalidExperiment)
	}
	return nil
}

func hasWhitespaceOrControl(value string) bool {
	return strings.ContainsAny(value, "\x00\r\n\t ")
}
