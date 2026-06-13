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

dropIndexIfExists("product_features", "idx_product_features_category_recommendable_purchases");
dropIndexIfExists("product_features", "idx_product_features_seller_recommendable_purchases");
dropIndexIfExists("product_features", "idx_product_features_brand_recommendable");
dropIndexIfExists("product_features", "idx_product_features_embedding_vector_id");

if (recommendationDB.getCollectionNames().includes("product_features")) {
  recommendationDB.product_features.createIndex(
    { category_id: 1, status: 1, in_stock: 1 },
    { name: "idx_product_features_category_status_stock" },
  );
  recommendationDB.product_features.createIndex(
    { seller_id: 1, status: 1, in_stock: 1 },
    { name: "idx_product_features_seller_status_stock" },
  );
}

for (const collectionName of [
  "user_feature_profiles",
  "user_product_counters",
  "product_cooccurrence_features",
  "feature_job_runs",
  "feature_processed_events",
]) {
  if (recommendationDB.getCollectionNames().includes(collectionName)) {
    recommendationDB.getCollection(collectionName).drop();
  }
}
