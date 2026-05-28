package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	RecommendationDatabaseName = "recommendation_db"

	CollectionUserInteractions       = "user_interactions"
	CollectionProductFeatures        = "product_features"
	CollectionUserFeatureProfiles    = "user_feature_profiles"
	CollectionUserProductCounters    = "user_product_counters"
	CollectionProductCooccurrence    = "product_cooccurrence_features"
	CollectionFeatureJobRuns         = "feature_job_runs"
	CollectionFeatureProcessedEvents = "feature_processed_events"
	CollectionRecommendationSets     = "recommendation_sets"
	CollectionABTestAssignments      = "ab_test_assignments"

	DefaultCacheKeyPrefix         = "reco:v1"
	DefaultRecommendationCacheTTL = 15 * time.Minute
	DefaultPersonalizedCacheTTL   = 5 * time.Minute
	DefaultGuestCacheTTL          = 5 * time.Minute
	DefaultInteractionRetention   = 180 * 24 * time.Hour
	DefaultGuestProfileRetention  = 30 * 24 * time.Hour
	DefaultUserProductRetention   = 180 * 24 * time.Hour
	DefaultFeatureEventRetention  = DefaultInteractionRetention
	DefaultABAssignmentTTL        = 90 * 24 * time.Hour
	DefaultABTestSalt             = "reco-ab-v1"
)

type StorageLayer string

const (
	StorageLayerMongoDB StorageLayer = "mongodb"
	StorageLayerRedis   StorageLayer = "redis"
)

type CacheValueType string

const (
	CacheValueTypeString     CacheValueType = "string"
	CacheValueTypeJSONString CacheValueType = "json_string"
	CacheValueTypeSortedSet  CacheValueType = "sorted_set"
	CacheValueTypeLock       CacheValueType = "set_nx_ex"
)

type RecommendationSetItem struct {
	ProductID string  `json:"product_id" bson:"product_id"`
	Score     float64 `json:"score" bson:"score"`
	Rank      int     `json:"rank" bson:"rank"`
	Reason    string  `json:"reason,omitempty" bson:"reason,omitempty"`
}

type RecommendationSet struct {
	ID          string                  `json:"id" bson:"_id"`
	ContextKey  string                  `json:"context_key" bson:"context_key"`
	Type        RecommendationType      `json:"recommendation_type" bson:"recommendation_type"`
	StrategyID  StrategyID              `json:"strategy_id" bson:"strategy_id"`
	Items       []RecommendationSetItem `json:"items" bson:"items"`
	Metadata    map[string]string       `json:"metadata,omitempty" bson:"metadata,omitempty"`
	GeneratedAt time.Time               `json:"generated_at" bson:"generated_at"`
	ExpiresAt   time.Time               `json:"expires_at" bson:"expires_at"`
}

type CachedRecommendation struct {
	RecommendationID string                  `json:"recommendation_id"`
	ContextKey       string                  `json:"context_key"`
	Type             RecommendationType      `json:"recommendation_type"`
	StrategyID       StrategyID              `json:"strategy_id"`
	Items            []RecommendationSetItem `json:"items"`
	CachedAt         time.Time               `json:"cached_at"`
	ExpiresAt        time.Time               `json:"expires_at"`
}

type MongoCollectionPlan struct {
	Name        string
	Purpose     string
	Indexes     []MongoIndexPlan
	DetailLevel string
}

type MongoIndexPlan struct {
	Name               string
	Keys               []string
	Unique             bool
	ExpireAfterSeconds int64
	Purpose            string
}

type RedisCachePlan struct {
	Name       string
	KeyPattern string
	ValueType  CacheValueType
	TTL        time.Duration
	Purpose    string
}

type CacheTTLPolicy struct {
	Default      time.Duration
	Personalized time.Duration
	Guest        time.Duration
	RebuildLock  time.Duration
}

type FailureBehavior struct {
	Failure  string
	Behavior string
}

