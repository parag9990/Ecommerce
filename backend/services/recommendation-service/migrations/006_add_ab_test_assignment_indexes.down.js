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

dropIndexIfExists("ab_test_assignments", "ux_ab_assignment_experiment_identity");
dropIndexIfExists("ab_test_assignments", "idx_ab_assignment_user_experiment");
dropIndexIfExists("ab_test_assignments", "idx_ab_assignment_anon_experiment");
dropIndexIfExists("ab_test_assignments", "idx_ab_assignment_variant_recent");
dropIndexIfExists("ab_test_assignments", "idx_ab_assignment_ttl");
