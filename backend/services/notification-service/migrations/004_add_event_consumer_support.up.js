"use strict";

const databaseName = process.env.NOTIFICATION_MONGO_DATABASE || "notification_db";
const templatesCollectionName = process.env.NOTIFICATION_MONGO_TEMPLATES_COLLECTION || "notification_templates";
const deliveriesCollectionName = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
const notificationDB = db.getSiblingDB(databaseName);
const templates = notificationDB.getCollection(templatesCollectionName);
const publishedAt = new Date("2026-05-27T00:00:00.000Z");

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

notificationDB.getCollection(deliveriesCollectionName).createIndex(
  { idempotency_key: 1 },
  {
    name: "uniq_event_notification_key",
    unique: true,
    partialFilterExpression: { idempotency_key: { $type: "string" } }
  }
);

const eventTemplates = [
  {
    _id: "tpl_welcome_user_email_v1",
    template_key: "welcome_user",
    channel: "email",
    subject: "Welcome to Ecommerce",
    body: "Hi {{.name}}, welcome to Ecommerce. Your account is ready.",
    status: "active",
    version: NumberInt(1),
    created_at: publishedAt,
    updated_at: publishedAt
  },
  {
    _id: "tpl_seller_approved_email_v1",
    template_key: "seller_approved",
    channel: "email",
    subject: "Your seller account is approved",
    body: "Hi {{.name}}, your seller account has been approved.",
    status: "active",
    version: NumberInt(1),
    created_at: publishedAt,
    updated_at: publishedAt
  },
  {
    _id: "tpl_address_updated_security_notice_email_v1",
    template_key: "address_updated_security_notice",
    channel: "email",
    subject: "Your delivery address was updated",
    body: "Hi {{.name}}, a delivery address on your account was updated. If this was not you, contact support.",
    status: "active",
    version: NumberInt(1),
    created_at: publishedAt,
    updated_at: publishedAt
  }
];

for (const template of eventTemplates) {
  assert.commandWorked(templates.updateOne(
    {
      template_key: template.template_key,
      channel: template.channel,
      version: template.version
    },
    { $setOnInsert: template },
    { upsert: true }
  ));
}
