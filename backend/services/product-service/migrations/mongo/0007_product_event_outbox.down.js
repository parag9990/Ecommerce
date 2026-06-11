const database = db.getSiblingDB("product_db");

if (database.getCollectionNames().includes("product_event_outbox")) {
  const collection = database.product_event_outbox;
  const existingIndexes = new Set(collection.getIndexes().map((index) => index.name));
  for (const indexName of [
    "idx_product_event_outbox_status_next_occurred",
    "idx_product_event_outbox_type_occurred",
    "ttl_product_event_outbox_published"
  ]) {
    if (existingIndexes.has(indexName)) {
      collection.dropIndex(indexName);
    }
  }

  if (collection.countDocuments({}) === 0) {
    collection.drop();
  } else {
    database.runCommand({
      collMod: "product_event_outbox",
      validator: {},
      validationLevel: "off"
    });
    print("Retained non-empty product_event_outbox collection and removed Task 7 validator/indexes.");
  }
}
