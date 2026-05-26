const databaseName = "wishlist_db";
const collectionName = "wishlists";

const wishlistDB = db.getSiblingDB(databaseName);

if (wishlistDB.getCollectionNames().includes(collectionName)) {
  try {
    wishlistDB[collectionName].dropIndex("idx_wishlists_price_drop_candidates");
  } catch (err) {
    if (err.codeName !== "IndexNotFound") {
      throw err;
    }
  }
}
