const databaseName = "wishlist_db";
const collectionName = "wishlist_events";

const wishlistDB = db.getSiblingDB(databaseName);

const wishlistEventValidator = {
  $jsonSchema: {
    bsonType: "object",
    title: "WishlistAnalyticsEvent",
    required: [
      "_id",
      "event_type",
      "version",
      "topic",
      "payload",
      "status",
      "attempts",
      "next_retry_at",
      "occurred_at",
      "created_at",
      "updated_at",
    ],
    properties: {
      _id: {
        bsonType: "string",
        description: "Globally unique wishlist analytics event id",
      },
      event_type: {
        enum: ["wishlist_item_added", "wishlist_item_removed"],
        description: "Wishlist analytics event type",
      },
      version: {
        bsonType: "int",
        minimum: 1,
        description: "Event schema version",
      },
      topic: {
        bsonType: "string",
        description: "Destination message topic",
      },
      payload: {
        bsonType: "object",
        required: ["user_id", "product_id", "action", "source"],
        properties: {
          user_id: { bsonType: "string" },
          product_id: { bsonType: "string" },
          variant_id: { bsonType: "string" },
          action: { enum: ["add", "remove"] },
          source: { enum: ["wishlist"] },
          availability: {
            enum: ["unknown", "in_stock", "out_of_stock", "deleted"],
          },
          last_known_price: {
            bsonType: "object",
            required: ["amount", "currency"],
            properties: {
              amount: { bsonType: "long" },
              currency: {
                bsonType: "string",
                pattern: "^[A-Z]{3}$",
              },
            },
          },
        },
      },
      trace_id: { bsonType: "string" },
      status: {
        enum: ["pending", "publishing", "published", "failed"],
      },
      attempts: {
        bsonType: "int",
        minimum: 0,
      },
      next_retry_at: { bsonType: "date" },
      last_error: { bsonType: "string" },
      occurred_at: { bsonType: "date" },
      published_at: { bsonType: ["date", "null"] },
      locked_by: { bsonType: "string" },
      locked_until: { bsonType: ["date", "null"] },
      created_at: { bsonType: "date" },
      updated_at: { bsonType: "date" },
    },
  },
};

if (!wishlistDB.getCollectionNames().includes(collectionName)) {
  wishlistDB.createCollection(collectionName, {
    validator: wishlistEventValidator,
    validationLevel: "strict",
    validationAction: "error",
  });
} else {
  wishlistDB.runCommand({
    collMod: collectionName,
    validator: wishlistEventValidator,
    validationLevel: "strict",
    validationAction: "error",
  });
}

wishlistDB[collectionName].createIndex(
  { status: 1, next_retry_at: 1, created_at: 1 },
  { name: "idx_wishlist_events_pending_retry" },
);

wishlistDB[collectionName].createIndex(
  { event_type: 1, occurred_at: -1 },
  { name: "idx_wishlist_events_type_time" },
);

wishlistDB[collectionName].createIndex(
  { published_at: 1 },
  {
    name: "ttl_wishlist_events_published_at",
    expireAfterSeconds: 2592000,
  },
);
