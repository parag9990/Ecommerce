const database = db.getSiblingDB("product_db");

if (!database.getCollectionNames().includes("categories")) {
  throw new Error("categories collection must exist before seeding local categories");
}

const now = new Date();

function category(id, name, slug, parentId, path, depth, sortOrder) {
  return {
    _id: id,
    name,
    slug,
    parent_id: parentId,
    path,
    depth: NumberInt(depth),
    level: NumberInt(depth),
    sort_order: NumberInt(sortOrder),
    is_active: true,
    attribute_schema: [],
    seo: null,
    created_at: now,
    updated_at: now
  };
}

const categories = [
  category("cat_apparel", "Apparel", "apparel", null, ["cat_apparel"], 0, 10),
  category("cat_mens_shirts", "Men's Shirts", "mens-shirts", "cat_apparel", ["cat_apparel", "cat_mens_shirts"], 1, 11),
  category("cat_womens_kurtas", "Women's Kurtas", "womens-kurtas", "cat_apparel", ["cat_apparel", "cat_womens_kurtas"], 1, 12),
  category("cat_footwear", "Footwear", "footwear", null, ["cat_footwear"], 0, 20),
  category("cat_sneakers", "Sneakers", "sneakers", "cat_footwear", ["cat_footwear", "cat_sneakers"], 1, 21),
  category("cat_electronics", "Electronics", "electronics", null, ["cat_electronics"], 0, 30),
  category("cat_mobile_accessories", "Mobile Accessories", "mobile-accessories", "cat_electronics", ["cat_electronics", "cat_mobile_accessories"], 1, 31),
  category("cat_home_kitchen", "Home & Kitchen", "home-kitchen", null, ["cat_home_kitchen"], 0, 40),
  category("cat_home_decor", "Home Decor", "home-decor", "cat_home_kitchen", ["cat_home_kitchen", "cat_home_decor"], 1, 41)
];

database.categories.bulkWrite(
  categories.map((item) => ({
    updateOne: {
      filter: { _id: item._id },
      update: {
        $setOnInsert: {
          _id: item._id,
          created_at: item.created_at
        },
        $set: {
          name: item.name,
          slug: item.slug,
          parent_id: item.parent_id,
          path: item.path,
          depth: item.depth,
          level: item.level,
          sort_order: item.sort_order,
          is_active: item.is_active,
          attribute_schema: item.attribute_schema,
          seo: item.seo,
          updated_at: item.updated_at
        }
      },
      upsert: true
    }
  })),
  { ordered: true }
);
