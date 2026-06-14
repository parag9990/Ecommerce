"use strict";

const databaseName = process.env.NOTIFICATION_MONGO_DATABASE || "notification_db";
const templatesCollectionName = process.env.NOTIFICATION_MONGO_TEMPLATES_COLLECTION || "notification_templates";
const deliveriesCollectionName = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
const notificationDB = db.getSiblingDB(databaseName);

function collectionExists(name) {
  return notificationDB.getCollectionInfos({ name: name }).length !== 0;
}

function dropIndexIfPresent(collection, name) {
  if (collection.getIndexes().some((index) => index.name === name)) {
    collection.dropIndex(name);
  }
}

if (collectionExists(templatesCollectionName)) {
  dropIndexIfPresent(notificationDB.getCollection(templatesCollectionName), "template_key_channel_version");
  assert.commandWorked(notificationDB.runCommand({
    collMod: templatesCollectionName,
    validator: {},
    validationLevel: "off"
  }));
}

if (collectionExists(deliveriesCollectionName)) {
  dropIndexIfPresent(notificationDB.getCollection(deliveriesCollectionName), "user_id_created_at");
  dropIndexIfPresent(notificationDB.getCollection(deliveriesCollectionName), "status_created_at");
  dropIndexIfPresent(notificationDB.getCollection(deliveriesCollectionName), "provider_message_id");
  assert.commandWorked(notificationDB.runCommand({
    collMod: deliveriesCollectionName,
    validator: {},
    validationLevel: "off"
  }));
}
