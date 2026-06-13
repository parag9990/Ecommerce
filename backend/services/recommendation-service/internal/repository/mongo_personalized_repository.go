package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *MongoFeatureRepository) GetUserFeatureProfile(ctx context.Context, profileKey string) (domain.UserFeatureProfile, error) {
	profileKey = strings.TrimSpace(profileKey)
	if profileKey == "" {
		return domain.UserFeatureProfile{}, fmt.Errorf("%w: profile_key is required", domain.ErrInvalidFeature)
	}
	if r == nil || r.userFeatureProfiles == nil {
		return domain.UserFeatureProfile{}, fmt.Errorf("%w: mongo repository is not configured", domain.ErrRecommendationStorage)
	}

	var profile domain.UserFeatureProfile
	if err := r.userFeatureProfiles.FindOne(ctx, bson.D{{Key: "profile_key", Value: profileKey}}).Decode(&profile); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.UserFeatureProfile{}, fmt.Errorf("%w: profile_key=%s", domain.ErrFeatureProfileNotFound, profileKey)
		}
		return domain.UserFeatureProfile{}, fmt.Errorf("%w: get user feature profile: %v", domain.ErrRecommendationStorage, err)
	}
	return profile, nil
}

func (r *MongoFeatureRepository) ListUserProductCounters(ctx context.Context, profileKey string, limit int) ([]domain.UserProductCounter, error) {
	profileKey = strings.TrimSpace(profileKey)
	if profileKey == "" {
		return nil, fmt.Errorf("%w: profile_key is required", domain.ErrInvalidFeature)
	}
	if r == nil || r.userProductCounters == nil {
		return nil, fmt.Errorf("%w: mongo repository is not configured", domain.ErrRecommendationStorage)
	}
	if limit <= 0 {
		limit = domain.DefaultPersonalizedDirectProductLimit
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "last_interaction_at", Value: -1}, {Key: "product_id", Value: 1}}).
		SetLimit(int64(limit))
	cursor, err := r.userProductCounters.Find(ctx, bson.D{{Key: "profile_key", Value: profileKey}}, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: list user product counters: %v", domain.ErrRecommendationStorage, err)
	}
	defer cursor.Close(ctx)

	var counters []domain.UserProductCounter
	if err := cursor.All(ctx, &counters); err != nil {
		return nil, fmt.Errorf("%w: decode user product counters: %v", domain.ErrRecommendationStorage, err)
	}
	return counters, nil
}

func (r *MongoFeatureRepository) ListPersonalizedCandidates(ctx context.Context, query domain.PersonalizedCandidateQuery) ([]domain.ProductFeature, error) {
	if r == nil || r.productFeatures == nil {
		return nil, fmt.Errorf("%w: mongo repository is not configured", domain.ErrRecommendationStorage)
	}
	query = query.Normalize()
	if query.Empty() {
		return nil, nil
	}

	orConditions := bson.A{}
	if len(query.CategoryIDs) > 0 {
		orConditions = append(orConditions, bson.D{{Key: "category_id", Value: bson.D{{Key: "$in", Value: query.CategoryIDs}}}})
	}
	if len(query.SellerIDs) > 0 {
		orConditions = append(orConditions, bson.D{{Key: "seller_id", Value: bson.D{{Key: "$in", Value: query.SellerIDs}}}})
	}
	if len(query.BrandIDs) > 0 {
		orConditions = append(orConditions, bson.D{{Key: "brand_id", Value: bson.D{{Key: "$in", Value: query.BrandIDs}}}})
	}
	if len(query.ProductIDs) > 0 {
		orConditions = append(orConditions, bson.D{{Key: "product_id", Value: bson.D{{Key: "$in", Value: query.ProductIDs}}}})
	}

	filter := bson.D{
		{Key: "$or", Value: orConditions},
		{Key: "product_id", Value: bson.D{{Key: "$type", Value: "string"}, {Key: "$ne", Value: ""}}},
		{Key: "status", Value: domain.ProductStatusActive},
		{Key: "stock_status", Value: domain.ProductStockStatusInStock},
		{Key: "quality_flags.is_recommendable", Value: true},
		{Key: "quality_flags.is_deleted", Value: bson.D{{Key: "$ne", Value: true}}},
	}
	opts := options.Find().
		SetSort(bson.D{
			{Key: "counters.purchases_7d", Value: -1},
			{Key: "counters.cart_adds_7d", Value: -1},
			{Key: "counters.wishlist_adds_7d", Value: -1},
			{Key: "product_id", Value: 1},
		}).
		SetLimit(int64(query.Limit))

	cursor, err := r.productFeatures.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: list personalized candidates: %v", domain.ErrRecommendationStorage, err)
	}
	defer cursor.Close(ctx)

	var features []domain.ProductFeature
	if err := cursor.All(ctx, &features); err != nil {
		return nil, fmt.Errorf("%w: decode personalized candidates: %v", domain.ErrRecommendationStorage, err)
	}
	return features, nil
}
