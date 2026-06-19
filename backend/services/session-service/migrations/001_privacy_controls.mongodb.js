// MongoDB creates a unique _id index automatically for privacy_settings.

db.analytics_deletion_requests.createIndex(
  { status: 1, created_at: -1 },
  { name: "analytics_deletion_requests_status_created_at" }
);

db.analytics_deletion_requests.createIndex(
  { requested_by: 1, created_at: -1 },
  { name: "analytics_deletion_requests_requested_by_created_at" }
);

db.analytics_deletion_requests.createIndex(
  { target_hash: 1, created_at: -1 },
  { name: "analytics_deletion_requests_target_hash_created_at" }
);

db.analytics_deletion_requests.createIndex(
  { created_at: 1 },
  {
    expireAfterSeconds: 63072000,
    name: "analytics_deletion_requests_created_at_ttl"
  }
);

db.admin_audit_events.createIndex(
  { actor_id: 1, created_at: -1 },
  { name: "admin_audit_actor_created_at" }
);

db.admin_audit_events.createIndex(
  { action: 1, created_at: -1 },
  { name: "admin_audit_action_created_at" }
);
