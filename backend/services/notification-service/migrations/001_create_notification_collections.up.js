"use strict";

const databaseName = process.env.NOTIFICATION_MONGO_DATABASE || "notification_db";
const templatesCollectionName = process.env.NOTIFICATION_MONGO_TEMPLATES_COLLECTION || "notification_templates";
const deliveriesCollectionName = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
const notificationDB = db.getSiblingDB(databaseName);

const templateValidator = {
  $jsonSchema: {
    bsonType: "object",
    required: ["_id", "template_key", "channel", "body", "status", "version", "created_at", "updated_at"],
    properties: {
      _id: { bsonType: "string" },
      template_key: { bsonType: "string", minLength: 1 },
      channel: { enum: ["email", "sms", "push", "whatsapp_like"] },
      subject: { bsonType: "string" },
      body: { bsonType: "string", minLength: 1 },
      status: { enum: ["active", "inactive"] },
      version: { bsonType: "int", minimum: 1 },
      created_at: { bsonType: "date" },
      updated_at: { bsonType: "date" }
    },
    oneOf: [
      {
        required: ["subject"],
        properties: {
          channel: { enum: ["email"] },
          subject: { bsonType: "string", minLength: 1 }
        }
      },
      { properties: { channel: { enum: ["sms", "push", "whatsapp_like"] } } }
    ]
  }
};

const deliveryValidator = {
  $jsonSchema: {
    bsonType: "object",
    required: ["_id", "user_id", "channel", "template_key", "status", "attempts", "created_at", "updated_at"],
    properties: {
      _id: { bsonType: "string" },
      user_id: { bsonType: "string", minLength: 1 },
      channel: { enum: ["email", "sms", "push", "whatsapp_like"] },
      template_key: { bsonType: "string", minLength: 1 },
      status: { bsonType: "string", minLength: 1 },
      provider: { bsonType: "string" },
      provider_message_id: { bsonType: "string" },
      attempts: { bsonType: "int", minimum: 0 },
      payload: { bsonType: "object" },
      created_at: { bsonType: "date" },
      updated_at: { bsonType: "date" }
    }
  }
};

function applyCollectionValidator(name, validator) {
  if (notificationDB.getCollectionInfos({ name: name }).length === 0) {
    assert.commandWorked(notificationDB.createCollection(name, {
      validator: validator,
      validationLevel: "strict",
      validationAction: "error"
    }));
    return;
  }
  assert.commandWorked(notificationDB.runCommand({
    collMod: name,
    validator: validator,
    validationLevel: "strict",
    validationAction: "error"
  }));
}

applyCollectionValidator(templatesCollectionName, templateValidator);
applyCollectionValidator(deliveriesCollectionName, deliveryValidator);

notificationDB.getCollection(templatesCollectionName).createIndex(
  { template_key: 1, channel: 1, version: -1 },
  { name: "template_key_channel_version", unique: true }
);
notificationDB.getCollection(deliveriesCollectionName).createIndex(
  { user_id: 1, created_at: -1 },
  { name: "user_id_created_at" }
);
notificationDB.getCollection(deliveriesCollectionName).createIndex(
  { status: 1, created_at: 1 },
  { name: "status_created_at" }
);
notificationDB.getCollection(deliveriesCollectionName).createIndex(
  { provider: 1, provider_message_id: 1 },
  { name: "provider_message_id" }
);
