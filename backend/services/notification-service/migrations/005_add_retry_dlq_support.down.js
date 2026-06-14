"use strict";

const databaseName = process.env.NOTIFICATION_MONGO_DATABASE || "notification_db";
const deliveriesCollectionName = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
const notificationDB = db.getSiblingDB(databaseName);
const deliveries = notificationDB.getCollection(deliveriesCollectionName);

assert.commandWorked(deliveries.updateMany(
  { status: { $in: ["retry_scheduled", "dead_lettered"] } },
  {
    $set: { status: "failed", updated_at: new Date() },
    $unset: {
      recipient_ciphertext: "",
      max_attempts: "",
      last_failure_code: "",
      next_retry_at: "",
      dead_lettered_at: "",
      processing_attempt: "",
      processing_lease_until: ""
    }
  }
));

assert.commandWorked(deliveries.updateMany(
  {
    status: { $nin: ["retry_scheduled", "dead_lettered"] },
    max_attempts: { $exists: true }
  },
  {
    $unset: {
      recipient_ciphertext: "",
      max_attempts: "",
      last_failure_code: "",
      next_retry_at: "",
      dead_lettered_at: "",
      processing_attempt: "",
      processing_lease_until: ""
    }
  }
));

if (deliveries.getIndexes().some((index) => index.name === "delivery_retry_schedule_lookup")) {
  deliveries.dropIndex("delivery_retry_schedule_lookup");
}
if (deliveries.getIndexes().some((index) => index.name === "delivery_failed_operations_view")) {
  deliveries.dropIndex("delivery_failed_operations_view");
}

const eventAwareDeliveryValidator = {
  $jsonSchema: {
    bsonType: "object",
    required: ["_id", "channel", "template_key", "status", "attempts", "created_at", "updated_at"],
    properties: {
      _id: { bsonType: "string" },
      user_id: { bsonType: "string", minLength: 1 },
      channel: { enum: ["email", "sms", "push", "whatsapp_like"] },
      template_key: { bsonType: "string", minLength: 1 },
      status: { enum: ["processing", "pending", "accepted", "rejected", "failed", "sent"] },
      provider: { bsonType: "string" },
      provider_message_id: { bsonType: "string" },
      attempts: { bsonType: "int", minimum: 0 },
      payload: { bsonType: "object" },
      idempotency_key: { bsonType: "string", minLength: 1 },
      source_event_id: { bsonType: "string", minLength: 1 },
      source_event_type: { bsonType: "string", minLength: 1 },
      trace_id: { bsonType: "string" },
      created_at: { bsonType: "date" },
      updated_at: { bsonType: "date" }
    },
    oneOf: [
      {
        required: ["idempotency_key", "source_event_id", "source_event_type", "user_id"],
        properties: {
          template_key: { not: { enum: ["otp_verification"] } }
        }
      },
      {
        properties: {
          template_key: { enum: ["otp_verification"] }
        }
      },
      {
        required: ["user_id"],
        properties: {
          template_key: { not: { enum: ["otp_verification"] } },
          idempotency_key: { bsonType: "null" }
        }
      }
    ]
  }
};

assert.commandWorked(notificationDB.runCommand({
  collMod: deliveriesCollectionName,
  validator: eventAwareDeliveryValidator,
  validationLevel: "strict",
  validationAction: "error"
}));
