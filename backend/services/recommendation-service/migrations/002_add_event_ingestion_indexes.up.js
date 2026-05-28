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

ensureCollection("user_interactions");

dropIndexIfExists("user_interactions", "idx_user_interactions_product_event_recent");
dropIndexIfExists("user_interactions", "idx_user_interactions_category_event_recent");

recommendationDB.user_interactions.createIndex(
  { dedupe_key: 1 },
  {
    name: "ux_user_interactions_dedupe_key",
    unique: true,
    partialFilterExpression: { dedupe_key: { $type: "string" } },
  },
);

recommendationDB.user_interactions.createIndex(
  { product_id: 1, normalized_event_type: 1, occurred_at: -1 },
  { name: "idx_user_interactions_product_normalized_event_recent" },
);

recommendationDB.user_interactions.createIndex(
  { category_id: 1, normalized_event_type: 1, occurred_at: -1 },
  { name: "idx_user_interactions_category_normalized_event_recent" },
);
