const databaseName =
  typeof process !== "undefined" && process.env.SESSION_MONGO_DATABASE
    ? process.env.SESSION_MONGO_DATABASE
    : "session_db";

const sessionDB = db.getSiblingDB(databaseName);

if (sessionDB.getCollectionInfos({ name: "journey_summaries" }).length > 0) {
  sessionDB.journey_summaries.drop();
}
