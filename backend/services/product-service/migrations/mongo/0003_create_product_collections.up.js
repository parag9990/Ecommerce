const database = db.getSiblingDB("product_db");

const moneySchema = {
  bsonType: "object",
  required: ["amount", "currency"],
  properties: {
    amount: { bsonType: ["int", "long"], minimum: 1 },
    currency: { bsonType: "string" }
  }
};

const definitions = [
  {
    name: "products",
    validator: {
      $jsonSchema: {
        bsonType: "object",
        required: ["_id", "seller_id", "title", "category_id", "status", "variants", "created_at", "updated_at"],
        properties: {
          _id: { bsonType: "string" },
          seller_id: { bsonType: "string" },
          title: { bsonType: "string", minLength: 1, maxLength: 200 },
          slug: { bsonType: "string" },
          description: { bsonType: ["string", "null"] },
          brand_id: { bsonType: ["string", "null"] },
          brand_name: { bsonType: ["string", "null"] },
          category_id: { bsonType: "string" },
          category_path: {
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
          },
          status: { enum: ["draft", "submitted", "published", "unpublished", "archived"] },
          attributes: { bsonType: "object" },
          images: {
            bsonType: "array",
            items: {
              bsonType: "object",
              required: ["url", "position", "is_primary", "status"],
              properties: {
                url: { bsonType: "string" },
                alt: { bsonType: ["string", "null"] },
                position: { bsonType: "int", minimum: 1 },
                is_primary: { bsonType: "bool" },
                status: { enum: ["active", "hidden", "processing", "failed"] }
              }
            }
          },
          variants: {
            bsonType: "array",
            minItems: 1,
            items: {
              bsonType: "object",
              required: ["variant_id", "sku", "price", "stock_quantity", "reserved_quantity", "status"],
              properties: {
                variant_id: { bsonType: "string" },
                sku: { bsonType: "string" },
                title: { bsonType: ["string", "null"] },
                attributes: { bsonType: "object" },
                price: moneySchema,
                mrp: { ...moneySchema, bsonType: ["object", "null"] },
                stock_quantity: { bsonType: ["int", "long"], minimum: 0 },
                reserved_quantity: { bsonType: ["int", "long"], minimum: 0 },
                safety_stock: { bsonType: ["int", "long", "null"], minimum: 0 },
                status: { enum: ["active", "inactive", "out_of_stock", "deleted"] }
              }
            }
          },
          rating_summary: {
            bsonType: "object",
            properties: {
              average: { bsonType: ["double", "int", "long"], minimum: 0, maximum: 5 },
              count: { bsonType: ["int", "long"], minimum: 0 }
            }
          },
          created_by: { bsonType: ["string", "null"] },
          updated_by: { bsonType: ["string", "null"] },
          created_at: { bsonType: "date" },
          updated_at: { bsonType: "date" },
          published_at: { bsonType: ["date", "null"] }
        }
      }
    },
    indexes: [
      { key: { seller_id: 1, status: 1, updated_at: -1 }, name: "idx_products_seller_status_updated" },
      { key: { category_id: 1, status: 1, updated_at: -1 }, name: "idx_products_category_status_updated" },
      { key: { "category_path.category_id": 1, status: 1, updated_at: -1 }, name: "idx_products_category_path_status_updated" },
      { key: { brand_id: 1, status: 1, updated_at: -1 }, name: "idx_products_brand_status_updated" },
      { key: { "variants.sku": 1 }, name: "uniq_products_variant_sku", unique: true },
      { key: { slug: 1 }, name: "uniq_products_slug", unique: true, sparse: true },
      { key: { title: "text", description: "text", brand_name: "text" }, name: "idx_products_dev_text_search" }
    ]
  },
  {
    name: "categories",
    validator: {
      $jsonSchema: {
        bsonType: "object",
        required: ["_id", "name", "slug", "path", "depth", "sort_order", "is_active", "created_at", "updated_at"],
        properties: {
          _id: { bsonType: "string" },
          name: { bsonType: "string", minLength: 1, maxLength: 120 },
          slug: { bsonType: "string" },
          parent_id: { bsonType: ["string", "null"] },
          path: { bsonType: "array", items: { bsonType: "string" } },
          depth: { bsonType: "int", minimum: 0 },
          sort_order: { bsonType: "int" },
          is_active: { bsonType: "bool" },
          attribute_schema: {
            bsonType: "array",
            items: {
              bsonType: "object",
              required: ["key", "label", "type", "required", "filterable", "variant_axis"],
              properties: {
                key: { bsonType: "string" },
                label: { bsonType: "string" },
                type: { bsonType: "string" },
                required: { bsonType: "bool" },
                filterable: { bsonType: "bool" },
                variant_axis: { bsonType: "bool" },
                allowed_values: { bsonType: "array", items: { bsonType: "string" } },
                unit: { bsonType: ["string", "null"] }
              }
            }
          },
          seo: {
            bsonType: ["object", "null"],
            properties: {
              title: { bsonType: ["string", "null"] },
              description: { bsonType: ["string", "null"] }
            }
          },
          created_at: { bsonType: "date" },
          updated_at: { bsonType: "date" }
        }
      }
    },
    indexes: [
      { key: { parent_id: 1, sort_order: 1 }, name: "idx_categories_parent_sort" },
      { key: { slug: 1 }, name: "uniq_categories_slug", unique: true },
      { key: { path: 1 }, name: "idx_categories_path" },
      { key: { is_active: 1, sort_order: 1 }, name: "idx_categories_active_sort" }
    ]
  },
  {
    name: "brands",
    validator: {
      $jsonSchema: {
        bsonType: "object",
        required: ["_id", "name", "slug", "status", "created_at", "updated_at"],
        properties: {
          _id: { bsonType: "string" },
          name: { bsonType: "string", minLength: 1, maxLength: 120 },
          slug: { bsonType: "string" },
          description: { bsonType: ["string", "null"] },
          logo_url: { bsonType: ["string", "null"] },
          website_url: { bsonType: ["string", "null"] },
          status: { enum: ["active", "inactive", "archived"] },
          created_by: { bsonType: ["string", "null"] },
          created_at: { bsonType: "date" },
          updated_at: { bsonType: "date" }
        }
      }
    },
    indexes: [
      { key: { slug: 1 }, name: "uniq_brands_slug", unique: true },
      { key: { status: 1, name: 1 }, name: "idx_brands_status_name" },
      { key: { name: "text" }, name: "idx_brands_name_text" }
    ]
  },
  {
    name: "inventory_snapshots",
    validator: {
      $jsonSchema: {
        bsonType: "object",
        required: ["_id", "product_id", "variant_id", "sku", "seller_id", "snapshot_type", "stock_quantity", "reserved_quantity", "available_quantity", "created_at"],
        properties: {
          _id: { bsonType: "string" },
          product_id: { bsonType: "string" },
          variant_id: { bsonType: "string" },
          sku: { bsonType: "string" },
          seller_id: { bsonType: "string" },
          snapshot_type: { enum: ["periodic", "manual_adjustment", "reservation", "release", "order_commit", "correction"] },
          stock_quantity: { bsonType: ["int", "long"], minimum: 0 },
          reserved_quantity: { bsonType: ["int", "long"], minimum: 0 },
          safety_stock: { bsonType: ["int", "long", "null"], minimum: 0 },
          available_quantity: { bsonType: ["int", "long"] },
          reason: { bsonType: ["string", "null"] },
          reference: {
            bsonType: ["object", "null"],
            properties: {
              type: { bsonType: ["string", "null"] },
              id: { bsonType: ["string", "null"] }
            }
          },
          created_at: { bsonType: "date" }
        }
      }
    },
    indexes: [
      { key: { product_id: 1, variant_id: 1, created_at: -1 }, name: "idx_inventory_product_variant_created" },
      { key: { seller_id: 1, created_at: -1 }, name: "idx_inventory_seller_created" },
      { key: { sku: 1, created_at: -1 }, name: "idx_inventory_sku_created" },
      { key: { snapshot_type: 1, created_at: -1 }, name: "idx_inventory_type_created" }
    ]
  },
  {
    name: "price_books",
    validator: {
      $jsonSchema: {
        bsonType: "object",
        required: ["_id", "seller_id", "name", "currency", "status", "priority", "entries", "created_at", "updated_at"],
        properties: {
          _id: { bsonType: "string" },
          seller_id: { bsonType: "string" },
          name: { bsonType: "string", minLength: 1, maxLength: 160 },
          currency: { bsonType: "string" },
          status: { enum: ["draft", "active", "expired", "archived"] },
          priority: { bsonType: "int" },
          starts_at: { bsonType: ["date", "null"] },
          ends_at: { bsonType: ["date", "null"] },
          entries: {
            bsonType: "array",
            items: {
              bsonType: "object",
              required: ["entry_id", "product_id", "variant_id", "sku", "price"],
              properties: {
                entry_id: { bsonType: "string" },
                product_id: { bsonType: "string" },
                variant_id: { bsonType: "string" },
                sku: { bsonType: "string" },
                price: moneySchema,
                mrp: { ...moneySchema, bsonType: ["object", "null"] },
                min_quantity: { bsonType: ["int", "long", "null"], minimum: 1 }
              }
            }
          },
          created_by: { bsonType: ["string", "null"] },
          updated_by: { bsonType: ["string", "null"] },
          created_at: { bsonType: "date" },
          updated_at: { bsonType: "date" }
        }
      }
    },
    indexes: [
      { key: { seller_id: 1, status: 1, starts_at: -1 }, name: "idx_price_books_seller_status_starts" },
      { key: { status: 1, starts_at: 1, ends_at: 1, priority: -1 }, name: "idx_price_books_active_window_priority" },
      { key: { "entries.product_id": 1, "entries.variant_id": 1, status: 1 }, name: "idx_price_books_entry_product_variant_status" },
      { key: { "entries.sku": 1, status: 1 }, name: "idx_price_books_entry_sku_status" }
    ]
  }
];

const existingNames = new Set(database.getCollectionNames());

for (const definition of definitions) {
  if (existingNames.has(definition.name)) {
    database.runCommand({
      collMod: definition.name,
      validator: definition.validator,
      validationLevel: "moderate",
      validationAction: "error"
    });
  } else {
    database.createCollection(definition.name, {
      validator: definition.validator,
      validationLevel: "moderate",
      validationAction: "error"
    });
  }

  for (const index of definition.indexes) {
    const { key, ...options } = index;
    database.getCollection(definition.name).createIndex(key, options);
  }
}
