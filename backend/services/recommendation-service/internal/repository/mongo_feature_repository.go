package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type MongoFeatureRepositoryConfig struct {
	InteractionRetention time.Duration
}

type MongoFeatureRepository struct {
	db                     *mongo.Database
	userInteractions       *mongo.Collection
	productFeatures        *mongo.Collection
	userFeatureProfiles    *mongo.Collection
	userProductCounters    *mongo.Collection
	productCooccurrence    *mongo.Collection
	featureJobRuns         *mongo.Collection
	featureProcessedEvents *mongo.Collection
	recommendationSets     *mongo.Collection
	abTestAssignments      *mongo.Collection
	logger                 *slog.Logger
	interactionTTL         time.Duration
}

func NewMongoFeatureRepository(db *mongo.Database, cfg MongoFeatureRepositoryConfig, logger *slog.Logger) (*MongoFeatureRepository, error) {
	if db == nil {
		return nil, errors.New("mongo database is required")
	}
	if cfg.InteractionRetention <= 0 {
		cfg.InteractionRetention = domain.DefaultInteractionRetention
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &MongoFeatureRepository{
		db:                     db,
		userInteractions:       db.Collection(domain.CollectionUserInteractions),
		productFeatures:        db.Collection(domain.CollectionProductFeatures),
		userFeatureProfiles:    db.Collection(domain.CollectionUserFeatureProfiles),
		userProductCounters:    db.Collection(domain.CollectionUserProductCounters),
		productCooccurrence:    db.Collection(domain.CollectionProductCooccurrence),
		featureJobRuns:         db.Collection(domain.CollectionFeatureJobRuns),
		featureProcessedEvents: db.Collection(domain.CollectionFeatureProcessedEvents),
		recommendationSets:     db.Collection(domain.CollectionRecommendationSets),
		abTestAssignments:      db.Collection(domain.CollectionABTestAssignments),
		logger:                 logger,
		interactionTTL:         cfg.InteractionRetention,
	}, nil
}

func (r *MongoFeatureRepository) Ping(ctx context.Context) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("%w: mongo repository is not configured", domain.ErrRecommendationStorage)
	}
	if err := r.db.Client().Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("%w: ping mongo: %v", domain.ErrRecommendationStorage, err)
	}
	return nil
}

func (r *MongoFeatureRepository) EnsureIndexes(ctx context.Context) error {
	if r == nil {
		return fmt.Errorf("%w: mongo repository is not configured", domain.ErrRecommendationStorage)
	}
	indexes := r.indexModels()
	for collectionName, models := range indexes {
		collection := r.db.Collection(collectionName)
		names, err := collection.Indexes().CreateMany(ctx, models)
		if err != nil {
			return fmt.Errorf("%w: ensure %s indexes: %v", domain.ErrRecommendationStorage, collectionName, err)
		}
		r.logger.InfoContext(ctx, "recommendation.mongo.indexes_ensured",
			slog.String("collection", collectionName),
			slog.Any("indexes", names),
		)
	}
	return nil
}

func (r *MongoFeatureRepository) InsertInteraction(ctx context.Context, interaction domain.UserInteraction) (domain.InteractionInsertResult, error) {
	if r == nil || r.userInteractions == nil {
		return domain.InteractionInsertResult{}, fmt.Errorf("%w: mongo repository is not configured", domain.ErrRecommendationStorage)
	}
	if err := interaction.Validate(); err != nil {
		return domain.InteractionInsertResult{}, err
	}
	if _, err := r.userInteractions.InsertOne(ctx, interaction); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.InteractionInsertResult{Duplicate: true}, nil
		}
		return domain.InteractionInsertResult{}, fmt.Errorf("%w: insert user interaction: %v", domain.ErrRecommendationStorage, err)
	}
	return domain.InteractionInsertResult{}, nil
}

