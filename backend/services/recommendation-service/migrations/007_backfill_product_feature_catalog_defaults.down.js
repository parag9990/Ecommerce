const databaseName =
  typeof process !== "undefined" && process.env.RECOMMENDATION_MONGO_DATABASE
    ? process.env.RECOMMENDATION_MONGO_DATABASE
    : "recommendation_db";

const recommendationDB = db.getSiblingDB(databaseName);

recommendationDB.product_features.updateMany(
  { "projection_defaults.status_defaulted": true },
  {
    $unset: {
      status: "",
      "projection_defaults.status_defaulted": "",
    },
  },
);

recommendationDB.product_features.updateMany(
  { "projection_defaults.stock_status_defaulted": true },
  {
    $unset: {
      stock_status: "",
      "projection_defaults.stock_status_defaulted": "",
    },
  },
);

recommendationDB.product_features.updateMany(
  {
    "projection_defaults.status_defaulted": { $exists: false },
    "projection_defaults.stock_status_defaulted": { $exists: false },
  },
  {
    $unset: {
      "projection_defaults.applied_at": "",
    },
  },
);
