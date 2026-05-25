const databaseName = "cart_db";
const collectionName = "carts";
const cartDB = db.getSiblingDB(databaseName);

if (cartDB.getCollectionNames().includes(collectionName)) {
  const indexNames = [
    "idx_carts_user_status",
    "idx_carts_guest_status",
    "idx_carts_expires_at_ttl",
    "uniq_active_cart_per_user",
    "uniq_active_cart_per_guest_session",
    "idx_carts_item_product_variant",
    "idx_carts_updated_at"
  ];

  for (const indexName of indexNames) {
    try {
      cartDB.carts.dropIndex(indexName);
    } catch (error) {
      if (error.codeName !== "IndexNotFound") {
        throw error;
      }
    }
  }

  cartDB.runCommand({
    collMod: collectionName,
    validator: {},
    validationLevel: "off"
  });
}
