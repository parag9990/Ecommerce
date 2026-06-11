const database = db.getSiblingDB("product_db");

const definitions = [
  {
    name: "price_books",
    indexes: [
      "idx_price_books_seller_status_starts",
      "idx_price_books_active_window_priority",
      "idx_price_books_entry_product_variant_status",
      "idx_price_books_entry_sku_status"
    ]
  },
  {
    name: "inventory_snapshots",
    indexes: [
      "idx_inventory_product_variant_created",
      "idx_inventory_seller_created",
      "idx_inventory_sku_created",
      "idx_inventory_type_created"
    ]
  },
  {
    name: "brands",
    indexes: [
      "uniq_brands_slug",
      "idx_brands_status_name",
      "idx_brands_name_text"
    ]
  },
  {
    name: "categories",
    indexes: [
      "idx_categories_parent_sort",
      "uniq_categories_slug",
      "idx_categories_path",
      "idx_categories_active_sort"
    ]
  },
  {
    name: "products",
    indexes: [
      "idx_products_seller_status_updated",
      "idx_products_category_status_updated",
      "idx_products_category_path_status_updated",
      "idx_products_brand_status_updated",
      "uniq_products_variant_sku",
      "uniq_products_slug",
      "idx_products_dev_text_search"
    ]
  }
];

for (const definition of definitions) {
  if (!database.getCollectionNames().includes(definition.name)) {
    continue;
  }

  const collection = database.getCollection(definition.name);
  const existingIndexes = new Set(collection.getIndexes().map((index) => index.name));
  for (const indexName of definition.indexes) {
    if (existingIndexes.has(indexName)) {
      collection.dropIndex(indexName);
    }
  }

  if (collection.countDocuments({}) === 0) {
    collection.drop();
    continue;
  }

  database.runCommand({
    collMod: definition.name,
    validator: {},
    validationLevel: "off"
  });
  print(`Retained non-empty ${definition.name} collection and removed Task 3 validator/indexes.`);
}
