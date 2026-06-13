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

dropIndexIfExists("session_events", "idx_events_device_event_time");

dropIndexIfExists("sessions", "idx_sessions_device_fingerprint_last_seen");
dropIndexIfExists("sessions", "idx_sessions_ip_hash_last_seen");
dropIndexIfExists("sessions", "idx_sessions_country_started");
dropIndexIfExists("sessions", "idx_sessions_browser_os_started");
dropIndexIfExists("sessions", "idx_sessions_device_type_started");
