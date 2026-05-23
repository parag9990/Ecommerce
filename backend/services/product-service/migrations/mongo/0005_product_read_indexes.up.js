const database = db.getSiblingDB("product_db");

const productInfos = database.getCollectionInfos({ name: "products" });
if (productInfos.length === 0) {
  throw new Error("products collection must exist before applying product read indexes migration");
}

const productValidator = productInfos[0].options.validator || { $jsonSchema: { bsonType: "object", properties: {} } };
if (!productValidator.$jsonSchema) {
  productValidator.$jsonSchema = { bsonType: "object", properties: {} };
}
productValidator.$jsonSchema.properties = productValidator.$jsonSchema.properties || {};
productValidator.$jsonSchema.properties.category_path = {
  bsonType: "array",
  items: { bsonType: "string" }
};

database.runCommand({
  collMod: "products",
  validator: productValidator,
  validationLevel: "moderate",
  validationAction: "error"
});

database.products.createIndex(
  { status: 1, published_at: -1, _id: 1 },
  { name: "idx_products_status_published" }
);
database.products.createIndex(
  { category_id: 1, status: 1, published_at: -1, _id: 1 },
  { name: "idx_products_category_status_published" }
);
database.products.createIndex(
  { category_path: 1, status: 1, updated_at: -1 },
  { name: "idx_products_category_path_ids_status_updated" }
);
database.products.createIndex(
  { status: 1, "variants.price.amount": 1, _id: 1 },
  { name: "idx_products_status_price_amount" }
);
database.products.createIndex(
  { status: 1, "rating_summary.average": -1, "rating_summary.count": -1, _id: 1 },
  { name: "idx_products_status_rating" }
);

const categoryInfos = database.getCollectionInfos({ name: "categories" });
if (categoryInfos.length === 0) {
  throw new Error("categories collection must exist before applying product read indexes migration");
}

database.categories.createIndex(
  { is_active: 1, parent_id: 1, sort_order: 1, name: 1 },
  { name: "idx_categories_active_parent_sort_name" }
);
