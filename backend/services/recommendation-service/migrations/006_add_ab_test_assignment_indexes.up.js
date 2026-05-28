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

ensureCollection("ab_test_assignments");

recommendationDB.ab_test_assignments.createIndex(
  { experiment_id: 1, assignment_key: 1 },
  { name: "ux_ab_assignment_experiment_identity", unique: true },
);

recommendationDB.ab_test_assignments.createIndex(
  { user_id: 1, experiment_id: 1 },
  { name: "idx_ab_assignment_user_experiment", sparse: true },
);

recommendationDB.ab_test_assignments.createIndex(
  { anonymous_id: 1, experiment_id: 1 },
  { name: "idx_ab_assignment_anon_experiment", sparse: true },
);

recommendationDB.ab_test_assignments.createIndex(
  { experiment_id: 1, variant_id: 1, assigned_at: -1 },
  { name: "idx_ab_assignment_variant_recent" },
);

recommendationDB.ab_test_assignments.createIndex(
  { expires_at: 1 },
  { name: "idx_ab_assignment_ttl", expireAfterSeconds: 0 },
);
