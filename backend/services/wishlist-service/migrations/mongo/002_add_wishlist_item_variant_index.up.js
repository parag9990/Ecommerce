const databaseName = "wishlist_db";
const collectionName = "wishlists";

const wishlistDB = db.getSiblingDB(databaseName);

if (wishlistDB.getCollectionNames().includes(collectionName)) {
  wishlistDB[collectionName].createIndex(
    { "items.product_id": 1, "items.variant_id": 1 },
    { name: "idx_wishlists_items_product_variant" },
  );
}
