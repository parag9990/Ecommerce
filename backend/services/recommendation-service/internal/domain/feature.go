package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	FeatureSchemaVersion        = 1
	DefaultRecentProductsLimit  = 20
	FeatureJobTypeIncremental   = "incremental_interaction_features"
	FeatureJobTypeWindowRebuild = "product_window_rebuild"
	RelationshipBoughtTogether  = "bought_together"
)

type EmbeddingStatus string

const (
	EmbeddingPending EmbeddingStatus = "pending"
	EmbeddingReady   EmbeddingStatus = "ready"
	EmbeddingFailed  EmbeddingStatus = "failed"
	EmbeddingExpired EmbeddingStatus = "expired"
)

type EmbeddingRef struct {
	Type        string          `json:"type" bson:"type"`
	VectorID    string          `json:"vector_id" bson:"vector_id"`
	Provider    string          `json:"provider" bson:"provider"`
	Model       string          `json:"model" bson:"model"`
	Dimension   int             `json:"dimension" bson:"dimension"`
	Version     int             `json:"version" bson:"version"`
	Status      EmbeddingStatus `json:"status" bson:"status"`
	GeneratedAt *time.Time      `json:"generated_at,omitempty" bson:"generated_at,omitempty"`
	ExpiresAt   *time.Time      `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
}

func (r EmbeddingRef) Validate() error {
	if strings.TrimSpace(r.Type) == "" || strings.TrimSpace(r.VectorID) == "" {
		return fmt.Errorf("%w: embedding type and vector_id are required", ErrInvalidFeature)
	}
	if r.Dimension <= 0 || r.Version <= 0 {
		return fmt.Errorf("%w: embedding dimension and version must be greater than zero", ErrInvalidFeature)
	}
	switch r.Status {
	case EmbeddingPending, EmbeddingReady, EmbeddingFailed, EmbeddingExpired:
		return nil
	default:
		return fmt.Errorf("%w: invalid embedding status", ErrInvalidFeature)
	}
}

type ProductPrice struct {
	Amount   int64  `json:"amount" bson:"amount"`
	Currency string `json:"currency" bson:"currency"`
	Bucket   string `json:"bucket" bson:"bucket"`
}

type ProductQualityFlags struct {
	IsRecommendable bool `json:"is_recommendable" bson:"is_recommendable"`
	IsAdult         bool `json:"is_adult" bson:"is_adult"`
	IsDeleted       bool `json:"is_deleted" bson:"is_deleted"`
}

type ProductCounters struct {
	Views24h           int64 `json:"views_24h" bson:"views_24h"`
	Views7d            int64 `json:"views_7d" bson:"views_7d"`
	Views30d           int64 `json:"views_30d" bson:"views_30d"`
	ViewsAll           int64 `json:"views_all" bson:"views_all"`
	CartAdds24h        int64 `json:"cart_adds_24h" bson:"cart_adds_24h"`
	CartAdds7d         int64 `json:"cart_adds_7d" bson:"cart_adds_7d"`
	CartAdds30d        int64 `json:"cart_adds_30d" bson:"cart_adds_30d"`
	CartAddsAll        int64 `json:"cart_adds_all" bson:"cart_adds_all"`
	WishlistAdds24h    int64 `json:"wishlist_adds_24h" bson:"wishlist_adds_24h"`
	WishlistAdds7d     int64 `json:"wishlist_adds_7d" bson:"wishlist_adds_7d"`
	WishlistAdds30d    int64 `json:"wishlist_adds_30d" bson:"wishlist_adds_30d"`
	WishlistAddsAll    int64 `json:"wishlist_adds_all" bson:"wishlist_adds_all"`
	WishlistRemovesAll int64 `json:"wishlist_removes_all" bson:"wishlist_removes_all"`
	Purchases24h       int64 `json:"purchases_24h" bson:"purchases_24h"`
	Purchases7d        int64 `json:"purchases_7d" bson:"purchases_7d"`
	Purchases30d       int64 `json:"purchases_30d" bson:"purchases_30d"`
	PurchasesAll       int64 `json:"purchases_all" bson:"purchases_all"`
}

type ProductFeature struct {
	ID             string              `json:"id" bson:"_id"`
	ProductID      string              `json:"product_id" bson:"product_id"`
	CategoryID     string              `json:"category_id,omitempty" bson:"category_id,omitempty"`
	SellerID       string              `json:"seller_id,omitempty" bson:"seller_id,omitempty"`
	BrandID        string              `json:"brand_id,omitempty" bson:"brand_id,omitempty"`
	Status         string              `json:"status,omitempty" bson:"status,omitempty"`
	StockStatus    string              `json:"stock_status,omitempty" bson:"stock_status,omitempty"`
	Price          *ProductPrice       `json:"price,omitempty" bson:"price,omitempty"`
	Attributes     map[string]string   `json:"attributes,omitempty" bson:"attributes,omitempty"`
	Counters       ProductCounters     `json:"counters" bson:"counters"`
	QualityFlags   ProductQualityFlags `json:"quality_flags" bson:"quality_flags"`
	EmbeddingRefs  []EmbeddingRef      `json:"embedding_refs,omitempty" bson:"embedding_refs,omitempty"`
	FeatureVersion int                 `json:"feature_version" bson:"feature_version"`
	LastEventAt    time.Time           `json:"last_event_at" bson:"last_event_at"`
	CreatedAt      time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at" bson:"updated_at"`
}

