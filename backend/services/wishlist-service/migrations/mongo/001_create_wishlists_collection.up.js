const databaseName = "wishlist_db";
const collectionName = "wishlists";

const wishlistDB = db.getSiblingDB(databaseName);

const wishlistValidator = {
  $jsonSchema: {
    bsonType: "object",
    title: "Wishlist",
    required: ["_id", "user_id", "visibility", "items", "created_at", "updated_at"],
    properties: {
      _id: {
        bsonType: "string",
        description: "Wishlist primary id, exposed as wishlist_id in APIs",
      },
      user_id: {
        bsonType: "string",
        description: "Authenticated buyer user id",
      },
      visibility: {
        enum: ["private"],
        description: "MVP supports private wishlists only",
      },
      items: {
        bsonType: "array",
        description: "Embedded wishlist items",
        items: {
          bsonType: "object",
          required: ["product_id", "added_at"],
          properties: {
            product_id: {
              bsonType: "string",
              description: "Saved product id",
            },
            variant_id: {
              bsonType: "string",
              description: "Optional selected product variant id",
            },
            added_at: {
              bsonType: "date",
              description: "When product was added to wishlist",
            },
            last_known_price: {
              bsonType: "object",
              required: ["amount", "currency"],
              properties: {
                amount: {
                  bsonType: "long",
                  description: "Price in minor unit",
                },
                currency: {
                  bsonType: "string",
                  pattern: "^[A-Z]{3}$",
                  description: "3-letter currency code",
                },
              },
            },
            availability: {
              enum: ["unknown", "in_stock", "out_of_stock", "deleted"],
              description: "Product availability snapshot",
            },
          },
        },
      },
      created_at: {
        bsonType: "date",
        description: "Wishlist creation timestamp",
      },
      updated_at: {
        bsonType: "date",
        description: "Wishlist last update timestamp",
      },
    },
  },
};

if (!wishlistDB.getCollectionNames().includes(collectionName)) {
  wishlistDB.createCollection(collectionName, {
    validator: wishlistValidator,
    validationLevel: "strict",
    validationAction: "error",
  });
} else {
  wishlistDB.runCommand({
    collMod: collectionName,
    validator: wishlistValidator,
    validationLevel: "strict",
    validationAction: "error",
  });
}

wishlistDB[collectionName].createIndex(
  { user_id: 1 },
  { unique: true, name: "uniq_wishlists_user_id" },
);

wishlistDB[collectionName].createIndex(
  { "items.product_id": 1 },
  { name: "idx_wishlists_items_product_id" },
);
