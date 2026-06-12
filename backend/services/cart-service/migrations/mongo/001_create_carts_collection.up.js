const databaseName = "cart_db";
const collectionName = "carts";
const cartDB = db.getSiblingDB(databaseName);

const moneySchema = {
  bsonType: "object",
  required: ["amount", "currency"],
  additionalProperties: false,
  properties: {
    amount: {
      bsonType: ["int", "long"],
      minimum: 0
    },
    currency: {
      bsonType: "string",
      pattern: "^[A-Z]{3}$"
    }
  }
};

const cartValidator = {
  $jsonSchema: {
    bsonType: "object",
    required: [
      "_id",
      "status",
      "items",
      "totals",
      "version",
      "created_at",
      "updated_at",
      "expires_at"
    ],
    properties: {
      _id: {
        bsonType: "string",
        minLength: 1
      },
      user_id: {
        bsonType: ["string", "null"]
      },
      guest_session_id: {
        bsonType: ["string", "null"]
      },
      status: {
        enum: ["active", "merged", "checked_out", "expired", "abandoned"]
      },
      items: {
        bsonType: "array",
        maxItems: 100,
        items: {
          bsonType: "object",
          required: [
            "item_id",
            "product_id",
            "variant_id",
            "seller_id",
            "title_snapshot",
            "unit_price",
            "quantity",
            "line_subtotal",
            "price_snapshot_at",
            "added_at",
            "updated_at"
          ],
          properties: {
            item_id: { bsonType: "string", minLength: 1 },
            product_id: { bsonType: "string", minLength: 1 },
            variant_id: { bsonType: "string", minLength: 1 },
            seller_id: { bsonType: "string", minLength: 1 },
            sku_snapshot: { bsonType: ["string", "null"] },
            title_snapshot: { bsonType: "string", minLength: 1 },
            image_url_snapshot: { bsonType: ["string", "null"] },
            variant_snapshot: { bsonType: "object" },
            unit_price: moneySchema,
            quantity: {
              bsonType: ["int", "long"],
              minimum: 1,
              maximum: 10
            },
            line_subtotal: moneySchema,
            price_snapshot_at: { bsonType: "date" },
            added_at: { bsonType: "date" },
            updated_at: { bsonType: "date" }
          }
        }
      },
      coupon_code: {
        bsonType: ["string", "null"]
      },
      coupon_preview: {
        bsonType: ["object", "null"],
        properties: {
          coupon_id: { bsonType: "string" },
          code: { bsonType: "string" },
          valid: { bsonType: "bool" },
          discount: moneySchema,
          reason: { bsonType: "string" }
        }
      },
      totals: {
        bsonType: "object",
        required: ["subtotal", "discount", "total", "currency", "item_count", "unique_item_count"],
        additionalProperties: false,
        properties: {
          subtotal: moneySchema,
          discount: moneySchema,
          total: moneySchema,
          currency: {
            bsonType: "string",
            pattern: "^[A-Z]{3}$"
          },
          item_count: {
            bsonType: ["int", "long"],
            minimum: 0
          },
          unique_item_count: {
            bsonType: ["int", "long"],
            minimum: 0
          }
        }
      },
      version: {
        bsonType: ["int", "long"],
        minimum: 1
      },
      created_at: { bsonType: "date" },
      updated_at: { bsonType: "date" },
      expires_at: { bsonType: "date" },
      merged_into_cart_id: { bsonType: ["string", "null"] },
      checked_out_order_id: { bsonType: ["string", "null"] }
    }
  }
};

if (!cartDB.getCollectionNames().includes(collectionName)) {
  cartDB.createCollection(collectionName, {
    validator: cartValidator,
    validationLevel: "strict",
    validationAction: "error"
  });
} else {
  cartDB.runCommand({
    collMod: collectionName,
    validator: cartValidator,
    validationLevel: "strict",
    validationAction: "error"
  });
}

cartDB.carts.createIndex(
  { user_id: 1, status: 1 },
  { name: "idx_carts_user_status" }
);

cartDB.carts.createIndex(
  { guest_session_id: 1, status: 1 },
  { name: "idx_carts_guest_status" }
);

cartDB.carts.createIndex(
  { expires_at: 1 },
  { name: "idx_carts_expires_at_ttl", expireAfterSeconds: 0 }
);

cartDB.carts.createIndex(
  { user_id: 1 },
  {
    name: "uniq_active_cart_per_user",
    unique: true,
    partialFilterExpression: {
      status: "active",
      user_id: { $type: "string" }
    }
  }
);

cartDB.carts.createIndex(
  { guest_session_id: 1 },
  {
    name: "uniq_active_cart_per_guest_session",
    unique: true,
    partialFilterExpression: {
      status: "active",
      guest_session_id: { $type: "string" }
    }
  }
);

cartDB.carts.createIndex(
  { "items.product_id": 1, "items.variant_id": 1 },
  { name: "idx_carts_item_product_variant" }
);

cartDB.carts.createIndex(
  { updated_at: -1 },
  { name: "idx_carts_updated_at" }
);
