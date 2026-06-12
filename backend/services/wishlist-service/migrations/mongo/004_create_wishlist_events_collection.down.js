const databaseName = "wishlist_db";
const collectionName = "wishlist_events";

const wishlistDB = db.getSiblingDB(databaseName);

if (wishlistDB.getCollectionNames().includes(collectionName)) {
  for (const indexName of [
    "ttl_wishlist_events_published_at",
    "idx_wishlist_events_type_time",
    "idx_wishlist_events_pending_retry",
  ]) {
    try {
      wishlistDB[collectionName].dropIndex(indexName);
    } catch (err) {
      if (err.codeName !== "IndexNotFound") {
        throw err;
      }
    }
  }

  wishlistDB.runCommand({
    collMod: collectionName,
    validator: {},
    validationLevel: "off",
    validationAction: "warn",
  });
}
