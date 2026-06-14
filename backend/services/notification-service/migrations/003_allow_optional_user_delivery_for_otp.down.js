"use strict";

const databaseName = process.env.NOTIFICATION_MONGO_DATABASE || "notification_db";
const deliveriesCollectionName = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
const notificationDB = db.getSiblingDB(databaseName);

const requiredUserDeliveryValidator = {
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

assert.commandWorked(notificationDB.runCommand({
  collMod: deliveriesCollectionName,
  validator: requiredUserDeliveryValidator,
  validationLevel: "strict",
  validationAction: "error"
}));
