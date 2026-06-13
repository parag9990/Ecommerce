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

ensureCollection("user_interactions");
ensureCollection("product_features");
ensureCollection("recommendation_sets");
ensureCollection("ab_test_assignments");

recommendationDB.user_interactions.createIndex(
  { user_id: 1, occurred_at: -1 },
  { name: "idx_user_interactions_user_recent" },
);
recommendationDB.user_interactions.createIndex(
  { anonymous_id: 1, occurred_at: -1 },
  { name: "idx_user_interactions_anon_recent" },
);
recommendationDB.user_interactions.createIndex(
  { product_id: 1, event_type: 1, occurred_at: -1 },
  { name: "idx_user_interactions_product_event_recent" },
);
recommendationDB.user_interactions.createIndex(
  { category_id: 1, event_type: 1, occurred_at: -1 },
  { name: "idx_user_interactions_category_event_recent" },
);
recommendationDB.user_interactions.createIndex(
  { occurred_at: 1 },
  { name: "idx_user_interactions_ttl", expireAfterSeconds: 15552000 },
);

recommendationDB.recommendation_sets.createIndex(
  { context_key: 1, strategy_id: 1 },
  { name: "ux_recommendation_sets_context_strategy", unique: true },
);
recommendationDB.recommendation_sets.createIndex(
  { expires_at: 1 },
  { name: "idx_recommendation_sets_expires_at_ttl", expireAfterSeconds: 0 },
);
recommendationDB.recommendation_sets.createIndex(
  { recommendation_type: 1, generated_at: -1 },
  { name: "idx_recommendation_sets_type_generated" },
);

recommendationDB.product_features.createIndex(
  { product_id: 1 },
  { name: "ux_product_features_product_id", unique: true },
);
recommendationDB.product_features.createIndex(
  { category_id: 1, status: 1, in_stock: 1 },
  { name: "idx_product_features_category_status_stock" },
);
recommendationDB.product_features.createIndex(
  { seller_id: 1, status: 1, in_stock: 1 },
  { name: "idx_product_features_seller_status_stock" },
);
