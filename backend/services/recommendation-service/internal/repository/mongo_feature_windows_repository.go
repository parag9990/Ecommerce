package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *MongoFeatureRepository) RebuildProductWindows(ctx context.Context, job domain.FeatureJobRun) (domain.FeatureJobStats, error) {
	if r == nil || r.db == nil {
		return domain.FeatureJobStats{}, fmt.Errorf("%w: mongo repository is not configured", domain.ErrRecommendationStorage)
	}
	if job.StartedAt.IsZero() || job.ID == "" {
		return domain.FeatureJobStats{}, fmt.Errorf("%w: window rebuild job identity is required", domain.ErrInvalidFeature)
	}
	now := job.StartedAt.UTC()
	stats := domain.FeatureJobStats{}
	session, err := r.db.Client().StartSession()
	if err != nil {
		return stats, fmt.Errorf("%w: start window rebuild transaction: %v", domain.ErrRecommendationStorage, err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		start30d := now.Add(-30 * 24 * time.Hour)
		count, err := r.userInteractions.CountDocuments(txCtx, bson.D{{Key: "occurred_at", Value: bson.D{{Key: "$gte", Value: start30d}}}})
		if err != nil {
			return nil, err
		}
		stats.EventsRead = int(count)
		if err := r.rebuildProductCounterWindows(txCtx, now, &stats); err != nil {
			return nil, err
		}
		pairs, err := r.rebuildCooccurrenceWindow(txCtx, now)
		if err != nil {
			return nil, err
		}
		stats.CooccurrencePairsUpdated = pairs
		job.Status = "success"
		job.FinishedAt = time.Now().UTC()
		job.Stats = stats
		job.CreatedAt = now
		_, err = r.featureJobRuns.InsertOne(txCtx, job)
		return nil, err
	}, options.Transaction())
	if err != nil {
		r.recordFailedFeatureRun(ctx, job, err)
		return domain.FeatureJobStats{}, fmt.Errorf("%w: rebuild feature windows: %v", domain.ErrRecommendationStorage, err)
	}
	return stats, nil
}

func (r *MongoFeatureRepository) rebuildProductCounterWindows(ctx context.Context, now time.Time, stats *domain.FeatureJobStats) error {
	zeros := bson.D{
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
	}
	if _, err := r.productFeatures.UpdateMany(ctx, bson.D{}, bson.D{{Key: "$set", Value: zeros}}); err != nil {
		return err
	}
	start24h := now.Add(-24 * time.Hour)
	start7d := now.Add(-7 * 24 * time.Hour)
	start30d := now.Add(-30 * 24 * time.Hour)
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "occurred_at", Value: bson.D{{Key: "$gte", Value: start30d}}}}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$product_id"},
			{Key: "views_24h", Value: windowSum(domain.InteractionProductView, start24h)},
			{Key: "views_7d", Value: windowSum(domain.InteractionProductView, start7d)},
			{Key: "views_30d", Value: windowSum(domain.InteractionProductView, start30d)},
			{Key: "cart_adds_24h", Value: windowSum(domain.InteractionAddToCart, start24h)},
			{Key: "cart_adds_7d", Value: windowSum(domain.InteractionAddToCart, start7d)},
			{Key: "cart_adds_30d", Value: windowSum(domain.InteractionAddToCart, start30d)},
			{Key: "wishlist_adds_24h", Value: windowSum(domain.InteractionWishlistAdd, start24h)},
			{Key: "wishlist_adds_7d", Value: windowSum(domain.InteractionWishlistAdd, start7d)},
			{Key: "wishlist_adds_30d", Value: windowSum(domain.InteractionWishlistAdd, start30d)},
			{Key: "purchases_24h", Value: windowSum(domain.InteractionPurchase, start24h)},
			{Key: "purchases_7d", Value: windowSum(domain.InteractionPurchase, start7d)},
			{Key: "purchases_30d", Value: windowSum(domain.InteractionPurchase, start30d)},
		}}},
	}
	cursor, err := r.userInteractions.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	type productWindows struct {
		ProductID       string `bson:"_id"`
		Views24h        int64  `bson:"views_24h"`
		Views7d         int64  `bson:"views_7d"`
		Views30d        int64  `bson:"views_30d"`
		CartAdds24h     int64  `bson:"cart_adds_24h"`
		CartAdds7d      int64  `bson:"cart_adds_7d"`
		CartAdds30d     int64  `bson:"cart_adds_30d"`
		WishlistAdds24h int64  `bson:"wishlist_adds_24h"`
		WishlistAdds7d  int64  `bson:"wishlist_adds_7d"`
		WishlistAdds30d int64  `bson:"wishlist_adds_30d"`
		Purchases24h    int64  `bson:"purchases_24h"`
		Purchases7d     int64  `bson:"purchases_7d"`
		Purchases30d    int64  `bson:"purchases_30d"`
	}
	models := make([]mongo.WriteModel, 0)
	for cursor.Next(ctx) {
		var windows productWindows
		if err := cursor.Decode(&windows); err != nil {
			return err
		}
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.D{{Key: "product_id", Value: windows.ProductID}}).
			SetUpdate(bson.D{{Key: "$set", Value: bson.D{
				{Key: "counters.views_24h", Value: windows.Views24h},
				{Key: "counters.views_7d", Value: windows.Views7d},
				{Key: "counters.views_30d", Value: windows.Views30d},
				{Key: "counters.cart_adds_24h", Value: windows.CartAdds24h},
				{Key: "counters.cart_adds_7d", Value: windows.CartAdds7d},
				{Key: "counters.cart_adds_30d", Value: windows.CartAdds30d},
				{Key: "counters.wishlist_adds_24h", Value: windows.WishlistAdds24h},
				{Key: "counters.wishlist_adds_7d", Value: windows.WishlistAdds7d},
				{Key: "counters.wishlist_adds_30d", Value: windows.WishlistAdds30d},
				{Key: "counters.purchases_24h", Value: windows.Purchases24h},
				{Key: "counters.purchases_7d", Value: windows.Purchases7d},
				{Key: "counters.purchases_30d", Value: windows.Purchases30d},
				{Key: "updated_at", Value: now},
			}}}))
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	if len(models) > 0 {
		if _, err := r.productFeatures.BulkWrite(ctx, models); err != nil {
			return err
		}
	}
	stats.ProductFeaturesUpdated = len(models)
	return nil
}