type StoragePlan struct {
	DatabaseName     string
	CacheKeyPrefix   string
	CacheKeyPattern  string
	OwnershipRule    string
	MongoCollections []MongoCollectionPlan
	RedisCaches      []RedisCachePlan
	FailureBehavior  []FailureBehavior
	TTLPolicy        CacheTTLPolicy
}

type CacheKeyBuilder struct {
	prefix              string
	maxIdentifierLength int
}

var cachePartPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

func NewCacheKeyBuilder(prefix string, maxIdentifierLength int) (CacheKeyBuilder, error) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = DefaultCacheKeyPrefix
	}
	if maxIdentifierLength <= 0 {
		maxIdentifierLength = 128
	}
	if err := ValidateCacheKeyPrefix(prefix); err != nil {
		return CacheKeyBuilder{}, err
	}
	return CacheKeyBuilder{prefix: prefix, maxIdentifierLength: maxIdentifierLength}, nil
}

func ValidateCacheKeyPrefix(prefix string) error {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return fmt.Errorf("%w: cache key prefix is required", ErrInvalidCacheKey)
	}
	if strings.HasPrefix(prefix, ":") || strings.HasSuffix(prefix, ":") || strings.Contains(prefix, "::") {
		return fmt.Errorf("%w: cache key prefix must not have empty segments", ErrInvalidCacheKey)
	}
	parts := strings.Split(prefix, ":")
	for _, part := range parts {
		if !cachePartPattern.MatchString(part) {
			return fmt.Errorf("%w: invalid cache key prefix segment %q", ErrInvalidCacheKey, part)
		}
	}
	return nil
}

func (b CacheKeyBuilder) Prefix() string {
	return b.prefix
}

func (b CacheKeyBuilder) Pattern() string {
	return b.prefix + ":{type}:{scope}:{id}:strategy:{strategy_id}"
}

func (b CacheKeyBuilder) TrendingGlobalKey() string {
	return b.join("trending", "global")
}

func (b CacheKeyBuilder) TrendingCategoryKey(categoryID string) (string, error) {
	if err := b.validateIdentifier("category_id", categoryID); err != nil {
		return "", err
	}
	return b.join("trending", "category", strings.TrimSpace(categoryID)), nil
}

func (b CacheKeyBuilder) TrendingSellerKey(sellerID string) (string, error) {
	if err := b.validateIdentifier("seller_id", sellerID); err != nil {
		return "", err
	}
	return b.join("trending", "seller", strings.TrimSpace(sellerID)), nil
}

func (b CacheKeyBuilder) SimilarProductKey(productID string) (string, error) {
	if err := b.validateIdentifier("product_id", productID); err != nil {
		return "", err
	}
	return b.join("similar", "product", strings.TrimSpace(productID)), nil
}

func (b CacheKeyBuilder) PersonalizedUserKey(userID string) (string, error) {
	if err := b.validateIdentifier("user_id", userID); err != nil {
		return "", err
	}
	return b.join("personalized", "user", strings.TrimSpace(userID)), nil
}

func (b CacheKeyBuilder) PersonalizedAnonymousKey(anonymousID string) (string, error) {
	if err := b.validateIdentifier("anonymous_id", anonymousID); err != nil {
		return "", err
	}
	return b.join("personalized", "anon", strings.TrimSpace(anonymousID)), nil
}

func (b CacheKeyBuilder) FrequentlyBoughtTogetherProductKey(productID string) (string, error) {
	if err := b.validateIdentifier("product_id", productID); err != nil {
		return "", err
	}
	return b.join("fbt", "product", strings.TrimSpace(productID)), nil
}

func (b CacheKeyBuilder) StrategyScopedKey(baseKey string, strategyID StrategyID) (string, error) {
	baseKey = strings.TrimSpace(baseKey)
	if baseKey == "" {
		return "", fmt.Errorf("%w: base cache key is required", ErrInvalidCacheKey)
	}
	if strings.ContainsAny(baseKey, " \t\n\r") {
		return "", fmt.Errorf("%w: base cache key must not contain whitespace", ErrInvalidCacheKey)
	}
	if err := ValidateStrategyID(strategyID); err != nil {
		return "", err
	}
	return baseKey + ":strategy:" + string(strategyID), nil
}