type RecentProduct struct {
	ProductID  string          `json:"product_id" bson:"product_id"`
	EventType  InteractionType `json:"event_type" bson:"event_type"`
	OccurredAt time.Time       `json:"occurred_at" bson:"occurred_at"`
}

type PriceAffinity struct {
	Currency         string   `json:"currency,omitempty" bson:"currency,omitempty"`
	PreferredBuckets []string `json:"preferred_buckets,omitempty" bson:"preferred_buckets,omitempty"`
	MinSeen          int64    `json:"min_seen,omitempty" bson:"min_seen,omitempty"`
	MaxSeen          int64    `json:"max_seen,omitempty" bson:"max_seen,omitempty"`
}

type UserFeatureProfile struct {
	ID             string             `json:"id" bson:"_id"`
	ProfileKey     string             `json:"profile_key" bson:"profile_key"`
	UserID         string             `json:"user_id,omitempty" bson:"user_id,omitempty"`
	AnonymousID    string             `json:"anonymous_id,omitempty" bson:"anonymous_id,omitempty"`
	CategoryScores map[string]float64 `json:"category_scores,omitempty" bson:"category_scores,omitempty"`
	SellerScores   map[string]float64 `json:"seller_scores,omitempty" bson:"seller_scores,omitempty"`
	BrandScores    map[string]float64 `json:"brand_scores,omitempty" bson:"brand_scores,omitempty"`
	PriceAffinity  *PriceAffinity     `json:"price_affinity,omitempty" bson:"price_affinity,omitempty"`
	RecentProducts []RecentProduct    `json:"recent_products,omitempty" bson:"recent_products,omitempty"`
	EmbeddingRefs  []EmbeddingRef     `json:"embedding_refs,omitempty" bson:"embedding_refs,omitempty"`
	LastEventAt    time.Time          `json:"last_event_at" bson:"last_event_at"`
	ExpiresAt      *time.Time         `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	FeatureVersion int                `json:"feature_version" bson:"feature_version"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}

type InteractionCounters struct {
	Views           int64 `json:"views" bson:"views"`
	CartAdds        int64 `json:"cart_adds" bson:"cart_adds"`
	WishlistAdds    int64 `json:"wishlist_adds" bson:"wishlist_adds"`
	WishlistRemoves int64 `json:"wishlist_removes" bson:"wishlist_removes"`
	Purchases       int64 `json:"purchases" bson:"purchases"`
}

type UserProductCounter struct {
	ID                 string              `json:"id" bson:"_id"`
	ProfileKey         string              `json:"profile_key" bson:"profile_key"`
	UserID             string              `json:"user_id,omitempty" bson:"user_id,omitempty"`
	AnonymousID        string              `json:"anonymous_id,omitempty" bson:"anonymous_id,omitempty"`
	ProductID          string              `json:"product_id" bson:"product_id"`
	CategoryID         string              `json:"category_id,omitempty" bson:"category_id,omitempty"`
	SellerID           string              `json:"seller_id,omitempty" bson:"seller_id,omitempty"`
	Counters           InteractionCounters `json:"counters" bson:"counters"`
	WeightedScoreInput int64               `json:"weighted_score_input" bson:"weighted_score_input"`
	FirstInteractionAt time.Time           `json:"first_interaction_at" bson:"first_interaction_at"`
	LastInteractionAt  time.Time           `json:"last_interaction_at" bson:"last_interaction_at"`
	LastEventID        string              `json:"last_event_id" bson:"last_event_id"`
	ExpiresAt          time.Time           `json:"expires_at" bson:"expires_at"`
	FeatureVersion     int                 `json:"feature_version" bson:"feature_version"`
	CreatedAt          time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at" bson:"updated_at"`
}

type CooccurrenceCounters struct {
	ViewedTogether30d int64 `json:"viewed_together_30d" bson:"viewed_together_30d"`
	CartedTogether30d int64 `json:"carted_together_30d" bson:"carted_together_30d"`
	BoughtTogether30d int64 `json:"bought_together_30d" bson:"bought_together_30d"`
	BoughtTogetherAll int64 `json:"bought_together_all" bson:"bought_together_all"`
}

type ProductCooccurrenceFeature struct {
	ID               string               `json:"id" bson:"_id"`
	SourceProductID  string               `json:"source_product_id" bson:"source_product_id"`
	RelatedProductID string               `json:"related_product_id" bson:"related_product_id"`
	CategoryID       string               `json:"category_id,omitempty" bson:"category_id,omitempty"`
	RelationshipType string               `json:"relationship_type" bson:"relationship_type"`
	Counters         CooccurrenceCounters `json:"counters" bson:"counters"`
	ScoreInput       int64                `json:"score_input" bson:"score_input"`
	LastSeenAt       time.Time            `json:"last_seen_at" bson:"last_seen_at"`
	FeatureVersion   int                  `json:"feature_version" bson:"feature_version"`
	CreatedAt        time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at" bson:"updated_at"`
}

