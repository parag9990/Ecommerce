const databaseName =
  typeof process !== "undefined" && process.env.SESSION_MONGO_DATABASE
    ? process.env.SESSION_MONGO_DATABASE
    : "session_db";

const sessionDB = db.getSiblingDB(databaseName);

sessionDB.analytics_aggregates.dropIndex("uniq_metric_bucket_time_segment");
sessionDB.analytics_aggregates.dropIndex("idx_metric_time");
sessionDB.analytics_aggregates.dropIndex("idx_segment_time");
sessionDB.analytics_aggregates.dropIndex("idx_retention_cohort");
sessionDB.sessions.dropIndex("idx_sessions_channel_started");

sessionDB.analytics_aggregates.drop();
