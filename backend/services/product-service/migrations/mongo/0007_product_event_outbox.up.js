const database = db.getSiblingDB("product_db");

const validator = {
  $jsonSchema: {
    bsonType: "object",
    required: [
      "_id",
      "topic",
      "event_type",
      "version",
      "source",
      "request_id",
      "trace_id",
      "payload",
      "status",
      "attempts",
      "next_attempt_at",
      "occurred_at"
    ],
    properties: {
      _id: { bsonType: "string" },
      topic: { bsonType: "string" },
      event_type: {
        enum: [
          "ProductCreated",
          "ProductUpdated",
          "ProductPublished",
          "ProductUnpublished",
          "ProductInventoryChanged"
        ]
      },
      version: { bsonType: "int", minimum: 1 },
      source: { bsonType: "string" },
      request_id: { bsonType: "string" },
      trace_id: { bsonType: "string" },
      payload: { bsonType: "object" },
      status: { enum: ["pending", "publishing", "published", "dead_lettered"] },
      attempts: { bsonType: "int", minimum: 0 },
      next_attempt_at: { bsonType: "date" },
      occurred_at: { bsonType: "date" },
      published_at: { bsonType: ["date", "null"] },
      last_error: { bsonType: ["string", "null"] }
    }
  }
};

if (database.getCollectionNames().includes("product_event_outbox")) {
  database.runCommand({
    collMod: "product_event_outbox",
    validator,
    validationLevel: "moderate",
    validationAction: "error"
  });
} else {
  database.createCollection("product_event_outbox", {
    validator,
    validationLevel: "moderate",
    validationAction: "error"
  });
}

database.product_event_outbox.createIndex(
  { status: 1, next_attempt_at: 1, occurred_at: 1 },
  { name: "idx_product_event_outbox_status_next_occurred" }
);
database.product_event_outbox.createIndex(
  { event_type: 1, occurred_at: -1 },
  { name: "idx_product_event_outbox_type_occurred" }
);
database.product_event_outbox.createIndex(
  { published_at: 1 },
  { expireAfterSeconds: 2592000, name: "ttl_product_event_outbox_published" }
);
