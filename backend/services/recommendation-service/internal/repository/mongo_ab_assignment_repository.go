package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *MongoFeatureRepository) GetExperimentAssignment(ctx context.Context, experimentID string, assignmentKey string, now time.Time) (domain.ExperimentAssignment, error) {
	if r == nil || r.abTestAssignments == nil {
		return domain.ExperimentAssignment{}, fmt.Errorf("%w: mongo repository is not configured", domain.ErrRecommendationStorage)
	}
	experimentID = strings.TrimSpace(experimentID)
	assignmentKey = strings.TrimSpace(assignmentKey)
	if experimentID == "" || assignmentKey == "" {
		return domain.ExperimentAssignment{}, fmt.Errorf("%w: experiment_id and assignment_key are required", domain.ErrInvalidExperiment)
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	filter := bson.D{
		{Key: "experiment_id", Value: experimentID},
		{Key: "assignment_key", Value: assignmentKey},
		{Key: "expires_at", Value: bson.D{{Key: "$gt", Value: now}}},
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "assigned_at", Value: -1}})
	var assignment domain.ExperimentAssignment
	if err := r.abTestAssignments.FindOne(ctx, filter, opts).Decode(&assignment); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.ExperimentAssignment{}, fmt.Errorf("%w: experiment_id=%s assignment_key=%s", domain.ErrExperimentAssignmentNotFound, experimentID, assignmentKey)
		}
		return domain.ExperimentAssignment{}, fmt.Errorf("%w: get ab assignment: %v", domain.ErrRecommendationStorage, err)
	}
	if err := assignment.Validate(); err != nil {
		return domain.ExperimentAssignment{}, err
	}
	return assignment, nil
}

func (r *MongoFeatureRepository) UpsertExperimentAssignment(ctx context.Context, assignment domain.ExperimentAssignment) error {
	if r == nil || r.abTestAssignments == nil {
		return fmt.Errorf("%w: mongo repository is not configured", domain.ErrRecommendationStorage)
	}
	if err := assignment.Validate(); err != nil {
		return err
	}
	filter := bson.D{
		{Key: "experiment_id", Value: assignment.ExperimentID},
		{Key: "assignment_key", Value: assignment.AssignmentKey},
	}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "experiment_id", Value: assignment.ExperimentID},
		{Key: "assignment_key", Value: assignment.AssignmentKey},
		{Key: "user_id", Value: assignment.UserID},
		{Key: "anonymous_id", Value: assignment.AnonymousID},
		{Key: "session_id", Value: assignment.SessionID},
		{Key: "context", Value: string(assignment.Context)},
		{Key: "recommendation_type", Value: string(assignment.RecommendationType)},
		{Key: "variant_id", Value: assignment.VariantID},
		{Key: "strategy_id", Value: string(assignment.StrategyID)},
		{Key: "assigned_at", Value: assignment.AssignedAt},
		{Key: "expires_at", Value: assignment.ExpiresAt},
	}}, {Key: "$setOnInsert", Value: bson.D{{Key: "_id", Value: assignment.ID}}}}
	if _, err := r.abTestAssignments.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("%w: upsert ab assignment: %v", domain.ErrRecommendationStorage, err)
	}
	return nil
}
