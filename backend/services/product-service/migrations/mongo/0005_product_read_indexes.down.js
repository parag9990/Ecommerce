const database = db.getSiblingDB("product_db");

function dropIndexIfExists(collection, name) {
  const exists = collection.getIndexes().some((index) => index.name === name);
  if (exists) {
    collection.dropIndex(name);
  }
}

dropIndexIfExists(database.products, "idx_products_status_published");
dropIndexIfExists(database.products, "idx_products_category_status_published");
dropIndexIfExists(database.products, "idx_products_category_path_ids_status_updated");
dropIndexIfExists(database.products, "idx_products_status_price_amount");
dropIndexIfExists(database.products, "idx_products_status_rating");
dropIndexIfExists(database.categories, "idx_categories_active_parent_sort_name");

const productInfos = database.getCollectionInfos({ name: "products" });
if (productInfos.length > 0) {
  const productValidator = productInfos[0].options.validator || {};
  if (productValidator.$jsonSchema) {
    productValidator.$jsonSchema.properties = productValidator.$jsonSchema.properties || {};
    productValidator.$jsonSchema.properties.category_path = {
      bsonType: "array",
      items: {
        bsonType: "object",
        required: ["category_id", "name", "slug"],
        properties: {
          category_id: { bsonType: "string" },
          name: { bsonType: "string" },
          slug: { bsonType: "string" }
        }
      }
    };

    database.runCommand({
      collMod: "products",
      validator: productValidator,
      validationLevel: "moderate",
      validationAction: "error"
    });
  }
}
