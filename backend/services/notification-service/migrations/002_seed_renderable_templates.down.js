"use strict";

const databaseName = process.env.NOTIFICATION_MONGO_DATABASE || "notification_db";
const templatesCollectionName = process.env.NOTIFICATION_MONGO_TEMPLATES_COLLECTION || "notification_templates";
const notificationDB = db.getSiblingDB(databaseName);
const templates = notificationDB.getCollection(templatesCollectionName);

assert.commandWorked(templates.deleteMany({
  _id: {
    $in: [
      "tpl_otp_verification_email_v1",
      "tpl_otp_verification_sms_v1",
      "tpl_order_status_update_email_v1",
      "tpl_order_status_update_sms_v1",
      "tpl_order_status_update_push_v1",
      "tpl_payment_status_update_email_v1",
      "tpl_payment_status_update_sms_v1",
      "tpl_promotional_offer_email_v1",
      "tpl_promotional_offer_push_v1"
    ]
  }
}));
