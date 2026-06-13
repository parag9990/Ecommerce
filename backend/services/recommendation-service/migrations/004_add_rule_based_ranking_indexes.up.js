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

ensureCollection("product_features");

recommendationDB.product_features.createIndex(
  {
    status: 1,
    stock_status: 1,
    "quality_flags.is_recommendable": 1,
    "quality_flags.is_deleted": 1,
    "counters.purchases_7d": -1,
    product_id: 1,
  },
  { name: "idx_product_features_rule_based_global" },
);

recommendationDB.product_features.createIndex(
  {
    category_id: 1,
    status: 1,
    stock_status: 1,
    "quality_flags.is_recommendable": 1,
    "quality_flags.is_deleted": 1,
    "counters.purchases_7d": -1,
    product_id: 1,
  },
  { name: "idx_product_features_rule_based_category" },
);

recommendationDB.product_features.createIndex(
  {
    seller_id: 1,
    status: 1,
    stock_status: 1,
    "quality_flags.is_recommendable": 1,
    "quality_flags.is_deleted": 1,
    "counters.purchases_7d": -1,
    product_id: 1,
  },
  { name: "idx_product_features_rule_based_seller" },
);