func (r *MongoFeatureRepository) ListEligibleProductFeatures(ctx context.Context) ([]domain.ProductFeature, error) {
	if r == nil || r.productFeatures == nil {
		return nil, fmt.Errorf("%w: mongo repository is not configured", domain.ErrRecommendationStorage)
	}
	filter := bson.D{
		{Key: "product_id", Value: bson.D{{Key: "$type", Value: "string"}, {Key: "$ne", Value: ""}}},
		{Key: "status", Value: domain.ProductStatusActive},
		{Key: "stock_status", Value: domain.ProductStockStatusInStock},
		{Key: "quality_flags.is_recommendable", Value: true},
		{Key: "quality_flags.is_deleted", Value: bson.D{{Key: "$ne", Value: true}}},
	}
	cursor, err := r.productFeatures.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%w: list eligible product features: %v", domain.ErrRecommendationStorage, err)
	}
	defer cursor.Close(ctx)

	var features []domain.ProductFeature
	if err := cursor.All(ctx, &features); err != nil {
		return nil, fmt.Errorf("%w: decode eligible product features: %v", domain.ErrRecommendationStorage, err)
	}
	return features, nil
}

func (r *MongoFeatureRepository) GetRecommendationSet(ctx context.Context, contextKey string, strategyID domain.StrategyID, now time.Time) (domain.RecommendationSet, error) {
	contextKey = strings.TrimSpace(contextKey)
	if contextKey == "" {
		return domain.RecommendationSet{}, fmt.Errorf("%w: context_key is required", domain.ErrInvalidRecommendationSet)
	}
	if err := domain.ValidateStrategyID(strategyID); err != nil {
		return domain.RecommendationSet{}, err
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	filter := bson.D{
		{Key: "context_key", Value: contextKey},
		{Key: "strategy_id", Value: string(strategyID)},
		{Key: "expires_at", Value: bson.D{{Key: "$gt", Value: now}}},
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "generated_at", Value: -1}})

	var set domain.RecommendationSet
	if err := r.recommendationSets.FindOne(ctx, filter, opts).Decode(&set); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.RecommendationSet{}, fmt.Errorf("%w: context_key=%s strategy_id=%s", domain.ErrRecommendationSetNotFound, contextKey, strategyID)
		}
		return domain.RecommendationSet{}, fmt.Errorf("%w: get recommendation set: %v", domain.ErrRecommendationStorage, err)
	}
	if err := set.Validate(); err != nil {
		return domain.RecommendationSet{}, err
	}
	return set, nil
}

func (r *MongoFeatureRepository) UpsertRecommendationSet(ctx context.Context, set domain.RecommendationSet) error {
	if err := set.Validate(); err != nil {
		return err
	}

	filter := bson.D{
		{Key: "context_key", Value: set.ContextKey},
		{Key: "strategy_id", Value: string(set.StrategyID)},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "context_key", Value: set.ContextKey},
			{Key: "recommendation_type", Value: string(set.Type)},
			{Key: "strategy_id", Value: string(set.StrategyID)},
			{Key: "items", Value: set.Items},
			{Key: "metadata", Value: set.Metadata},
			{Key: "generated_at", Value: set.GeneratedAt},
			{Key: "expires_at", Value: set.ExpiresAt},
		}},
		{Key: "$setOnInsert", Value: bson.D{{Key: "_id", Value: set.ID}}},
	}
	if _, err := r.recommendationSets.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("%w: upsert recommendation set: %v", domain.ErrRecommendationStorage, err)
	}
	return nil
}