func (r *MongoFeatureRepository) rebuildCooccurrenceWindow(ctx context.Context, now time.Time) (int, error) {
	if _, err := r.productCooccurrence.UpdateMany(ctx, bson.D{}, bson.D{{Key: "$set", Value: bson.D{{Key: "counters.bought_together_30d", Value: int64(0)}}}}); err != nil {
		return 0, err
	}
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "normalized_event_type", Value: domain.InteractionPurchase},
			{Key: "occurred_at", Value: bson.D{{Key: "$gte", Value: now.Add(-30 * 24 * time.Hour)}}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$event_id"},
			{Key: "product_ids", Value: bson.D{{Key: "$addToSet", Value: "$product_id"}}},
		}}},
	}
	cursor, err := r.userInteractions.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)
	type orderProducts struct {
		ProductIDs []string `bson:"product_ids"`
	}
	increments := make(map[string]int64)
	for cursor.Next(ctx) {
		var order orderProducts
		if err := cursor.Decode(&order); err != nil {
			return 0, err
		}
		for i := 0; i < len(order.ProductIDs); i++ {
			for j := i + 1; j < len(order.ProductIDs); j++ {
				key, _, _, err := domain.NewProductPairKey(order.ProductIDs[i], order.ProductIDs[j])
				if err != nil {
					return 0, err
				}
				increments[key]++
			}
		}
	}
	if err := cursor.Err(); err != nil {
		return 0, err
	}
	models := make([]mongo.WriteModel, 0, len(increments))
	for key, count := range increments {
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.D{{Key: "_id", Value: key}}).
			SetUpdate(bson.D{{Key: "$set", Value: bson.D{{Key: "counters.bought_together_30d", Value: count}, {Key: "updated_at", Value: now}}}}))
	}
	if len(models) > 0 {
		if _, err := r.productCooccurrence.BulkWrite(ctx, models); err != nil {
			return 0, err
		}
	}
	return len(models), nil
}

func windowSum(eventType domain.InteractionType, start time.Time) bson.D {
	return bson.D{{Key: "$sum", Value: bson.D{{Key: "$cond", Value: bson.A{
		bson.D{{Key: "$and", Value: bson.A{
			bson.D{{Key: "$eq", Value: bson.A{"$normalized_event_type", eventType}}},
			bson.D{{Key: "$gte", Value: bson.A{"$occurred_at", start}}},
		}}},
		"$quantity",
		int64(0),
	}}}}}
}
