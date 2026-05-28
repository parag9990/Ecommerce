const databaseName =
  typeof process !== "undefined" && process.env.RECOMMENDATION_MONGO_DATABASE
    ? process.env.RECOMMENDATION_MONGO_DATABASE
    : "recommendation_db";

const recommendationDB = db.getSiblingDB(databaseName);

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

dropIndexIfExists("user_interactions", "idx_user_interactions_user_recent");
dropIndexIfExists("user_interactions", "idx_user_interactions_anon_recent");
dropIndexIfExists("user_interactions", "idx_user_interactions_product_event_recent");
dropIndexIfExists("user_interactions", "idx_user_interactions_category_event_recent");
dropIndexIfExists("user_interactions", "idx_user_interactions_ttl");

dropIndexIfExists("recommendation_sets", "ux_recommendation_sets_context_strategy");
dropIndexIfExists("recommendation_sets", "idx_recommendation_sets_expires_at_ttl");
dropIndexIfExists("recommendation_sets", "idx_recommendation_sets_type_generated");

dropIndexIfExists("product_features", "ux_product_features_product_id");
dropIndexIfExists("product_features", "idx_product_features_category_status_stock");
dropIndexIfExists("product_features", "idx_product_features_seller_status_stock");