func (r *MongoFeatureRepository) indexModels() map[string][]mongo.IndexModel {
	ttlSeconds := int32(r.interactionTTL.Seconds())
	return map[string][]mongo.IndexModel{
		domain.CollectionUserInteractions: {
			{
				Keys: bson.D{{Key: "dedupe_key", Value: 1}},
				Options: options.Index().
					SetName("ux_user_interactions_dedupe_key").
					SetUnique(true).
					SetPartialFilterExpression(bson.D{{Key: "dedupe_key", Value: bson.D{{Key: "$type", Value: "string"}}}}),
			},
			{
				Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "occurred_at", Value: -1}},
				Options: options.Index().SetName("idx_user_interactions_user_recent"),
			},
			{
				Keys:    bson.D{{Key: "anonymous_id", Value: 1}, {Key: "occurred_at", Value: -1}},
				Options: options.Index().SetName("idx_user_interactions_anon_recent"),
			},
			{
				Keys:    bson.D{{Key: "product_id", Value: 1}, {Key: "normalized_event_type", Value: 1}, {Key: "occurred_at", Value: -1}},
				Options: options.Index().SetName("idx_user_interactions_product_normalized_event_recent"),
			},
			{
				Keys:    bson.D{{Key: "category_id", Value: 1}, {Key: "normalized_event_type", Value: 1}, {Key: "occurred_at", Value: -1}},
				Options: options.Index().SetName("idx_user_interactions_category_normalized_event_recent"),
			},
			{
				Keys:    bson.D{{Key: "occurred_at", Value: 1}},
				Options: options.Index().SetName("idx_user_interactions_ttl").SetExpireAfterSeconds(ttlSeconds),
			},
		},
		domain.CollectionRecommendationSets: {
			{
				Keys: bson.D{{Key: "context_key", Value: 1}, {Key: "strategy_id", Value: 1}},
				Options: options.Index().
					SetName("ux_recommendation_sets_context_strategy").
					SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "expires_at", Value: 1}},
				Options: options.Index().SetName("idx_recommendation_sets_expires_at_ttl").SetExpireAfterSeconds(0),
			},
			{
				Keys:    bson.D{{Key: "recommendation_type", Value: 1}, {Key: "generated_at", Value: -1}},
				Options: options.Index().SetName("idx_recommendation_sets_type_generated"),
			},
		},
		domain.CollectionABTestAssignments: {
			{
				Keys: bson.D{{Key: "experiment_id", Value: 1}, {Key: "assignment_key", Value: 1}},
				Options: options.Index().
					SetName("ux_ab_assignment_experiment_identity").
					SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "experiment_id", Value: 1}},
				Options: options.Index().SetName("idx_ab_assignment_user_experiment").SetSparse(true),
			},
			{
				Keys:    bson.D{{Key: "anonymous_id", Value: 1}, {Key: "experiment_id", Value: 1}},
				Options: options.Index().SetName("idx_ab_assignment_anon_experiment").SetSparse(true),
			},
			{
				Keys:    bson.D{{Key: "experiment_id", Value: 1}, {Key: "variant_id", Value: 1}, {Key: "assigned_at", Value: -1}},
				Options: options.Index().SetName("idx_ab_assignment_variant_recent"),
			},
			{
				Keys:    bson.D{{Key: "expires_at", Value: 1}},
				Options: options.Index().SetName("idx_ab_assignment_ttl").SetExpireAfterSeconds(0),
			},
		},
		domain.CollectionProductFeatures: {
			{
				Keys:    bson.D{{Key: "product_id", Value: 1}},
				Options: options.Index().SetName("ux_product_features_product_id").SetUnique(true),
			},
			{
				Keys: bson.D{
					{Key: "status", Value: 1},
					{Key: "stock_status", Value: 1},
					{Key: "quality_flags.is_recommendable", Value: 1},
					{Key: "quality_flags.is_deleted", Value: 1},
					{Key: "counters.purchases_7d", Value: -1},
					{Key: "product_id", Value: 1},
				},
				Options: options.Index().SetName("idx_product_features_rule_based_global"),
			},
			{
				Keys: bson.D{
					{Key: "category_id", Value: 1},
					{Key: "status", Value: 1},
					{Key: "stock_status", Value: 1},
					{Key: "quality_flags.is_recommendable", Value: 1},
					{Key: "quality_flags.is_deleted", Value: 1},
					{Key: "counters.purchases_7d", Value: -1},
					{Key: "product_id", Value: 1},
				},
				Options: options.Index().SetName("idx_product_features_rule_based_category"),
			},
			{
				Keys: bson.D{
					{Key: "seller_id", Value: 1},
					{Key: "status", Value: 1},
					{Key: "stock_status", Value: 1},
					{Key: "quality_flags.is_recommendable", Value: 1},
					{Key: "quality_flags.is_deleted", Value: 1},
					{Key: "counters.purchases_7d", Value: -1},
					{Key: "product_id", Value: 1},
				},
				Options: options.Index().SetName("idx_product_features_rule_based_seller"),
			},
			{
				Keys:    bson.D{{Key: "category_id", Value: 1}, {Key: "quality_flags.is_recommendable", Value: 1}, {Key: "counters.purchases_7d", Value: -1}},
				Options: options.Index().SetName("idx_product_features_category_recommendable_purchases"),
			},
			{
				Keys:    bson.D{{Key: "seller_id", Value: 1}, {Key: "quality_flags.is_recommendable", Value: 1}, {Key: "counters.purchases_7d", Value: -1}},
				Options: options.Index().SetName("idx_product_features_seller_recommendable_purchases"),
			},
			{
				Keys:    bson.D{{Key: "brand_id", Value: 1}, {Key: "quality_flags.is_recommendable", Value: 1}},
				Options: options.Index().SetName("idx_product_features_brand_recommendable"),
			},
			{
				Keys: bson.D{
					{Key: "brand_id", Value: 1},
					{Key: "status", Value: 1},
					{Key: "stock_status", Value: 1},
					{Key: "quality_flags.is_recommendable", Value: 1},
					{Key: "quality_flags.is_deleted", Value: 1},
					{Key: "counters.purchases_7d", Value: -1},
					{Key: "product_id", Value: 1},
				},
				Options: options.Index().SetName("idx_product_features_personalized_brand"),
			},
			{
				Keys:    bson.D{{Key: "embedding_refs.vector_id", Value: 1}},
				Options: options.Index().SetName("idx_product_features_embedding_vector_id").SetSparse(true),
			},
		},
		domain.CollectionUserFeatureProfiles: {
			{
				Keys:    bson.D{{Key: "profile_key", Value: 1}},
				Options: options.Index().SetName("ux_user_feature_profiles_profile_key").SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "user_id", Value: 1}},
				Options: options.Index().SetName("idx_user_feature_profiles_user_id").SetSparse(true),
			},
			{
				Keys:    bson.D{{Key: "anonymous_id", Value: 1}},
				Options: options.Index().SetName("idx_user_feature_profiles_anonymous_id").SetSparse(true),
			},
			{
				Keys:    bson.D{{Key: "expires_at", Value: 1}},
				Options: options.Index().SetName("idx_user_feature_profiles_expires_at_ttl").SetExpireAfterSeconds(0),
			},
		},
		domain.CollectionUserProductCounters: {
			{
				Keys:    bson.D{{Key: "profile_key", Value: 1}, {Key: "product_id", Value: 1}},
				Options: options.Index().SetName("ux_user_product_counters_profile_product").SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "profile_key", Value: 1}, {Key: "last_interaction_at", Value: -1}},
				Options: options.Index().SetName("idx_user_product_counters_profile_recent"),
			},
			{
				Keys:    bson.D{{Key: "product_id", Value: 1}, {Key: "weighted_score_input", Value: -1}},
				Options: options.Index().SetName("idx_user_product_counters_product_score"),
			},
			{
				Keys:    bson.D{{Key: "category_id", Value: 1}, {Key: "last_interaction_at", Value: -1}},
				Options: options.Index().SetName("idx_user_product_counters_category_recent"),
			},
			{
				Keys:    bson.D{{Key: "expires_at", Value: 1}},
				Options: options.Index().SetName("idx_user_product_counters_expires_at_ttl").SetExpireAfterSeconds(0),
			},
		},
		domain.CollectionProductCooccurrence: {
			{
				Keys:    bson.D{{Key: "source_product_id", Value: 1}, {Key: "relationship_type", Value: 1}, {Key: "score_input", Value: -1}},
				Options: options.Index().SetName("idx_product_cooccurrence_source_type_score"),
			},
			{
				Keys:    bson.D{{Key: "related_product_id", Value: 1}},
				Options: options.Index().SetName("idx_product_cooccurrence_related"),
			},
			{
				Keys:    bson.D{{Key: "category_id", Value: 1}, {Key: "score_input", Value: -1}},
				Options: options.Index().SetName("idx_product_cooccurrence_category_score"),
			},
			{
				Keys:    bson.D{{Key: "last_seen_at", Value: -1}},
				Options: options.Index().SetName("idx_product_cooccurrence_last_seen"),
			},
		},
		domain.CollectionFeatureJobRuns: {
			{
				Keys:    bson.D{{Key: "job_type", Value: 1}, {Key: "started_at", Value: -1}},
				Options: options.Index().SetName("idx_feature_job_runs_type_started"),
			},
			{
				Keys:    bson.D{{Key: "status", Value: 1}, {Key: "started_at", Value: -1}},
				Options: options.Index().SetName("idx_feature_job_runs_status_started"),
			},
		},
		domain.CollectionFeatureProcessedEvents: {
			{
				Keys:    bson.D{{Key: "event_id", Value: 1}},
				Options: options.Index().SetName("ux_feature_processed_events_event_id").SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "expires_at", Value: 1}},
				Options: options.Index().SetName("idx_feature_processed_events_expires_at_ttl").SetExpireAfterSeconds(0),
			},
		},
	}
}
