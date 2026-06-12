const databaseName = "wishlist_db";
const collectionName = "wishlists";

const wishlistDB = db.getSiblingDB(databaseName);

if (wishlistDB.getCollectionNames().includes(collectionName)) {
  try {
    wishlistDB[collectionName].dropIndex("idx_wishlists_items_product_id");
  } catch (err) {
    if (err.codeName !== "IndexNotFound") {
      throw err;
    }
  }

  try {
    wishlistDB[collectionName].dropIndex("uniq_wishlists_user_id");
  } catch (err) {
    if (err.codeName !== "IndexNotFound") {
      throw err;
    }
  }

  wishlistDB.runCommand({
    collMod: collectionName,
    validator: {},
    validationLevel: "off",
    validationAction: "warn",
  });
}