func (b CacheKeyBuilder) DirtyProductKey(productID string) (string, error) {
	if err := b.validateIdentifier("product_id", productID); err != nil {
		return "", err
	}
	return b.join("dirty", "product", strings.TrimSpace(productID)), nil
}

func (b CacheKeyBuilder) DirtyCategoryKey(categoryID string) (string, error) {
	if err := b.validateIdentifier("category_id", categoryID); err != nil {
		return "", err
	}
	return b.join("dirty", "category", strings.TrimSpace(categoryID)), nil
}

func (b CacheKeyBuilder) DirtySellerKey(sellerID string) (string, error) {
	if err := b.validateIdentifier("seller_id", sellerID); err != nil {
		return "", err
	}
	return b.join("dirty", "seller", strings.TrimSpace(sellerID)), nil
}

func (b CacheKeyBuilder) RebuildLockKey(cacheKey string) (string, error) {
	cacheKey = strings.TrimSpace(cacheKey)
	if cacheKey == "" {
		return "", fmt.Errorf("%w: cache key is required", ErrInvalidCacheKey)
	}
	if strings.ContainsAny(cacheKey, " \t\n\r") {
		return "", fmt.Errorf("%w: cache key must not contain whitespace", ErrInvalidCacheKey)
	}
	return b.join("lock", cacheKey), nil
}

func (b CacheKeyBuilder) validateIdentifier(field string, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%w: %s is required", ErrInvalidCacheKey, field)
	}
	if len(value) > b.maxIdentifierLength {
		return fmt.Errorf("%w: %s is too long", ErrInvalidCacheKey, field)
	}
	if !cachePartPattern.MatchString(value) {
		return fmt.Errorf("%w: %s contains unsupported characters", ErrInvalidCacheKey, field)
	}
	return nil
}

func (b CacheKeyBuilder) join(parts ...string) string {
	all := make([]string, 0, len(parts)+1)
	all = append(all, b.prefix)
	all = append(all, parts...)
	return strings.Join(all, ":")
}

func (p CacheTTLPolicy) TTLFor(typ RecommendationType, anonymous bool) time.Duration {
	if typ == RecommendationTypePersonalized && anonymous && p.Guest > 0 {
		return p.Guest
	}
	if typ == RecommendationTypePersonalized && p.Personalized > 0 {
		return p.Personalized
	}
	if p.Default > 0 {
		return p.Default
	}
	return DefaultRecommendationCacheTTL
}

func (s RecommendationSet) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return fmt.Errorf("%w: recommendation set id is required", ErrInvalidRecommendationSet)
	}
	if strings.TrimSpace(s.ContextKey) == "" {
		return fmt.Errorf("%w: context_key is required", ErrInvalidRecommendationSet)
	}
	if !s.Type.IsValid() {
		return fmt.Errorf("%w: invalid recommendation_type", ErrInvalidRecommendationSet)
	}
	if err := ValidateStrategyID(s.StrategyID); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRecommendationSet, err)
	}
	if s.GeneratedAt.IsZero() {
		return fmt.Errorf("%w: generated_at is required", ErrInvalidRecommendationSet)
	}
	if s.ExpiresAt.IsZero() || !s.ExpiresAt.After(s.GeneratedAt) {
		return fmt.Errorf("%w: expires_at must be after generated_at", ErrInvalidRecommendationSet)
	}
	for i, item := range s.Items {
		if strings.TrimSpace(item.ProductID) == "" {
			return fmt.Errorf("%w: item %d product_id is required", ErrInvalidRecommendationSet, i)
		}
		if item.Rank <= 0 {
			return fmt.Errorf("%w: item %d rank must be greater than zero", ErrInvalidRecommendationSet, i)
		}
	}
	return nil
}

