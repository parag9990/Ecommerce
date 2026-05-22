const databaseName =
  typeof process !== "undefined" && process.env.SESSION_MONGO_DATABASE
    ? process.env.SESSION_MONGO_DATABASE
    : "session_db";

const sessionDB = db.getSiblingDB(databaseName);

sessionDB.sessions.createIndex(
  { "device.type": 1, started_at: -1 },
  { name: "idx_sessions_device_type_started" }
);

sessionDB.sessions.createIndex(
  { "device.browser": 1, "device.os": 1, started_at: -1 },
  { name: "idx_sessions_browser_os_started" }
);

sessionDB.sessions.createIndex(
  { "geo.country": 1, started_at: -1 },
  { name: "idx_sessions_country_started" }
);

sessionDB.sessions.createIndex(
  { ip_hash: 1, last_seen_at: -1 },
  { name: "idx_sessions_ip_hash_last_seen" }
);

sessionDB.sessions.createIndex(
  { device_fingerprint_hash: 1, last_seen_at: -1 },
  {
    name: "idx_sessions_device_fingerprint_last_seen",
    partialFilterExpression: { device_fingerprint_hash: { $type: "string" } },
  }
);

sessionDB.session_events.createIndex(
  { "device.type": 1, event_type: 1, occurred_at: -1 },
  { name: "idx_events_device_event_time" }
);
