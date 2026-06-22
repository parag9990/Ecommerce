const databaseName =
  typeof process !== "undefined" && process.env.SESSION_MONGO_DATABASE
    ? process.env.SESSION_MONGO_DATABASE
    : "session_db";

const sessionDB = db.getSiblingDB(databaseName);
if (sessionDB.getCollectionInfos({ name: "report_schedules" }).length > 0) {
  sessionDB.report_schedules.drop();
}
