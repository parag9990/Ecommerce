"use strict";

const databaseName = process.env.NOTIFICATION_MONGO_DATABASE || "notification_db";
const deliveriesCollectionName = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
const providerEventsCollectionName = process.env.NOTIFICATION_MONGO_PROVIDER_EVENTS_COLLECTION || "provider_events";
const notificationDB = db.getSiblingDB(databaseName);
const deliveries = notificationDB.getCollection(deliveriesCollectionName);

const analyticsAwareDeliveryValidator = {
  $jsonSchema: {
    bsonType: "object",
    required: ["_id", "channel", "template_key", "status", "attempts", "created_at", "updated_at"],
    properties: {
      _id: { bsonType: "string" },
      user_id: { bsonType: "string", minLength: 1 },
      channel: { enum: ["email", "sms", "push", "whatsapp_like"] },
      template_key: { bsonType: "string", minLength: 1 },
      campaign_id: { bsonType: "string", minLength: 1 },
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
      sent_at: { bsonType: "date" },
      delivered_at: { bsonType: "date" },
      failed_at: { bsonType: "date" },
      opened_at: { bsonType: "date" },
      failure_code: { bsonType: "string", minLength: 1 },
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
  validator: analyticsAwareDeliveryValidator,
  validationLevel: "strict",
  validationAction: "error"
}));

const providerEventValidator = {
  $jsonSchema: {
    bsonType: "object",
    required: [
      "_id", "provider", "provider_event_id", "delivery_id", "type",
      "channel", "template_key", "occurred_at", "received_at"
    ],
    properties: {
      _id: { bsonType: "string", minLength: 1 },
      provider: { bsonType: "string", minLength: 1 },
      provider_event_id: { bsonType: "string", minLength: 1 },
      provider_message_id: { bsonType: "string", minLength: 1 },
      delivery_id: { bsonType: "string", minLength: 1 },
      type: { enum: ["sent", "delivered", "failed", "opened"] },
      channel: { enum: ["email", "sms", "push", "whatsapp_like"] },
      template_key: { bsonType: "string", minLength: 1 },
      campaign_id: { bsonType: "string", minLength: 1 },
      occurred_at: { bsonType: "date" },
      received_at: { bsonType: "date" },
      failure_code: { bsonType: "string", minLength: 1 }
    }
  }
};

if (notificationDB.getCollectionInfos({ name: providerEventsCollectionName }).length === 0) {
  assert.commandWorked(notificationDB.createCollection(providerEventsCollectionName, {
    validator: providerEventValidator,
    validationLevel: "strict",
    validationAction: "error"
  }));
} else {
  assert.commandWorked(notificationDB.runCommand({
    collMod: providerEventsCollectionName,
    validator: providerEventValidator,
    validationLevel: "strict",
    validationAction: "error"
  }));
}

notificationDB.getCollection(providerEventsCollectionName).createIndex(
  { provider: 1, provider_event_id: 1 },
  { name: "provider_event_dedup", unique: true }
);
notificationDB.getCollection(providerEventsCollectionName).createIndex(
  { delivery_id: 1, type: 1, occurred_at: 1 },
  { name: "delivery_event_timeline" }
);
notificationDB.getCollection(providerEventsCollectionName).createIndex(
  { template_key: 1, channel: 1, type: 1, occurred_at: 1 },
  { name: "template_channel_quality" }
);
deliveries.createIndex(
  { campaign_id: 1, created_at: 1 },
  { name: "campaign_delivery_range", sparse: true }
);
