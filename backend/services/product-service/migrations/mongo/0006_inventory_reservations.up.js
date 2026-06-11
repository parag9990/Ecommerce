const database = db.getSiblingDB("product_db");

const validator = {
  $jsonSchema: {
    bsonType: "object",
    required: ["_id", "order_id", "status", "items", "expires_at", "created_at", "updated_at"],
    properties: {
      _id: { bsonType: "string" },
      order_id: { bsonType: "string" },
      idempotency_key: { bsonType: ["string", "null"] },
      status: { enum: ["reserved", "committed", "released", "expired"] },
      items: {
        bsonType: "array",
        minItems: 1,
        items: {
          bsonType: "object",
          required: ["product_id", "variant_id", "sku", "seller_id", "quantity"],
          properties: {
            product_id: { bsonType: "string" },
            variant_id: { bsonType: "string" },
            sku: { bsonType: "string" },
            seller_id: { bsonType: "string" },
            quantity: { bsonType: ["int", "long"], minimum: 1 }
          }
        }
      },
      expires_at: { bsonType: "date" },
      created_at: { bsonType: "date" },
      updated_at: { bsonType: "date" },
      committed_at: { bsonType: ["date", "null"] },
      released_at: { bsonType: ["date", "null"] },
      expired_at: { bsonType: ["date", "null"] },
      reason: { bsonType: ["string", "null"] }
    }
  }
};

if (database.getCollectionNames().includes("inventory_reservations")) {
  database.runCommand({
    collMod: "inventory_reservations",
    validator,
    validationLevel: "moderate",
    validationAction: "error"
  });
} else {
  database.createCollection("inventory_reservations", {
    validator,
    validationLevel: "moderate",
    validationAction: "error"
  });
}

database.inventory_reservations.createIndex(
  { order_id: 1 },
  { unique: true, name: "uq_inventory_reservations_order" }
);
database.inventory_reservations.createIndex(
  { idempotency_key: 1 },
  { unique: true, sparse: true, name: "uq_inventory_reservations_idempotency" }
);
database.inventory_reservations.createIndex(
  { status: 1, expires_at: 1 },
  { name: "idx_inventory_reservations_status_expires" }
);
database.inventory_reservations.createIndex(
  { expires_at: 1 },
  { expireAfterSeconds: 86400, name: "ttl_inventory_reservations_cleanup" }
);
