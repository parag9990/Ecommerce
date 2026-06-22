const databaseName = "wishlist_db";
const collectionName = "wishlist_events";

const wishlistDB = db.getSiblingDB(databaseName);

if (wishlistDB.getCollectionNames().includes(collectionName)) {
  try {
    wishlistDB[collectionName].dropIndex("idx_wishlist_events_claim_lease");
  } catch (err) {
    if (err.codeName !== "IndexNotFound") {
      throw err;
    }
  }
}
