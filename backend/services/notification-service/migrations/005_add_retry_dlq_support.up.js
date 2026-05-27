"use strict";

const databaseName = process.env.NOTIFICATION_MONGO_DATABASE || "notification_db";
const deliveriesCollectionName = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
const notificationDB = db.getSiblingDB(databaseName);
const deliveries = notificationDB.getCollection(deliveriesCollectionName);

const retryAwareDeliveryValidator = {
  $jsonSchema: {
    bsonType: "object",
    required: ["_id", "channel", "template_key", "status", "attempts", "created_at", "updated_at"],
    properties: {
      _id: { bsonType: "string" },
      user_id: { bsonType: "string", minLength: 1 },
      channel: { enum: ["email", "sms", "push", "whatsapp_like"] },
      template_key: { bsonType: "string", minLength: 1 },
      status: { enum: ["processing", "pending", "accepted", "rejected", "failed", "sent", "retry_scheduled", "dead_lettered"] },
      provider: { bsonType: "string" },
      provider_message_id: { bsonType: "string" },
      attempts: { bsonType: "int", minimum: 0 },
      max_attempts: { bsonType: "int", minimum: 1, maximum: 10 },
      payload: { bsonType: "object" },
      idempotency_key: { bsonType: "string", minLength: 1 },
      source_event_id: { bsonType: "string", minLength: 1 },
      source_event_type: { bsonType: "string", minLength: 1 },
      trace_id: { bsonType: "string" },
      recipient_ciphertext: { bsonType: "string", minLength: 1 },
      last_failure_code: { bsonType: "string", minLength: 1 },
      next_retry_at: { bsonType: "date" },
      dead_lettered_at: { bsonType: "date" },
      processing_attempt: { bsonType: "int", minimum: 1, maximum: 10 },
      processing_lease_until: { bsonType: "date" },
      created_at: { bsonType: "date" },
      updated_at: { bsonType: "date" }
    },
    oneOf: [
      {
        required: ["idempotency_key", "source_event_id", "source_event_type", "user_id", "recipient_ciphertext", "max_attempts"],
        properties: {
          template_key: { not: { enum: ["otp_verification"] } }
        }
      },
      {
        required: ["idempotency_key", "source_event_id", "source_event_type", "user_id"],
        properties: {
          template_key: { not: { enum: ["otp_verification"] } },
          status: { enum: ["processing", "accepted", "rejected", "failed"] },
          recipient_ciphertext: { bsonType: "null" },
          max_attempts: { bsonType: "null" }
        }
      },
      {
        properties: {
          template_key: { enum: ["otp_verification"] }
        },
        not: {
          anyOf: [
            { required: ["recipient_ciphertext"] },
            { required: ["max_attempts"] },
            { required: ["next_retry_at"] },
            { required: ["dead_lettered_at"] },
            { required: ["processing_lease_until"] }
          ]
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
  validator: retryAwareDeliveryValidator,
  validationLevel: "strict",
  validationAction: "error"
}));

deliveries.createIndex(
  { status: 1, next_retry_at: 1 },
  { name: "delivery_retry_schedule_lookup" }
);
deliveries.createIndex(
  { status: 1, updated_at: -1 },
  { name: "delivery_failed_operations_view" }
);
