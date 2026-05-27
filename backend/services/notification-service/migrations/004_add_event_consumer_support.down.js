"use strict";

const databaseName = process.env.NOTIFICATION_MONGO_DATABASE || "notification_db";
const templatesCollectionName = process.env.NOTIFICATION_MONGO_TEMPLATES_COLLECTION || "notification_templates";
const deliveriesCollectionName = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
const notificationDB = db.getSiblingDB(databaseName);
const deliveries = notificationDB.getCollection(deliveriesCollectionName);
const templates = notificationDB.getCollection(templatesCollectionName);

if (deliveries.getIndexes().some((index) => index.name === "uniq_event_notification_key")) {
  deliveries.dropIndex("uniq_event_notification_key");
}

assert.commandWorked(templates.deleteMany({
  _id: {
    $in: [
      "tpl_welcome_user_email_v1",
      "tpl_seller_approved_email_v1",
      "tpl_address_updated_security_notice_email_v1"
    ]
  }
}));

const otpAwareDeliveryValidator = {
  $jsonSchema: {
    bsonType: "object",
    required: ["_id", "channel", "template_key", "status", "attempts", "created_at", "updated_at"],
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
    },
    oneOf: [
      {
        properties: {
          template_key: { enum: ["otp_verification"] }
        }
      },
      {
        required: ["user_id"],
        properties: {
          template_key: { not: { enum: ["otp_verification"] } }
        }
      }
    ]
  }
};

assert.commandWorked(notificationDB.runCommand({
  collMod: deliveriesCollectionName,
  validator: otpAwareDeliveryValidator,
  validationLevel: "strict",
  validationAction: "error"
}));
