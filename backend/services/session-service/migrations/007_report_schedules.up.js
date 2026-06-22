const databaseName =
  typeof process !== "undefined" && process.env.SESSION_MONGO_DATABASE
    ? process.env.SESSION_MONGO_DATABASE
    : "session_db";

const sessionDB = db.getSiblingDB(databaseName);

if (sessionDB.getCollectionInfos({ name: "report_schedules" }).length === 0) {
  sessionDB.createCollection("report_schedules", {
    validator: {
      $jsonSchema: {
        bsonType: "object",
        required: ["_id", "name", "report_type", "format", "frequency", "timezone", "time_of_day", "status", "recipients", "created_by", "created_at", "updated_at"],
        properties: {
          _id: { bsonType: "string" },
          name: { bsonType: "string", minLength: 3, maxLength: 80 },
          report_type: { enum: ["overview", "active_sessions", "journey_summary", "funnel", "heatmap", "retention"] },
          format: { enum: ["csv"] },
          frequency: { enum: ["daily", "weekly", "monthly"] },
          timezone: { bsonType: "string" },
          time_of_day: { bsonType: "string" },
          day_of_week: { bsonType: ["string", "null"] },
          day_of_month: { bsonType: ["int", "long", "null"], minimum: 1, maximum: 28 },
          status: { enum: ["active", "paused", "failed"] },
          recipients: { bsonType: "array", minItems: 1, maxItems: 20, items: { bsonType: "string" } },
          filters: { bsonType: ["object", "null"] },
          created_by: { bsonType: "string" },
          last_run_at: { bsonType: ["date", "null"] },
          next_run_at: { bsonType: ["date", "null"] },
          last_error: { bsonType: ["string", "null"] },
          created_at: { bsonType: "date" },
          updated_at: { bsonType: "date" }
        }
      }
    },
    validationLevel: "moderate",
    validationAction: "error"
  });
}

sessionDB.report_schedules.createIndex({ status: 1, next_run_at: 1 }, { name: "report_schedules_due" });
sessionDB.report_schedules.createIndex({ created_by: 1, created_at: -1 }, { name: "report_schedules_actor_created" });
sessionDB.report_schedules.createIndex({ updated_at: -1 }, { name: "report_schedules_updated" });
