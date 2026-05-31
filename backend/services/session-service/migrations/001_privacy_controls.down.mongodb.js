function dropIndexIfExists(collection, name) {
  if (collection.getIndexes().some((index) => index.name === name)) {
    collection.dropIndex(name);
  }
}

dropIndexIfExists(
  db.analytics_deletion_requests,
  "analytics_deletion_requests_status_created_at"
);
dropIndexIfExists(
  db.analytics_deletion_requests,
  "analytics_deletion_requests_requested_by_created_at"
);
dropIndexIfExists(
  db.analytics_deletion_requests,
  "analytics_deletion_requests_target_hash_created_at"
);
dropIndexIfExists(
  db.analytics_deletion_requests,
  "analytics_deletion_requests_created_at_ttl"
);

dropIndexIfExists(db.admin_audit_events, "admin_audit_actor_created_at");
dropIndexIfExists(db.admin_audit_events, "admin_audit_action_created_at");
