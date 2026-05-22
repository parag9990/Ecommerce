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

ensureCollection("analytics_aggregates", {
  $jsonSchema: {
    bsonType: "object",
    required: ["metric", "schema_version", "calculated_at", "updated_at"],
    properties: {
      _id: { bsonType: "string" },
      metric: {
        bsonType: "string",
        description: "Metric family, for example funnel.checkout or retention.users",
      },
      bucket: { enum: ["hourly", "daily", "custom", null] },
      bucket_start: { bsonType: ["date", "null"] },
      bucket_end: { bsonType: ["date", "null"] },
      cohort_start: { bsonType: ["date", "null"] },
      cohort_end: { bsonType: ["date", "null"] },
      segment: { bsonType: ["object", "null"] },
      steps: {
        bsonType: ["array", "null"],
        items: {
          bsonType: "object",
          required: ["name", "count", "unique_sessions"],
          properties: {
            name: { bsonType: "string" },
            event_type: { bsonType: ["string", "null"] },
            count: { bsonType: ["int", "long", "double"] },
            unique_sessions: { bsonType: ["int", "long", "double"] },
            unique_users: { bsonType: ["int", "long", "double", "null"] },
            conversion_from_previous: { bsonType: ["double", "int", "long", "null"] },
            dropoff_from_previous: { bsonType: ["double", "int", "long", "null"] },
          },
        },
      },
      values: { bsonType: ["object", "null"] },
      cohort_size: { bsonType: ["int", "long", "double", "null"] },
      retention: { bsonType: ["object", "null"] },
      rates: { bsonType: ["object", "null"] },
      schema_version: { bsonType: "int" },
      calculated_at: { bsonType: "date" },
      created_at: { bsonType: ["date", "null"] },
      updated_at: { bsonType: "date" },
    },
  },
});

sessionDB.analytics_aggregates.createIndex(
  {
    metric: 1,
    bucket: 1,
    bucket_start: 1,
    bucket_end: 1,
    "segment.channel": 1,
    "segment.device_type": 1,
    "segment.country": 1,
    "segment.campaign": 1,
  },
  {
    unique: true,
    name: "uniq_metric_bucket_time_segment",
    partialFilterExpression: {
      bucket_start: { $type: "date" },
      bucket_end: { $type: "date" },
    },
  }
);

sessionDB.analytics_aggregates.createIndex(
  { metric: 1, bucket_start: -1 },
  { name: "idx_metric_time" }
);

sessionDB.analytics_aggregates.createIndex(
  { "segment.channel": 1, "segment.device_type": 1, bucket_start: -1 },
  { name: "idx_segment_time" }
);

sessionDB.analytics_aggregates.createIndex(
  { metric: 1, cohort_start: -1 },
  { name: "idx_retention_cohort" }
);

sessionDB.sessions.createIndex(
  { channel: 1, started_at: -1 },
  { name: "idx_sessions_channel_started" }
);
