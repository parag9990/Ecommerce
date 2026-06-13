const databaseName =
  typeof process !== "undefined" && process.env.RECOMMENDATION_MONGO_DATABASE
    ? process.env.RECOMMENDATION_MONGO_DATABASE
    : "recommendation_db";

const recommendationDB = db.getSiblingDB(databaseName);

function ensureCollection(name) {
  if (!recommendationDB.getCollectionNames().includes(name)) {
    recommendationDB.createCollection(name);
  }
}

function dropIndexIfExists(collectionName, indexName) {
  if (!recommendationDB.getCollectionNames().includes(collectionName)) {
    return;
  }
  const collection = recommendationDB.getCollection(collectionName);
  const exists = collection.getIndexes().some((index) => index.name === indexName);
  if (exists) {
    collection.dropIndex(indexName);
  }
}

ensureCollection("product_features");
ensureCollection("user_feature_profiles");
ensureCollection("user_product_counters");
ensureCollection("product_cooccurrence_features");
ensureCollection("feature_job_runs");
ensureCollection("feature_processed_events");

dropIndexIfExists("product_features", "idx_product_features_category_status_stock");
dropIndexIfExists("product_features", "idx_product_features_seller_status_stock");

recommendationDB.product_features.createIndex(
  { product_id: 1 },
  { name: "ux_product_features_product_id", unique: true },
);
recommendationDB.product_features.createIndex(
  { category_id: 1, "quality_flags.is_recommendable": 1, "counters.purchases_7d": -1 },
  { name: "idx_product_features_category_recommendable_purchases" },
);
recommendationDB.product_features.createIndex(
  { seller_id: 1, "quality_flags.is_recommendable": 1, "counters.purchases_7d": -1 },
  { name: "idx_product_features_seller_recommendable_purchases" },
);
recommendationDB.product_features.createIndex(
  { brand_id: 1, "quality_flags.is_recommendable": 1 },
  { name: "idx_product_features_brand_recommendable" },
);
recommendationDB.product_features.createIndex(
  { "embedding_refs.vector_id": 1 },
  { name: "idx_product_features_embedding_vector_id", sparse: true },
);

recommendationDB.user_feature_profiles.createIndex(
  { profile_key: 1 },
  { name: "ux_user_feature_profiles_profile_key", unique: true },
);
recommendationDB.user_feature_profiles.createIndex(
  { user_id: 1 },
  { name: "idx_user_feature_profiles_user_id", sparse: true },
);
recommendationDB.user_feature_profiles.createIndex(
  { anonymous_id: 1 },
  { name: "idx_user_feature_profiles_anonymous_id", sparse: true },
);
recommendationDB.user_feature_profiles.createIndex(
  { expires_at: 1 },
  { name: "idx_user_feature_profiles_expires_at_ttl", expireAfterSeconds: 0 },
);

recommendationDB.user_product_counters.createIndex(
  { profile_key: 1, product_id: 1 },
  { name: "ux_user_product_counters_profile_product", unique: true },
);
recommendationDB.user_product_counters.createIndex(
  { profile_key: 1, last_interaction_at: -1 },
  { name: "idx_user_product_counters_profile_recent" },
);
recommendationDB.user_product_counters.createIndex(
  { product_id: 1, weighted_score_input: -1 },
  { name: "idx_user_product_counters_product_score" },
);
recommendationDB.user_product_counters.createIndex(
  { category_id: 1, last_interaction_at: -1 },
  { name: "idx_user_product_counters_category_recent" },
);
recommendationDB.user_product_counters.createIndex(
  { expires_at: 1 },
  { name: "idx_user_product_counters_expires_at_ttl", expireAfterSeconds: 0 },
);

recommendationDB.product_cooccurrence_features.createIndex(
  { source_product_id: 1, relationship_type: 1, score_input: -1 },
  { name: "idx_product_cooccurrence_source_type_score" },
);
recommendationDB.product_cooccurrence_features.createIndex(
  { related_product_id: 1 },
  { name: "idx_product_cooccurrence_related" },
);
recommendationDB.product_cooccurrence_features.createIndex(
  { category_id: 1, score_input: -1 },
  { name: "idx_product_cooccurrence_category_score" },
);
recommendationDB.product_cooccurrence_features.createIndex(
  { last_seen_at: -1 },
  { name: "idx_product_cooccurrence_last_seen" },
);

recommendationDB.feature_job_runs.createIndex(
  { job_type: 1, started_at: -1 },
  { name: "idx_feature_job_runs_type_started" },
);
recommendationDB.feature_job_runs.createIndex(
  { status: 1, started_at: -1 },
  { name: "idx_feature_job_runs_status_started" },
);

recommendationDB.feature_processed_events.createIndex(
  { event_id: 1 },
  { name: "ux_feature_processed_events_event_id", unique: true },
);
recommendationDB.feature_processed_events.createIndex(
  { expires_at: 1 },
  { name: "idx_feature_processed_events_expires_at_ttl", expireAfterSeconds: 0 },
);
