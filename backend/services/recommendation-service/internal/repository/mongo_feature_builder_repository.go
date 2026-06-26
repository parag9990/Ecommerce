package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var errFeatureBatchAlreadyApplied = errors.New("feature batch already applied")

func (r *MongoFeatureRepository) ListPendingFeatureBatches(ctx context.Context, limit int) ([][]domain.UserInteraction, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("%w: mongo repository is not configured", domain.ErrRecommendationStorage)
	}
	if limit <= 0 {
		return nil, fmt.Errorf("%w: pending feature batch limit must be greater than zero", domain.ErrInvalidFeature)
	}
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: domain.CollectionFeatureProcessedEvents},
			{Key: "localField", Value: "event_id"},
			{Key: "foreignField", Value: "event_id"},
			{Key: "as", Value: "processed_events"},
		}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "processed_events", Value: bson.D{{Key: "$size", Value: 0}}}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "occurred_at", Value: 1}, {Key: "_id", Value: 1}}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$event_id"},
			{Key: "first_occurred_at", Value: bson.D{{Key: "$first", Value: "$occurred_at"}}},
			{Key: "interactions", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "first_occurred_at", Value: 1}, {Key: "_id", Value: 1}}}},
		bson.D{{Key: "$limit", Value: int64(limit)}},
	}
	cursor, err := r.userInteractions.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("%w: list pending feature batches: %v", domain.ErrRecommendationStorage, err)
	}
	defer cursor.Close(ctx)
	var results []struct {
		Interactions []domain.UserInteraction `bson:"interactions"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("%w: decode pending feature batches: %v", domain.ErrRecommendationStorage, err)
	}
	batches := make([][]domain.UserInteraction, 0, len(results))
	for _, result := range results {
		if len(result.Interactions) > 0 {
			batches = append(batches, result.Interactions)
		}
	}
	return batches, nil
}

func (r *MongoFeatureRepository) ApplyFeatureBatch(ctx context.Context, batch domain.FeatureBatch) (domain.FeatureBuildResult, error) {
	if r == nil || r.db == nil {
		return domain.FeatureBuildResult{}, fmt.Errorf("%w: mongo repository is not configured", domain.ErrRecommendationStorage)
	}
	if err := batch.Validate(); err != nil {
		return domain.FeatureBuildResult{}, err
	}

	stats := domain.FeatureJobStats{EventsRead: len(batch.Interactions)}
	session, err := r.db.Client().StartSession()
	if err != nil {
		return domain.FeatureBuildResult{}, fmt.Errorf("%w: start feature transaction: %v", domain.ErrRecommendationStorage, err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		claim := bson.D{
			{Key: "_id", Value: batch.Interactions[0].EventID},
			{Key: "event_id", Value: batch.Interactions[0].EventID},
			{Key: "processed_by", Value: "feature_builder_v1"},
			{Key: "processed_at", Value: batch.ProcessedAt},
			{Key: "expires_at", Value: batch.ProcessedAt.Add(batch.ProcessedEventRetention)},
		}
		if _, err := r.featureProcessedEvents.InsertOne(txCtx, claim); err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return nil, errFeatureBatchAlreadyApplied
			}
			return nil, err
		}

		for _, interaction := range batch.Interactions {
			if err := r.upsertProductFeature(txCtx, interaction, batch.ProcessedAt); err != nil {
				return nil, err
			}
			stats.ProductFeaturesUpdated++
			if err := r.upsertUserProfile(txCtx, interaction, batch); err != nil {
				return nil, err
			}
			stats.UserProfilesUpdated++
			if err := r.upsertUserProductCounter(txCtx, interaction, batch); err != nil {
				return nil, err
			}
			stats.UserProductCountersUpdated++
		}
		pairs, err := r.upsertPurchasePairs(txCtx, batch.Interactions, batch.ProcessedAt)
		if err != nil {
			return nil, err
		}
		stats.CooccurrencePairsUpdated = pairs

		run := domain.FeatureJobRun{
			ID:         batch.ID,
			JobType:    domain.FeatureJobTypeIncremental,
			Status:     "success",
			StartedAt:  batch.StartedAt,
			FinishedAt: batch.ProcessedAt,
			Watermark: domain.FeatureWatermark{
				LastOccurredAt: latestInteractionTime(batch.Interactions),
				LastEventID:    batch.Interactions[0].EventID,
			},
			Stats:     stats,
			CreatedAt: batch.ProcessedAt,
		}
		_, err = r.featureJobRuns.InsertOne(txCtx, run)
		return nil, err
	}, options.Transaction())
	if errors.Is(err, errFeatureBatchAlreadyApplied) {
		return domain.FeatureBuildResult{Duplicates: len(batch.Interactions)}, nil
	}
	if err != nil {
		r.recordFailedFeatureRun(ctx, domain.FeatureJobRun{
			ID:        batch.ID,
			JobType:   domain.FeatureJobTypeIncremental,
			StartedAt: batch.StartedAt,
			Watermark: domain.FeatureWatermark{
				LastOccurredAt: latestInteractionTime(batch.Interactions),
				LastEventID:    batch.Interactions[0].EventID,
			},
			Stats: domain.FeatureJobStats{EventsRead: len(batch.Interactions)},
		}, err)
		return domain.FeatureBuildResult{}, fmt.Errorf("%w: update derived feature collections: %v", domain.ErrRecommendationStorage, err)
	}
	return domain.FeatureBuildResult{Applied: len(batch.Interactions), Stats: stats}, nil
}

func (r *MongoFeatureRepository) recordFailedFeatureRun(ctx context.Context, run domain.FeatureJobRun, cause error) {
	now := time.Now().UTC()
	run.Status = "failed"
	run.FinishedAt = now
	run.CreatedAt = now
	run.Error = cause.Error()
	if _, err := r.featureJobRuns.InsertOne(ctx, run); err != nil {
		r.logger.WarnContext(ctx, "recommendation.feature_builder.failed_run_persist_failed",
			"job_type", run.JobType,
			"error", err.Error(),
		)
	}
}

func (r *MongoFeatureRepository) upsertProductFeature(ctx context.Context, interaction domain.UserInteraction, updatedAt time.Time) error {
	inc := productLifetimeIncrements(interaction)
	set := bson.D{{Key: "updated_at", Value: updatedAt}}
	appendOptionalSet(&set, "category_id", interaction.CategoryID)
	appendOptionalSet(&set, "seller_id", interaction.SellerID)
	appendOptionalSet(&set, "brand_id", interaction.BrandID)
	update := bson.D{
		{Key: "$inc", Value: inc},
		{Key: "$set", Value: set},
		{Key: "$max", Value: bson.D{{Key: "last_event_at", Value: interaction.OccurredAt}}},
		{Key: "$setOnInsert", Value: productFeatureSetOnInsert(interaction, updatedAt)},
	}
	if _, err := r.productFeatures.UpdateOne(ctx, bson.D{{Key: "product_id", Value: interaction.ProductID}}, update, options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("upsert product features: %w", err)
	}
	return nil
}

func productFeatureSetOnInsert(interaction domain.UserInteraction, createdAt time.Time) bson.D {
	return bson.D{
		{Key: "_id", Value: interaction.ProductID},
		{Key: "product_id", Value: interaction.ProductID},
		{Key: "status", Value: domain.ProductStatusActive},
		{Key: "stock_status", Value: domain.ProductStockStatusInStock},
		{Key: "counters.views_24h", Value: int64(0)},
		{Key: "counters.views_7d", Value: int64(0)},
		{Key: "counters.views_30d", Value: int64(0)},
		{Key: "counters.cart_adds_24h", Value: int64(0)},
		{Key: "counters.cart_adds_7d", Value: int64(0)},
		{Key: "counters.cart_adds_30d", Value: int64(0)},
		{Key: "counters.wishlist_adds_24h", Value: int64(0)},
		{Key: "counters.wishlist_adds_7d", Value: int64(0)},
		{Key: "counters.wishlist_adds_30d", Value: int64(0)},
		{Key: "counters.purchases_24h", Value: int64(0)},
		{Key: "counters.purchases_7d", Value: int64(0)},
		{Key: "counters.purchases_30d", Value: int64(0)},
		{Key: "quality_flags", Value: domain.ProductQualityFlags{IsRecommendable: true}},
		{Key: "embedding_refs", Value: bson.A{}},
		{Key: "feature_version", Value: domain.FeatureSchemaVersion},
		{Key: "created_at", Value: createdAt},
	}
}

func (r *MongoFeatureRepository) upsertUserProfile(ctx context.Context, interaction domain.UserInteraction, batch domain.FeatureBatch) error {
	profileKey, err := domain.NewProfileKey(interaction.UserID, interaction.AnonymousID)
	if err != nil {
		return err
	}
	set := bson.D{{Key: "updated_at", Value: batch.ProcessedAt}}
	if interaction.UserID != "" {
		set = append(set, bson.E{Key: "user_id", Value: interaction.UserID})
	} else {
		set = append(set,
			bson.E{Key: "anonymous_id", Value: interaction.AnonymousID},
			bson.E{Key: "expires_at", Value: interaction.OccurredAt.Add(batch.GuestProfileRetention)},
		)
	}
	inc := bson.D{}
	score := float64(interaction.WeightedFeatureScore())
	appendScoreIncrement(&inc, "category_scores", interaction.CategoryID, score)
	appendScoreIncrement(&inc, "seller_scores", interaction.SellerID, score)
	appendScoreIncrement(&inc, "brand_scores", interaction.BrandID, score)

	update := bson.D{
		{Key: "$set", Value: set},
		{Key: "$max", Value: bson.D{{Key: "last_event_at", Value: interaction.OccurredAt}}},
		{Key: "$push", Value: bson.D{{Key: "recent_products", Value: bson.D{
			{Key: "$each", Value: bson.A{domain.RecentProduct{ProductID: interaction.ProductID, EventType: interaction.NormalizedEventType, OccurredAt: interaction.OccurredAt}}},
			{Key: "$position", Value: 0},
			{Key: "$slice", Value: batch.RecentProductsLimit},
		}}}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: profileKey},
			{Key: "profile_key", Value: profileKey},
			{Key: "embedding_refs", Value: bson.A{}},
			{Key: "feature_version", Value: domain.FeatureSchemaVersion},
			{Key: "created_at", Value: batch.ProcessedAt},
		}},
	}
	if len(inc) > 0 {
		update = append(update, bson.E{Key: "$inc", Value: inc})
	}
	if _, err := r.userFeatureProfiles.UpdateOne(ctx, bson.D{{Key: "profile_key", Value: profileKey}}, update, options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("upsert user feature profile: %w", err)
	}
	return nil
}

func (r *MongoFeatureRepository) upsertUserProductCounter(ctx context.Context, interaction domain.UserInteraction, batch domain.FeatureBatch) error {
	profileKey, err := domain.NewProfileKey(interaction.UserID, interaction.AnonymousID)
	if err != nil {
		return err
	}
	set := bson.D{
		{Key: "updated_at", Value: batch.ProcessedAt},
		{Key: "last_event_id", Value: interaction.EventID},
		{Key: "expires_at", Value: interaction.OccurredAt.Add(batch.UserProductRetention)},
	}
	appendOptionalSet(&set, "category_id", interaction.CategoryID)
	appendOptionalSet(&set, "seller_id", interaction.SellerID)
	if interaction.UserID != "" {
		set = append(set, bson.E{Key: "user_id", Value: interaction.UserID})
	} else {
		set = append(set, bson.E{Key: "anonymous_id", Value: interaction.AnonymousID})
	}
	update := bson.D{
		{Key: "$inc", Value: userProductIncrements(interaction)},
		{Key: "$set", Value: set},
		{Key: "$max", Value: bson.D{{Key: "last_interaction_at", Value: interaction.OccurredAt}}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: profileKey + ":" + interaction.ProductID},
			{Key: "profile_key", Value: profileKey},
			{Key: "product_id", Value: interaction.ProductID},
			{Key: "first_interaction_at", Value: interaction.OccurredAt},
			{Key: "feature_version", Value: domain.FeatureSchemaVersion},
			{Key: "created_at", Value: batch.ProcessedAt},
		}},
	}
	filter := bson.D{{Key: "profile_key", Value: profileKey}, {Key: "product_id", Value: interaction.ProductID}}
	if _, err := r.userProductCounters.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("upsert user-product counters: %w", err)
	}
	return nil
}

func (r *MongoFeatureRepository) upsertPurchasePairs(ctx context.Context, interactions []domain.UserInteraction, updatedAt time.Time) (int, error) {
	purchases := make(map[string]domain.UserInteraction)
	for _, interaction := range interactions {
		if interaction.NormalizedEventType == domain.InteractionPurchase {
			purchases[interaction.ProductID] = interaction
		}
	}
	products := make([]string, 0, len(purchases))
	for productID := range purchases {
		products = append(products, productID)
	}
	sort.Strings(products)
	count := 0
	for i := 0; i < len(products); i++ {
		for j := i + 1; j < len(products); j++ {
			pairKey, sourceID, relatedID, err := domain.NewProductPairKey(products[i], products[j])
			if err != nil {
				return count, err
			}
			source := purchases[sourceID]
			set := bson.D{{Key: "updated_at", Value: updatedAt}}
			appendOptionalSet(&set, "category_id", source.CategoryID)
			update := bson.D{
				{Key: "$inc", Value: bson.D{
					{Key: "counters.bought_together_all", Value: int64(1)},
					{Key: "score_input", Value: int64(8)},
				}},
				{Key: "$set", Value: set},
				{Key: "$max", Value: bson.D{{Key: "last_seen_at", Value: source.OccurredAt}}},
				{Key: "$setOnInsert", Value: bson.D{
					{Key: "_id", Value: pairKey},
					{Key: "source_product_id", Value: sourceID},
					{Key: "related_product_id", Value: relatedID},
					{Key: "relationship_type", Value: domain.RelationshipBoughtTogether},
					{Key: "counters.viewed_together_30d", Value: int64(0)},
					{Key: "counters.carted_together_30d", Value: int64(0)},
					{Key: "counters.bought_together_30d", Value: int64(0)},
					{Key: "feature_version", Value: domain.FeatureSchemaVersion},
					{Key: "created_at", Value: updatedAt},
				}},
			}
			if _, err := r.productCooccurrence.UpdateOne(ctx, bson.D{{Key: "_id", Value: pairKey}}, update, options.UpdateOne().SetUpsert(true)); err != nil {
				return count, fmt.Errorf("upsert product co-occurrence: %w", err)
			}
			count++
		}
	}
	return count, nil
}

func productLifetimeIncrements(interaction domain.UserInteraction) bson.D {
	quantity := int64(interaction.Quantity)
	switch interaction.NormalizedEventType {
	case domain.InteractionProductView:
		return bson.D{{Key: "counters.views_all", Value: quantity}}
	case domain.InteractionAddToCart:
		return bson.D{{Key: "counters.cart_adds_all", Value: quantity}}
	case domain.InteractionWishlistAdd:
		return bson.D{{Key: "counters.wishlist_adds_all", Value: quantity}}
	case domain.InteractionWishlistRemove:
		return bson.D{{Key: "counters.wishlist_removes_all", Value: quantity}}
	case domain.InteractionPurchase:
		return bson.D{{Key: "counters.purchases_all", Value: quantity}}
	default:
		return bson.D{}
	}
}

func userProductIncrements(interaction domain.UserInteraction) bson.D {
	inc := bson.D{{Key: "weighted_score_input", Value: interaction.WeightedFeatureScore()}}
	for _, value := range productLifetimeIncrements(interaction) {
		key := value.Key[len("counters."):]
		if key == "views_all" {
			key = "views"
		} else if key == "cart_adds_all" {
			key = "cart_adds"
		} else if key == "wishlist_adds_all" {
			key = "wishlist_adds"
		} else if key == "wishlist_removes_all" {
			key = "wishlist_removes"
		} else if key == "purchases_all" {
			key = "purchases"
		}
		inc = append(inc, bson.E{Key: "counters." + key, Value: value.Value})
	}
	return inc
}

func appendOptionalSet(set *bson.D, key string, value string) {
	if value != "" {
		*set = append(*set, bson.E{Key: key, Value: value})
	}
}

func appendScoreIncrement(inc *bson.D, bucket string, key string, score float64) {
	if key != "" {
		*inc = append(*inc, bson.E{Key: bucket + "." + key, Value: score})
	}
}

func latestInteractionTime(interactions []domain.UserInteraction) time.Time {
	latest := interactions[0].OccurredAt
	for _, interaction := range interactions[1:] {
		if interaction.OccurredAt.After(latest) {
			latest = interaction.OccurredAt
		}
	}
	return latest
}
