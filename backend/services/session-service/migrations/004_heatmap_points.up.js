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

ensureCollection("heatmap_points", {
  $jsonSchema: {
    bsonType: "object",
    required: [
      "_id",
      "heatmap_type",
      "path",
      "normalized_path",
      "device_type",
      "viewport_bucket",
      "day",
      "x",
      "y",
      "weight",
      "unique_sessions",
      "sample_events",
      "schema_version",
      "first_seen_at",
      "last_seen_at",
      "updated_at",
    ],
    properties: {
      _id: { bsonType: "string" },
      heatmap_type: { enum: ["click", "scroll"] },
      path: { bsonType: "string" },
      normalized_path: { bsonType: "string" },
      device_type: { enum: ["desktop", "mobile", "tablet", "bot", "unknown"] },
      viewport_bucket: { bsonType: "string" },
      day: { bsonType: "string" },
      x_bucket: { bsonType: ["int", "long"] },
      y_bucket: { bsonType: ["int", "long"] },
      depth_bucket: { bsonType: ["int", "long"] },
      x: { bsonType: ["int", "long"] },
      y: { bsonType: ["int", "long"] },
      weight: { bsonType: ["int", "long"] },
      unique_sessions: { bsonType: ["int", "long"] },
      sample_events: { bsonType: ["int", "long"] },
      schema_version: { bsonType: "int" },
      first_seen_at: { bsonType: "date" },
      last_seen_at: { bsonType: "date" },
      created_at: { bsonType: "date" },
      updated_at: { bsonType: "date" },
    },
  },
});

ensureCollection("heatmap_bucket_sessions", {
  $jsonSchema: {
    bsonType: "object",
    required: ["_id", "bucket_key", "point_id", "first_seen_at", "updated_at"],
    properties: {
      _id: { bsonType: "string" },
      bucket_key: { bsonType: "string" },
      point_id: { bsonType: "string" },
      first_seen_at: { bsonType: "date" },
      updated_at: { bsonType: "date" },
    },
  },
});

ensureCollection("heatmap_checkpoints", {
  $jsonSchema: {
    bsonType: "object",
    required: ["_id", "worker_name", "last_processed_at", "updated_at"],
    properties: {
      _id: { bsonType: "string" },
      worker_name: { bsonType: "string" },
      last_processed_at: { bsonType: "date" },
      updated_at: { bsonType: "date" },
    },
  },
});

ensureCollection("heatmap_processed_events", {
  $jsonSchema: {
    bsonType: "object",
    required: ["_id", "event_id", "point_id", "processed_at", "retain_until"],
    properties: {
      _id: { bsonType: "string" },
      event_id: { bsonType: "string" },
      point_id: { bsonType: "string" },
      processed_at: { bsonType: "date" },
      retain_until: { bsonType: "date" },
    },
  },
});

sessionDB.heatmap_points.createIndex(
  { path: 1, device_type: 1, day: 1 },
  { name: "idx_heatmap_route_device_day" }
);

sessionDB.heatmap_points.createIndex(
  {
    heatmap_type: 1,
    path: 1,
    device_type: 1,
    viewport_bucket: 1,
    day: 1,
    x_bucket: 1,
    y_bucket: 1,
  },
  {
    unique: true,
    name: "uniq_heatmap_click_bucket",
    partialFilterExpression: { heatmap_type: "click" },
  }
);

sessionDB.heatmap_points.createIndex(
  {
    heatmap_type: 1,
    path: 1,
    device_type: 1,
    viewport_bucket: 1,
    day: 1,
    depth_bucket: 1,
  },
  {
    unique: true,
    name: "uniq_heatmap_scroll_bucket",
    partialFilterExpression: { heatmap_type: "scroll" },
  }
);

if (sessionDB.heatmap_bucket_sessions.getIndexes().some((index) => index.name === "uniq_heatmap_bucket_session")) {
  sessionDB.heatmap_bucket_sessions.dropIndex("uniq_heatmap_bucket_session");
}

sessionDB.heatmap_bucket_sessions.updateMany(
  { session_id: { $exists: true } },
  { $unset: { session_id: "" } }
);

sessionDB.heatmap_bucket_sessions.createIndex(
  { point_id: 1 },
  { name: "idx_heatmap_bucket_sessions_point" }
);

sessionDB.heatmap_checkpoints.createIndex(
  { worker_name: 1 },
  { unique: true, name: "uniq_heatmap_checkpoint_worker" }
);

sessionDB.heatmap_processed_events.createIndex(
  { event_id: 1 },
  { unique: true, name: "uniq_heatmap_processed_event" }
);

sessionDB.heatmap_processed_events.createIndex(
  { retain_until: 1 },
  { expireAfterSeconds: 0, name: "ttl_heatmap_processed_events" }
);

sessionDB.heatmap_processed_events.updateMany(
  { session_id: { $exists: true } },
  { $unset: { session_id: "" } }
);

sessionDB.session_events.createIndex(
  { event_type: 1, occurred_at: 1, path: 1 },
  { name: "idx_events_heatmap_scan" }
);