func (s RecommendationSet) IsExpired(now time.Time) bool {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return !s.ExpiresAt.After(now)
}

func (c CachedRecommendation) Validate() error {
	if strings.TrimSpace(c.RecommendationID) == "" {
		return fmt.Errorf("%w: recommendation_id is required", ErrInvalidRecommendationSet)
	}
	set := RecommendationSet{
		ID:          c.RecommendationID,
		ContextKey:  c.ContextKey,
		Type:        c.Type,
		StrategyID:  c.StrategyID,
		Items:       c.Items,
		GeneratedAt: c.CachedAt,
		ExpiresAt:   c.ExpiresAt,
	}
	return set.Validate()
}

func BuildStoragePlan(databaseName string, cacheKeyPrefix string, ttl CacheTTLPolicy, interactionRetention time.Duration) StoragePlan {
	if databaseName == "" {
		databaseName = RecommendationDatabaseName
	}
	if cacheKeyPrefix == "" {
		cacheKeyPrefix = DefaultCacheKeyPrefix
	}
	if ttl.Default <= 0 {
		ttl.Default = DefaultRecommendationCacheTTL
	}
	if ttl.Personalized <= 0 {
		ttl.Personalized = DefaultPersonalizedCacheTTL
	}
	if ttl.Guest <= 0 {
		ttl.Guest = DefaultGuestCacheTTL
	}
	if ttl.RebuildLock <= 0 {
		ttl.RebuildLock = time.Minute
	}
	if interactionRetention <= 0 {
		interactionRetention = DefaultInteractionRetention
	}
	interactionTTLSeconds := int64(interactionRetention.Seconds())
	builder, _ := NewCacheKeyBuilder(cacheKeyPrefix, 128)

	return StoragePlan{
		DatabaseName:    databaseName,
		CacheKeyPrefix:  cacheKeyPrefix,
		CacheKeyPattern: builder.Pattern(),
		OwnershipRule:   "Recommendation Service owns recommendation_db and does not directly read other service databases.",
		MongoCollections: []MongoCollectionPlan{
			{
				Name:        CollectionUserInteractions,
				Purpose:     "Durable product view, click, cart, wishlist, and purchase interaction history.",
				DetailLevel: "raw event ingestion collection with idempotent dedupe keys",
				Indexes: []MongoIndexPlan{
					{Name: "ux_user_interactions_dedupe_key", Keys: []string{"dedupe_key:1"}, Unique: true, Purpose: "Treat duplicate at-least-once queue deliveries as successful no-ops."},
					{Name: "idx_user_interactions_user_recent", Keys: []string{"user_id:1", "occurred_at:-1"}, Purpose: "Known-user recent behavior reads."},
					{Name: "idx_user_interactions_anon_recent", Keys: []string{"anonymous_id:1", "occurred_at:-1"}, Purpose: "Guest recent behavior reads."},
					{Name: "idx_user_interactions_product_normalized_event_recent", Keys: []string{"product_id:1", "normalized_event_type:1", "occurred_at:-1"}, Purpose: "Product popularity and co-occurrence calculations."},
					{Name: "idx_user_interactions_category_normalized_event_recent", Keys: []string{"category_id:1", "normalized_event_type:1", "occurred_at:-1"}, Purpose: "Category-level trending calculations."},
					{Name: "idx_user_interactions_ttl", Keys: []string{"occurred_at:1"}, ExpireAfterSeconds: interactionTTLSeconds, Purpose: "Raw interaction retention cleanup."},
				},
			},
			{
				Name:        CollectionRecommendationSets,
				Purpose:     "Durable precomputed recommendation lists per context and strategy.",
				DetailLevel: "basic document shape and indexes",
				Indexes: []MongoIndexPlan{
					{Name: "ux_recommendation_sets_context_strategy", Keys: []string{"context_key:1", "strategy_id:1"}, Unique: true, Purpose: "Prevent duplicate generated sets for a context/strategy pair."},
					{Name: "idx_recommendation_sets_expires_at_ttl", Keys: []string{"expires_at:1"}, ExpireAfterSeconds: 0, Purpose: "Remove expired generated lists."},
					{Name: "idx_recommendation_sets_type_generated", Keys: []string{"recommendation_type:1", "generated_at:-1"}, Purpose: "Inspect latest sets by recommendation type."},
				},
			},
			{
				Name:        CollectionProductFeatures,
				Purpose:     "Product-level popularity, catalog projection fields, quality flags, and embedding references.",
				DetailLevel: "feature store product snapshot with lifetime and periodically rebuilt window counters",
				Indexes: []MongoIndexPlan{
					{Name: "ux_product_features_product_id", Keys: []string{"product_id:1"}, Unique: true, Purpose: "Stable product projection lookup."},
					{Name: "idx_product_features_rule_based_global", Keys: []string{"status:1", "stock_status:1", "quality_flags.is_recommendable:1", "quality_flags.is_deleted:1", "counters.purchases_7d:-1", "product_id:1"}, Purpose: "Task 5 eligible global popularity reads."},
					{Name: "idx_product_features_rule_based_category", Keys: []string{"category_id:1", "status:1", "stock_status:1", "quality_flags.is_recommendable:1", "quality_flags.is_deleted:1", "counters.purchases_7d:-1", "product_id:1"}, Purpose: "Task 5 eligible category popularity reads."},
					{Name: "idx_product_features_rule_based_seller", Keys: []string{"seller_id:1", "status:1", "stock_status:1", "quality_flags.is_recommendable:1", "quality_flags.is_deleted:1", "counters.purchases_7d:-1", "product_id:1"}, Purpose: "Task 5 eligible seller popularity reads."},
					{Name: "idx_product_features_category_recommendable_purchases", Keys: []string{"category_id:1", "quality_flags.is_recommendable:1", "counters.purchases_7d:-1"}, Purpose: "Category popularity candidate reads."},
					{Name: "idx_product_features_seller_recommendable_purchases", Keys: []string{"seller_id:1", "quality_flags.is_recommendable:1", "counters.purchases_7d:-1"}, Purpose: "Seller popularity candidate reads."},
					{Name: "idx_product_features_brand_recommendable", Keys: []string{"brand_id:1", "quality_flags.is_recommendable:1"}, Purpose: "Brand candidate reads."},
					{Name: "idx_product_features_personalized_brand", Keys: []string{"brand_id:1", "status:1", "stock_status:1", "quality_flags.is_recommendable:1", "quality_flags.is_deleted:1", "counters.purchases_7d:-1", "product_id:1"}, Purpose: "Task 6 safe brand-affinity personalized candidate reads."},
					{Name: "idx_product_features_embedding_vector_id", Keys: []string{"embedding_refs.vector_id:1"}, Purpose: "Embedding reference lookup."},
				},
			},
			{
				Name:        CollectionUserFeatureProfiles,
				Purpose:     "Compact known-user or anonymous interest profile with optional embedding references.",
				DetailLevel: "derived feature store collection; anonymous profiles expire automatically",
				Indexes: []MongoIndexPlan{
					{Name: "ux_user_feature_profiles_profile_key", Keys: []string{"profile_key:1"}, Unique: true, Purpose: "One interest profile per identity."},
					{Name: "idx_user_feature_profiles_user_id", Keys: []string{"user_id:1"}, Purpose: "Privacy deletion and user lookup."},
					{Name: "idx_user_feature_profiles_anonymous_id", Keys: []string{"anonymous_id:1"}, Purpose: "Guest lookup."},
					{Name: "idx_user_feature_profiles_expires_at_ttl", Keys: []string{"expires_at:1"}, ExpireAfterSeconds: 0, Purpose: "Remove stale anonymous profiles."},
				},
			},
			{
				Name:        CollectionUserProductCounters,
				Purpose:     "Per identity/product interaction counters and weighted affinity inputs.",
				DetailLevel: "derived high-cardinality features with retention cleanup",
				Indexes: []MongoIndexPlan{
					{Name: "ux_user_product_counters_profile_product", Keys: []string{"profile_key:1", "product_id:1"}, Unique: true, Purpose: "One compact counter document per identity/product."},
					{Name: "idx_user_product_counters_profile_recent", Keys: []string{"profile_key:1", "last_interaction_at:-1"}, Purpose: "Recent affinity reads."},
					{Name: "idx_user_product_counters_product_score", Keys: []string{"product_id:1", "weighted_score_input:-1"}, Purpose: "Product affinity analysis."},
					{Name: "idx_user_product_counters_category_recent", Keys: []string{"category_id:1", "last_interaction_at:-1"}, Purpose: "Category behavior reads."},
					{Name: "idx_user_product_counters_expires_at_ttl", Keys: []string{"expires_at:1"}, ExpireAfterSeconds: 0, Purpose: "Remove old pair signals."},
				},
			},
			{
				Name:        CollectionProductCooccurrence,
				Purpose:     "Stable product-pair signals derived from multi-product purchases.",
				DetailLevel: "derived co-occurrence candidates; no ranking is performed here",
				Indexes: []MongoIndexPlan{
					{Name: "idx_product_cooccurrence_source_type_score", Keys: []string{"source_product_id:1", "relationship_type:1", "score_input:-1"}, Purpose: "Source product pair reads."},
					{Name: "idx_product_cooccurrence_related", Keys: []string{"related_product_id:1"}, Purpose: "Canonical pair reverse lookup."},
					{Name: "idx_product_cooccurrence_category_score", Keys: []string{"category_id:1", "score_input:-1"}, Purpose: "Category pair discovery."},
					{Name: "idx_product_cooccurrence_last_seen", Keys: []string{"last_seen_at:-1"}, Purpose: "Freshness maintenance."},
				},
			},
			{
				Name:        CollectionFeatureJobRuns,
				Purpose:     "Feature builder run status, watermark, and update statistics.",
				DetailLevel: "operational audit metadata",
				Indexes: []MongoIndexPlan{
					{Name: "idx_feature_job_runs_type_started", Keys: []string{"job_type:1", "started_at:-1"}, Purpose: "Latest run inspection."},
					{Name: "idx_feature_job_runs_status_started", Keys: []string{"status:1", "started_at:-1"}, Purpose: "Failed or running job monitoring."},
				},
			},
			{
				Name:        CollectionFeatureProcessedEvents,
				Purpose:     "Short-lived idempotency claims for transactional feature updates.",
				DetailLevel: "implementation safety collection",
				Indexes: []MongoIndexPlan{
					{Name: "ux_feature_processed_events_event_id", Keys: []string{"event_id:1"}, Unique: true, Purpose: "Prevent feature counter double application."},
					{Name: "idx_feature_processed_events_expires_at_ttl", Keys: []string{"expires_at:1"}, ExpireAfterSeconds: 0, Purpose: "Bound idempotency claim storage."},
				},
			},
			{
				Name:        CollectionABTestAssignments,
				Purpose:     "Durable recommendation strategy A/B assignment ownership.",
				DetailLevel: "one stable bucket assignment per experiment and assignment identity",
				Indexes: []MongoIndexPlan{
					{Name: "ux_ab_assignment_experiment_identity", Keys: []string{"experiment_id:1", "assignment_key:1"}, Unique: true, Purpose: "Prevent duplicate bucket assignment for the same experiment identity."},
					{Name: "idx_ab_assignment_user_experiment", Keys: []string{"user_id:1", "experiment_id:1"}, Purpose: "Debug known-user experiment assignment."},
					{Name: "idx_ab_assignment_anon_experiment", Keys: []string{"anonymous_id:1", "experiment_id:1"}, Purpose: "Debug anonymous visitor experiment assignment."},
					{Name: "idx_ab_assignment_variant_recent", Keys: []string{"experiment_id:1", "variant_id:1", "assigned_at:-1"}, Purpose: "Inspect recent assignment distribution per variant."},
					{Name: "idx_ab_assignment_ttl", Keys: []string{"expires_at:1"}, ExpireAfterSeconds: 0, Purpose: "Remove old experiment assignments automatically."},
				},
			},
		},
		RedisCaches: []RedisCachePlan{
			{Name: "global_trending", KeyPattern: cacheKeyPrefix + ":trending:global:strategy:{strategy_id}", ValueType: CacheValueTypeJSONString, TTL: ttl.Default, Purpose: "Global top list response cache partitioned by strategy."},
			{Name: "category_trending", KeyPattern: cacheKeyPrefix + ":trending:category:{category_id}:strategy:{strategy_id}", ValueType: CacheValueTypeJSONString, TTL: ttl.Default, Purpose: "Category top list response cache partitioned by strategy."},
			{Name: "seller_popular", KeyPattern: cacheKeyPrefix + ":trending:seller:{seller_id}:strategy:{strategy_id}", ValueType: CacheValueTypeJSONString, TTL: ttl.Default, Purpose: "Seller storefront popular products cache partitioned by strategy."},
			{Name: "similar_products", KeyPattern: cacheKeyPrefix + ":similar:product:{product_id}:strategy:{strategy_id}", ValueType: CacheValueTypeJSONString, TTL: ttl.Default, Purpose: "Product detail similar products cache partitioned by strategy."},
			{Name: "personalized_user", KeyPattern: cacheKeyPrefix + ":personalized:user:{user_id}:strategy:{strategy_id}", ValueType: CacheValueTypeJSONString, TTL: ttl.Personalized, Purpose: "Known-user personalized feed cache partitioned by strategy."},
			{Name: "personalized_guest", KeyPattern: cacheKeyPrefix + ":personalized:anon:{anonymous_id}:strategy:{strategy_id}", ValueType: CacheValueTypeJSONString, TTL: ttl.Guest, Purpose: "Anonymous-session personalized feed cache partitioned by strategy."},
			{Name: "frequently_bought_together", KeyPattern: cacheKeyPrefix + ":fbt:product:{product_id}:strategy:{strategy_id}", ValueType: CacheValueTypeJSONString, TTL: ttl.Default, Purpose: "Bought-together product cache partitioned by strategy."},
			{Name: "dirty_product", KeyPattern: cacheKeyPrefix + ":dirty:product:{product_id}", ValueType: CacheValueTypeString, TTL: ttl.Default, Purpose: "Best-effort marker set when product-level interactions may stale cached rankings."},
			{Name: "dirty_category", KeyPattern: cacheKeyPrefix + ":dirty:category:{category_id}", ValueType: CacheValueTypeString, TTL: ttl.Default, Purpose: "Best-effort marker set when category-level interactions may stale cached rankings."},
			{Name: "dirty_seller", KeyPattern: cacheKeyPrefix + ":dirty:seller:{seller_id}", ValueType: CacheValueTypeString, TTL: ttl.Default, Purpose: "Best-effort marker set when seller-level interactions may stale cached rankings."},
			{Name: "rebuild_lock", KeyPattern: cacheKeyPrefix + ":lock:{cache_key}", ValueType: CacheValueTypeLock, TTL: ttl.RebuildLock, Purpose: "Short rebuild lock for future asynchronous cache refreshes."},
		},
		FailureBehavior: []FailureBehavior{
			{Failure: "redis_down", Behavior: "Read generated recommendation sets from MongoDB."},
			{Failure: "mongodb_down", Behavior: "Serve available Redis cached recommendations; otherwise return a safe empty response."},
			{Failure: "cache_corrupt_json", Behavior: "Delete the corrupt cache key and fall back to MongoDB."},
			{Failure: "both_down", Behavior: "Return a safe empty recommendation list so product, cart, and checkout flows continue."},
		},
		TTLPolicy: ttl,
	}
}
