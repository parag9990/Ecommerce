const databaseName =
  typeof process !== "undefined" && process.env.SESSION_MONGO_DATABASE
    ? process.env.SESSION_MONGO_DATABASE
    : "session_db";

const sessionDB = db.getSiblingDB(databaseName);

sessionDB.heatmap_points.dropIndex("idx_heatmap_route_device_day");
sessionDB.heatmap_points.dropIndex("uniq_heatmap_click_bucket");
sessionDB.heatmap_points.dropIndex("uniq_heatmap_scroll_bucket");
sessionDB.heatmap_bucket_sessions.dropIndex("uniq_heatmap_bucket_session");
sessionDB.heatmap_bucket_sessions.dropIndex("idx_heatmap_bucket_sessions_point");
sessionDB.heatmap_checkpoints.dropIndex("uniq_heatmap_checkpoint_worker");
sessionDB.session_events.dropIndex("idx_events_heatmap_scan");

sessionDB.heatmap_points.drop();
sessionDB.heatmap_bucket_sessions.drop();
sessionDB.heatmap_checkpoints.drop();
