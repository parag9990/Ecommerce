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

dropIndexIfExists("session_events", "ttl_raw_events_90_days");
dropIndexIfExists("session_events", "idx_session_events_request_id");
dropIndexIfExists("session_events", "idx_event_type_time");
dropIndexIfExists("session_events", "idx_user_event_history");
dropIndexIfExists("session_events", "idx_session_timeline");

dropIndexIfExists("sessions", "idx_status_last_seen");
dropIndexIfExists("sessions", "idx_user_sessions");
dropIndexIfExists("sessions", "idx_anonymous_sessions");
dropIndexIfExists("sessions", "uniq_session_id");

if (sessionDB.getCollectionInfos({ name: "session_events" }).length > 0) {
  sessionDB.runCommand({
    collMod: "session_events",
    validator: {},
    validationLevel: "off",
  });
}

if (sessionDB.getCollectionInfos({ name: "sessions" }).length > 0) {
  sessionDB.runCommand({
    collMod: "sessions",
    validator: {},
    validationLevel: "off",
  });
}
