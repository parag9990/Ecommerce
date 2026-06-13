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

ensureCollection("journey_summaries", {
  $jsonSchema: {
    bsonType: "object",
    required: [
      "_id",
      "session_id",
      "anonymous_id",
      "duration_seconds",
      "total_events",
      "products_viewed",
      "searches",
      "cart_actions",
      "checkout_started",
      "payment_completed",
      "milestones",
      "top_paths",
      "schema_version",
      "calculated_at",
      "created_at",
      "updated_at",
    ],
    properties: {
      _id: { bsonType: "string" },
      session_id: { bsonType: "string" },
      anonymous_id: { bsonType: "string" },
      user_id: { bsonType: "string" },
      entry_page: { bsonType: "string" },
      exit_page: { bsonType: "string" },
      first_event_at: { bsonType: "date" },
      last_event_at: { bsonType: "date" },
      duration_seconds: { bsonType: ["long", "int"] },
      total_events: { bsonType: "int", minimum: 0 },
      products_viewed: { bsonType: "int", minimum: 0 },
      searches: { bsonType: "int", minimum: 0 },
      cart_actions: { bsonType: "int", minimum: 0 },
      checkout_started: { bsonType: "bool" },
      payment_completed: { bsonType: "bool" },
      milestones: {
        bsonType: "array",
        items: {
          bsonType: "object",
          required: ["name", "event_id", "occurred_at"],
          properties: {
            name: { bsonType: "string" },
            event_id: { bsonType: "string" },
            path: { bsonType: "string" },
            occurred_at: { bsonType: "date" },
          },
        },
      },
      top_paths: {
        bsonType: "array",
        items: {
          bsonType: "object",
          required: ["path", "count"],
          properties: {
            path: { bsonType: "string" },
            count: { bsonType: "int", minimum: 1 },
          },
        },
      },
      schema_version: { bsonType: "int" },
      calculated_at: { bsonType: "date" },
      created_at: { bsonType: "date" },
      updated_at: { bsonType: "date" },
    },
  },
});

sessionDB.journey_summaries.createIndex(
  { session_id: 1 },
  { unique: true, name: "uniq_journey_session" }
);

sessionDB.journey_summaries.createIndex(
  { user_id: 1, last_event_at: -1 },
  {
    name: "idx_journey_user_recent",
    partialFilterExpression: { user_id: { $type: "string" } },
  }
);

sessionDB.journey_summaries.createIndex(
  { last_event_at: -1 },
  { name: "idx_journey_recent" }
);
