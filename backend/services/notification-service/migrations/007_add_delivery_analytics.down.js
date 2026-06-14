"use strict";

const databaseName = process.env.NOTIFICATION_MONGO_DATABASE || "notification_db";
const deliveriesCollectionName = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
const providerEventsCollectionName = process.env.NOTIFICATION_MONGO_PROVIDER_EVENTS_COLLECTION || "provider_events";
const notificationDB = db.getSiblingDB(databaseName);
const deliveries = notificationDB.getCollection(deliveriesCollectionName);

assert.commandWorked(deliveries.updateMany(
  {},
  {
    $unset: {
      campaign_id: "",
      sent_at: "",
      delivered_at: "",
      failed_at: "",
      opened_at: "",
      failure_code: ""
    }
  }
));

if (deliveries.getIndexes().some((index) => index.name === "campaign_delivery_range")) {
  deliveries.dropIndex("campaign_delivery_range");
}
if (notificationDB.getCollectionInfos({ name: providerEventsCollectionName }).length !== 0) {
  assert(notificationDB.getCollection(providerEventsCollectionName).drop());
}

const preferenceAwareDeliveryValidator = {
  $jsonSchema: {
    bsonType: "object",
    required: ["_id", "channel", "template_key", "status", "attempts", "created_at", "updated_at"],
    properties: {
      _id: { bsonType: "string" },
      user_id: { bsonType: "string", minLength: 1 },
      channel: { enum: ["email", "sms", "push", "whatsapp_like"] },
      template_key: { bsonType: "string", minLength: 1 },
      status: { enum: ["processing", "pending", "accepted", "rejected", "failed", "sent", "retry_scheduled", "dead_lettered", "suppressed"] },
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
      suppression_reason: {
        enum: [
          "security_channel_not_allowed",
          "channel_opted_out",
          "marketing_opted_out",
          "channel_consent_unavailable"
        ]
      },
      next_retry_at: { bsonType: "date" },
      dead_lettered_at: { bsonType: "date" },
      processing_attempt: { bsonType: "int", minimum: 1, maximum: 10 },
      processing_lease_until: { bsonType: "date" },
      created_at: { bsonType: "date" },
      updated_at: { bsonType: "date" }
    },
    oneOf: [
      {
        required: ["idempotency_key", "source_event_id", "source_event_type", "user_id", "suppression_reason"],
        properties: { status: { enum: ["suppressed"] } },
        not: {
          anyOf: [
            { required: ["recipient_ciphertext"] },
            { required: ["max_attempts"] },
            { required: ["next_retry_at"] },
            { required: ["processing_lease_until"] }
          ]
        }
      },
      {
        required: ["idempotency_key", "source_event_id", "source_event_type", "user_id", "recipient_ciphertext", "max_attempts"],
        properties: {
          template_key: { not: { enum: ["otp_verification"] } },
          status: { not: { enum: ["suppressed"] } }
        },
        not: { required: ["suppression_reason"] }
      },
      {
        required: ["idempotency_key", "source_event_id", "source_event_type", "user_id"],
        properties: {
          template_key: { not: { enum: ["otp_verification"] } },
          status: { enum: ["processing", "accepted", "rejected", "failed"] },
          recipient_ciphertext: { bsonType: "null" },
          max_attempts: { bsonType: "null" }
        },
        not: { required: ["suppression_reason"] }
      },
      {
        properties: { template_key: { enum: ["otp_verification"] } },
        not: {
          anyOf: [
            { required: ["recipient_ciphertext"] },
            { required: ["max_attempts"] },
            { required: ["next_retry_at"] },
            { required: ["dead_lettered_at"] },
            { required: ["processing_lease_until"] },
            { required: ["suppression_reason"] }
          ]
        }
      },
      {
        required: ["user_id"],
        properties: {
          template_key: { not: { enum: ["otp_verification"] } },
          idempotency_key: { bsonType: "null" }
        },
        not: { required: ["suppression_reason"] }
      }
    ]
  }
};

assert.commandWorked(notificationDB.runCommand({
  collMod: deliveriesCollectionName,
  validator: preferenceAwareDeliveryValidator,
  validationLevel: "strict",
  validationAction: "error"
}));
