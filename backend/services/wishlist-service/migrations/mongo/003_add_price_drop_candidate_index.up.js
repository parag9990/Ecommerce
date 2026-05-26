const databaseName = "wishlist_db";
const collectionName = "wishlists";

const wishlistDB = db.getSiblingDB(databaseName);

if (wishlistDB.getCollectionNames().includes(collectionName)) {
  wishlistDB[collectionName].createIndex(
    {
      "items.product_id": 1,
      "items.variant_id": 1,
      "items.last_known_price.currency": 1,
      "items.last_known_price.amount": 1,
      "items.availability": 1,
    },
    { name: "idx_wishlists_price_drop_candidates" },
  );
}
