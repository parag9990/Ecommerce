const database = db.getSiblingDB("product_db");

const infos = database.getCollectionInfos({ name: "products" });
if (infos.length === 0) {
  throw new Error("products collection must exist before applying seller product workflow migration");
}

const validator = infos[0].options.validator || { $jsonSchema: { bsonType: "object", properties: {} } };
if (!validator.$jsonSchema) {
  validator.$jsonSchema = { bsonType: "object", properties: {} };
}

const schema = validator.$jsonSchema;
schema.properties = schema.properties || {};

schema.properties.status = schema.properties.status || {};
const statuses = schema.properties.status.enum || [];
if (!statuses.includes("rejected")) {
  statuses.push("rejected");
}
schema.properties.status.enum = statuses;

schema.properties.images = schema.properties.images || {
  bsonType: "array",
  items: { bsonType: "object", properties: {} }
};
schema.properties.images.items = schema.properties.images.items || {
  bsonType: "object",
  properties: {}
};
const imageItem = schema.properties.images.items;
imageItem.required = imageItem.required || [];
if (!imageItem.required.includes("image_id")) {
  imageItem.required.push("image_id");
}
imageItem.properties = imageItem.properties || {};
imageItem.properties.image_id = { bsonType: "string" };
imageItem.properties.alt_text = { bsonType: ["string", "null"] };
imageItem.properties.variant_ids = { bsonType: "array", items: { bsonType: "string" } };
imageItem.properties.width = { bsonType: ["int", "null"], minimum: 0 };
imageItem.properties.height = { bsonType: ["int", "null"], minimum: 0 };

database.runCommand({
  collMod: "products",
  validator,
  validationLevel: "moderate",
  validationAction: "error"
});
