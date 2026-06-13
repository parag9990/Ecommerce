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

ensureCollection("sessions", {
  $jsonSchema: {
    bsonType: "object",
    required: [
      "_id",
      "session_id",
      "anonymous_id",
      "schema_version",
      "status",
      "entry_page",
      "channel",
      "user_agent",
      "device",
      "ip_hash",
      "started_at",
      "last_seen_at",
      "created_at",
      "updated_at",
    ],
    properties: {
      _id: { bsonType: "string" },
      session_id: { bsonType: "string" },
      anonymous_id: { bsonType: "string" },
      user_id: { bsonType: "string" },
      schema_version: { bsonType: "int" },
      status: {
        enum: ["active", "ended", "expired", "revoked"],
      },
      entry_page: { bsonType: "string" },
      exit_page: { bsonType: "string" },
      channel: {
        enum: [
          "user_app_web",
          "seller_dashboard_web",
          "superadmin_panel_web",
          "session_analytics_dashboard_web",
          "mobile_web",
          "android_app",
          "ios_app",
          "unknown",
        ],
      },
      referrer: { bsonType: "string" },
      user_agent: { bsonType: "string" },
      device: {
        bsonType: "object",
        required: ["type"],
        properties: {
          type: {
            enum: ["desktop", "mobile", "tablet", "bot", "unknown"],
          },
          browser: { bsonType: "string" },
          browser_version: { bsonType: "string" },
          os: { bsonType: "string" },
          os_version: { bsonType: "string" },
          model: { bsonType: "string" },
        },
      },
      ip_hash: { bsonType: "string" },
      ip_version: {
        enum: ["ipv4", "ipv6", "unknown"],
      },
      started_at: { bsonType: "date" },
      last_seen_at: { bsonType: "date" },
      ended_at: { bsonType: "date" },
      revoked_at: { bsonType: "date" },
      end_reason: {
        enum: [
          "logout",
          "inactivity_timeout",
          "session_replaced",
          "security_revoke",
          "admin_revoke",
          "unknown",
        ],
      },
      created_at: { bsonType: "date" },
      updated_at: { bsonType: "date" },
    },
  },
});

ensureCollection("session_events", {
  $jsonSchema: {
    bsonType: "object",
    required: [
      "_id",
      "event_id",
      "session_id",
      "anonymous_id",
      "event_type",
      "occurred_at",
      "received_at",
      "schema_version",
    ],
    properties: {
      _id: { bsonType: "string" },
      event_id: { bsonType: "string" },
      session_id: { bsonType: "string" },
      anonymous_id: { bsonType: "string" },
      user_id: { bsonType: "string" },
      event_type: { bsonType: "string" },
      path: { bsonType: "string" },
      properties: { bsonType: "object" },
      occurred_at: { bsonType: "date" },
      received_at: { bsonType: "date" },
      schema_version: { bsonType: "int" },
      request_id: { bsonType: "string" },
    },
  },
});

sessionDB.sessions.createIndex(
  { session_id: 1 },
  { unique: true, name: "uniq_session_id" }
);

sessionDB.sessions.createIndex(
  { anonymous_id: 1, started_at: -1 },
  { name: "idx_anonymous_sessions" }
);

sessionDB.sessions.createIndex(
  { user_id: 1, started_at: -1 },
  {
    name: "idx_user_sessions",
    partialFilterExpression: { user_id: { $type: "string" } },
  }
);

sessionDB.sessions.createIndex(
  { status: 1, last_seen_at: -1 },
  { name: "idx_status_last_seen" }
);

sessionDB.session_events.createIndex(
  { session_id: 1, occurred_at: 1 },
  { name: "idx_session_timeline" }
);

sessionDB.session_events.createIndex(
  { user_id: 1, occurred_at: -1 },
  {
    name: "idx_user_event_history",
    partialFilterExpression: { user_id: { $type: "string" } },
  }
);

sessionDB.session_events.createIndex(
  { event_type: 1, occurred_at: -1 },
  { name: "idx_event_type_time" }
);

sessionDB.session_events.createIndex(
  { request_id: 1 },
  {
    name: "idx_session_events_request_id",
    partialFilterExpression: { request_id: { $type: "string" } },
  }
);

sessionDB.session_events.createIndex(
  { occurred_at: 1 },
  { name: "ttl_raw_events_90_days", expireAfterSeconds: 7776000 }
);
