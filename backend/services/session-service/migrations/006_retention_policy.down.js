const databaseName =
  typeof process !== "undefined" && process.env.SESSION_MONGO_DATABASE
    ? process.env.SESSION_MONGO_DATABASE
    : "session_db";

const sessionDB = db.getSiblingDB(databaseName);

function dropIndexIfExists(collectionName, indexName) {
  const collection = sessionDB.getCollection(collectionName);
  const exists = collection.getIndexes().some((index) => index.name === indexName);
  if (exists) {
    collection.dropIndex(indexName);
  }
}

dropIndexIfExists("retention_deletion_audits", "idx_retention_deletion_user_hash");
dropIndexIfExists("retention_deletion_audits", "idx_retention_deletion_completed");
dropIndexIfExists("retention_deletion_audits", "uniq_retention_deletion_request");

dropIndexIfExists("analytics_aggregates", "idx_aggregates_retention_due");
dropIndexIfExists("analytics_aggregates", "idx_aggregates_pii_policy");
dropIndexIfExists("heatmap_bucket_sessions", "idx_heatmap_markers_first_seen");
dropIndexIfExists("heatmap_points", "idx_heatmap_retention_due");
dropIndexIfExists("journey_summaries", "idx_journey_retention_class_recent");
dropIndexIfExists("journey_summaries", "idx_journey_retention_due");
dropIndexIfExists("sessions", "idx_sessions_retention_class_last_seen");
dropIndexIfExists("sessions", "idx_sessions_retention_due");

if (sessionDB.getCollectionInfos({ name: "retention_deletion_audits" }).length > 0) {
  sessionDB.runCommand({
    collMod: "retention_deletion_audits",
    validator: {},
    validationLevel: "off",
  });
}