type FeatureWatermark struct {
	LastOccurredAt time.Time `json:"last_occurred_at" bson:"last_occurred_at"`
	LastEventID    string    `json:"last_event_id" bson:"last_event_id"`
}

type FeatureJobStats struct {
	EventsRead                 int `json:"events_read" bson:"events_read"`
	ProductFeaturesUpdated     int `json:"product_features_updated" bson:"product_features_updated"`
	UserProfilesUpdated        int `json:"user_profiles_updated" bson:"user_profiles_updated"`
	UserProductCountersUpdated int `json:"user_product_counters_updated" bson:"user_product_counters_updated"`
	CooccurrencePairsUpdated   int `json:"cooccurrence_pairs_updated" bson:"cooccurrence_pairs_updated"`
}

type FeatureJobRun struct {
	ID         string           `json:"id" bson:"_id"`
	JobType    string           `json:"job_type" bson:"job_type"`
	Status     string           `json:"status" bson:"status"`
	StartedAt  time.Time        `json:"started_at" bson:"started_at"`
	FinishedAt time.Time        `json:"finished_at" bson:"finished_at"`
	Watermark  FeatureWatermark `json:"watermark" bson:"watermark"`
	Stats      FeatureJobStats  `json:"stats" bson:"stats"`
	Error      string           `json:"error,omitempty" bson:"error,omitempty"`
	CreatedAt  time.Time        `json:"created_at" bson:"created_at"`
}

type FeatureBatch struct {
	ID                      string
	Interactions            []UserInteraction
	StartedAt               time.Time
	ProcessedAt             time.Time
	GuestProfileRetention   time.Duration
	UserProductRetention    time.Duration
	ProcessedEventRetention time.Duration
	RecentProductsLimit     int
}

type FeatureBuildResult struct {
	Applied    int
	Duplicates int
	Stats      FeatureJobStats
}

func NewProfileKey(userID string, anonymousID string) (string, error) {
	if userID = strings.TrimSpace(userID); userID != "" {
		return "user:" + userID, nil
	}
	if anonymousID = strings.TrimSpace(anonymousID); anonymousID != "" {
		return "anon:" + anonymousID, nil
	}
	return "", fmt.Errorf("%w: user_id or anonymous_id is required for profile key", ErrInvalidFeature)
}

func NewProductPairKey(productA string, productB string) (string, string, string, error) {
	productA = strings.TrimSpace(productA)
	productB = strings.TrimSpace(productB)
	if productA == "" || productB == "" || productA == productB {
		return "", "", "", fmt.Errorf("%w: two different products are required for a pair", ErrInvalidFeature)
	}
	products := []string{productA, productB}
	sort.Strings(products)
	return products[0] + ":" + products[1], products[0], products[1], nil
}

func NewFeatureJobID(jobType string, interactions []UserInteraction, at time.Time) string {
	parts := []string{jobType, at.UTC().Format(time.RFC3339Nano)}
	for _, interaction := range interactions {
		parts = append(parts, interaction.DedupeKey)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "#")))
	return "feature_job_" + hex.EncodeToString(sum[:16])
}

func (b FeatureBatch) Validate() error {
	if strings.TrimSpace(b.ID) == "" || len(b.Interactions) == 0 {
		return fmt.Errorf("%w: feature batch id and interactions are required", ErrInvalidFeature)
	}
	if b.StartedAt.IsZero() || b.ProcessedAt.IsZero() {
		return fmt.Errorf("%w: feature batch timestamps are required", ErrInvalidFeature)
	}
	if b.GuestProfileRetention <= 0 || b.UserProductRetention <= 0 || b.ProcessedEventRetention <= 0 {
		return fmt.Errorf("%w: feature retention values must be greater than zero", ErrInvalidFeature)
	}
	if b.RecentProductsLimit <= 0 {
		return fmt.Errorf("%w: recent product limit must be greater than zero", ErrInvalidFeature)
	}
	eventID := b.Interactions[0].EventID
	for _, interaction := range b.Interactions {
		if err := interaction.Validate(); err != nil {
			return err
		}
		if interaction.EventID != eventID {
			return fmt.Errorf("%w: a feature batch must contain one source event", ErrInvalidFeature)
		}
		if _, err := NewProfileKey(interaction.UserID, interaction.AnonymousID); err != nil {
			return err
		}
		for field, value := range map[string]string{
			"category_id": interaction.CategoryID,
			"seller_id":   interaction.SellerID,
			"brand_id":    interaction.BrandID,
		} {
			if strings.Contains(value, ".") || strings.HasPrefix(value, "$") {
				return fmt.Errorf("%w: %s cannot be used as a feature map key", ErrInvalidFeature, field)
			}
		}
	}
	return nil
}

func (i UserInteraction) WeightedFeatureScore() int64 {
	return int64(i.Weight * i.Quantity)
}
