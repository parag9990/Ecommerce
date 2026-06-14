"use strict";

const databaseName = process.env.NOTIFICATION_MONGO_DATABASE || "notification_db";
const templatesCollectionName = process.env.NOTIFICATION_MONGO_TEMPLATES_COLLECTION || "notification_templates";
const notificationDB = db.getSiblingDB(databaseName);
const templates = notificationDB.getCollection(templatesCollectionName);
const publishedAt = new Date("2026-05-27T00:00:00.000Z");

const seededTemplates = [
  {
    _id: "tpl_otp_verification_email_v1",
    template_key: "otp_verification",
    channel: "email",
    subject: "Your verification code",
    body: "Your verification code is {{.otp}}. It expires in {{.expires_in_minutes}} minutes.",
    status: "active",
    version: NumberInt(1),
    created_at: publishedAt,
    updated_at: publishedAt
  },
  {
    _id: "tpl_otp_verification_sms_v1",
    template_key: "otp_verification",
    channel: "sms",
    body: "Your code is {{.otp}}. Valid for {{.expires_in_minutes}} minutes.",
    status: "active",
    version: NumberInt(1),
    created_at: publishedAt,
    updated_at: publishedAt
  },
  {
    _id: "tpl_order_status_update_email_v1",
    template_key: "order_status_update",
    channel: "email",
    subject: "Order {{.order_id}} is {{.status}}",
    body: "Hi {{.name}}, your order {{.order_id}} is now {{.status}}.",
    status: "active",
    version: NumberInt(1),
    created_at: publishedAt,
    updated_at: publishedAt
  },
  {
    _id: "tpl_order_status_update_sms_v1",
    template_key: "order_status_update",
    channel: "sms",
    body: "Order {{.order_id}} is now {{.status}}.",
    status: "active",
    version: NumberInt(1),
    created_at: publishedAt,
    updated_at: publishedAt
  },
  {
    _id: "tpl_order_status_update_push_v1",
    template_key: "order_status_update",
    channel: "push",
    body: "Hi {{.name}}, order {{.order_id}} is now {{.status}}.",
    status: "active",
    version: NumberInt(1),
    created_at: publishedAt,
    updated_at: publishedAt
  },
  {
    _id: "tpl_payment_status_update_email_v1",
    template_key: "payment_status_update",
    channel: "email",
    subject: "Payment {{.payment_status}} for order {{.order_id}}",
    body: "Hi {{.name}}, payment of {{.amount}} for order {{.order_id}} is {{.payment_status}}.",
    status: "active",
    version: NumberInt(1),
    created_at: publishedAt,
    updated_at: publishedAt
  },
  {
    _id: "tpl_payment_status_update_sms_v1",
    template_key: "payment_status_update",
    channel: "sms",
    body: "Payment of {{.amount}} for order {{.order_id}} is {{.payment_status}}.",
    status: "active",
    version: NumberInt(1),
    created_at: publishedAt,
    updated_at: publishedAt
  },
  {
    _id: "tpl_promotional_offer_email_v1",
    template_key: "promotional_offer",
    channel: "email",
    subject: "{{.offer_title}} - use {{.coupon_code}}",
    body: "Hi {{.name}}, {{.offer_title}}! Use {{.coupon_code}} before {{.valid_until}}.",
    status: "active",
    version: NumberInt(1),
    created_at: publishedAt,
    updated_at: publishedAt
  },
  {
    _id: "tpl_promotional_offer_push_v1",
    template_key: "promotional_offer",
    channel: "push",
    body: "Hi {{.name}}, {{.offer_title}}! Use {{.coupon_code}} before {{.valid_until}}.",
    status: "active",
    version: NumberInt(1),
    created_at: publishedAt,
    updated_at: publishedAt
  }
];

for (const template of seededTemplates) {
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
