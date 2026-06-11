const database = db.getSiblingDB("product_db");

if (database.getCollectionNames().includes("inventory_reservations")) {
  const collection = database.inventory_reservations;
  const existingIndexes = new Set(collection.getIndexes().map((index) => index.name));
  for (const indexName of [
    "uq_inventory_reservations_order",
    "uq_inventory_reservations_idempotency",
    "idx_inventory_reservations_status_expires",
    "ttl_inventory_reservations_cleanup"
  ]) {
    if (existingIndexes.has(indexName)) {
      collection.dropIndex(indexName);
    }
  }

  if (collection.countDocuments({}) === 0) {
    collection.drop();
  } else {
    database.runCommand({
      collMod: "inventory_reservations",
      validator: {},
      validationLevel: "off"
    });
    print("Retained non-empty inventory_reservations collection and removed Task 6 validator/indexes.");
  }
}
