const databaseName =
  typeof process !== "undefined" && process.env.RECOMMENDATION_MONGO_DATABASE
    ? process.env.RECOMMENDATION_MONGO_DATABASE
    : "recommendation_db";

const recommendationDB = db.getSiblingDB(databaseName);
const appliedAt = new Date();

recommendationDB.product_features.updateMany(
  {
    $or: [
      { status: { $exists: false } },
      { status: null },
      { status: "" },
    ],
  },
  {
    $set: {
      status: "active",
      "projection_defaults.status_defaulted": true,
      "projection_defaults.applied_at": appliedAt,
    },
  },
);

recommendationDB.product_features.updateMany(
  {
    $or: [
      { stock_status: { $exists: false } },
      { stock_status: null },
      { stock_status: "" },
    ],
  },
  {
    $set: {
      stock_status: "in_stock",
      "projection_defaults.stock_status_defaulted": true,
      "projection_defaults.applied_at": appliedAt,
    },
  },
);
