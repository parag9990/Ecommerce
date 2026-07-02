const database = db.getSiblingDB("product_db");

database.categories.deleteMany({
  _id: {
    $in: [
      "cat_apparel",
      "cat_mens_shirts",
      "cat_womens_kurtas",
      "cat_footwear",
      "cat_sneakers",
      "cat_electronics",
      "cat_mobile_accessories",
      "cat_home_kitchen",
      "cat_home_decor"
    ]
  }
});
