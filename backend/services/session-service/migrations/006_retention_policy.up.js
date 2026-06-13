const databaseName =
  typeof process !== "undefined" && process.env.SESSION_MONGO_DATABASE
    ? process.env.SESSION_MONGO_DATABASE
    : "session_db";

const sessionDB = db.getSiblingDB(databaseName);

function ensureCollection(name, validator) {
  const exists = sessionDB.getCollectionInfos({ name }).length > 0;
  if (!exists) {
    sessionDB.createCollection(name, {
      validator,
      validationLevel: "moderate",
      validationAction: "error",
    });
    return;
  }

  sessionDB.runCommand({
    collMod: name,
    validator,
    validationLevel: "moderate",
    validationAction: "error",
  });
}

function retentionDays(envName, fallback) {
  if (typeof process !== "undefined" && process.env[envName]) {
    const parsed = Number(process.env[envName]);
    if (Number.isFinite(parsed) && parsed > 0) {
      return parsed;
    }
  }
  return fallback;
}

const sessionRetentionDays = retentionDays("SESSION_METADATA_RETENTION_DAYS", 365);
const journeyRetentionDays = retentionDays("SESSION_JOURNEY_RETENTION_DAYS", 365);
const heatmapRetentionDays = retentionDays("SESSION_HEATMAP_RETENTION_DAYS", 730);

ensureCollection("retention_deletion_audits", {
  $jsonSchema: {
    bsonType: "object",
    required: [
      "_id",
      "request_id",
      "reason",
      "hard_delete",
      "completed_at",
      "retention_class",
      "created_at",
    ],
    properties: {
      _id: { bsonType: "string" },
      request_id: { bsonType: "string" },
      user_id_hash: { bsonType: ["string", "null"] },
      anonymous_id_hashes: { bsonType: ["array", "null"], items: { bsonType: "string" } },
      reason: { bsonType: "string" },
      hard_delete: { bsonType: "bool" },
      requested_by: { bsonType: ["string", "null"] },
      requested_at: { bsonType: ["date", "null"] },
      raw_events_deleted: { bsonType: ["int", "long"] },
      sessions_anonymized: { bsonType: ["int", "long"] },
      sessions_deleted: { bsonType: ["int", "long"] },
      journeys_anonymized: { bsonType: ["int", "long"] },
      journeys_deleted: { bsonType: ["int", "long"] },
      redis_keys_deleted: { bsonType: ["int", "long"] },
      redis_errors: { bsonType: ["int", "long"] },
      completed_at: { bsonType: "date" },
      retention_class: { enum: ["deletion_audit"] },
      created_at: { bsonType: "date" },
    },
  },
});

sessionDB.sessions.createIndex(
  { retain_until: 1, legal_hold: 1, anonymized_at: 1 },
  { name: "idx_sessions_retention_due" }
);

sessionDB.sessions.createIndex(
  { retention_class: 1, last_seen_at: -1 },
  { name: "idx_sessions_retention_class_last_seen" }
);

sessionDB.journey_summaries.createIndex(
  { retain_until: 1, legal_hold: 1, anonymized_at: 1 },
  { name: "idx_journey_retention_due" }
);

sessionDB.journey_summaries.createIndex(
  { retention_class: 1, last_event_at: -1 },
  { name: "idx_journey_retention_class_recent" }
);

sessionDB.heatmap_points.createIndex(
  { retain_until: 1 },
  { name: "idx_heatmap_retention_due" }
);

sessionDB.heatmap_bucket_sessions.createIndex(
  { first_seen_at: 1 },
  { name: "idx_heatmap_markers_first_seen" }
);

sessionDB.analytics_aggregates.createIndex(
  { contains_pii: 1, updated_at: -1 },
  { name: "idx_aggregates_pii_policy" }
);

sessionDB.analytics_aggregates.createIndex(
  { bucket_end: 1, calculated_at: 1 },
  { name: "idx_aggregates_retention_due" }
);

sessionDB.retention_deletion_audits.createIndex(
  { request_id: 1 },
  { unique: true, name: "uniq_retention_deletion_request" }
);

sessionDB.retention_deletion_audits.createIndex(
  { completed_at: -1 },
  { name: "idx_retention_deletion_completed" }
);

sessionDB.retention_deletion_audits.createIndex(
  { user_id_hash: 1, completed_at: -1 },
  { name: "idx_retention_deletion_user_hash" }
);

sessionDB.sessions.updateMany(
  { retain_until: { $exists: false } },
  [
    {
      $set: {
        retain_until: {
          $dateAdd: {
            startDate: { $ifNull: ["$ended_at", "$last_seen_at"] },
            unit: "day",
            amount: sessionRetentionDays,
          },
        },
        retention_class: "session_metadata",
        legal_hold: { $ifNull: ["$legal_hold", false] },
      },
    },
  ]
);

sessionDB.journey_summaries.updateMany(
  { retain_until: { $exists: false } },
  [
    {
      $set: {
        retain_until: {
          $dateAdd: {
            startDate: { $ifNull: ["$last_event_at", "$updated_at"] },
            unit: "day",
            amount: journeyRetentionDays,
          },
        },
        retention_class: "derived_summary",
        legal_hold: { $ifNull: ["$legal_hold", false] },
      },
    },
  ]
);

sessionDB.heatmap_points.updateMany(
  { retain_until: { $exists: false } },
  [
    {
      $set: {
        retain_until: {
          $dateAdd: {
            startDate: "$last_seen_at",
            unit: "day",
            amount: heatmapRetentionDays,
          },
        },
        retention_class: "aggregate_fine",
      },
    },
  ]
);

sessionDB.analytics_aggregates.updateMany(
  { retention_class: { $exists: false } },
  {
    $set: {
      retention_class: "aggregate_business",
      contains_pii: false,
    },
  }
);

sessionDB.session_events.updateMany(
  { retention_class: { $exists: false } },
  { $set: { retention_class: "raw_event" } }
);
