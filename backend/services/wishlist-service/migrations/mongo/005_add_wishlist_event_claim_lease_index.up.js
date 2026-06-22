const databaseName = "wishlist_db";
const collectionName = "wishlist_events";

const wishlistDB = db.getSiblingDB(databaseName);

if (wishlistDB.getCollectionNames().includes(collectionName)) {
  wishlistDB[collectionName].createIndex(
    { status: 1, locked_until: 1, next_retry_at: 1, created_at: 1 },
    { name: "idx_wishlist_events_claim_lease" },
  );
}
