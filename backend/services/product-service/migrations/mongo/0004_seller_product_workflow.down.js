const database = db.getSiblingDB("product_db");

const infos = database.getCollectionInfos({ name: "products" });
if (infos.length > 0) {
  const validator = infos[0].options.validator || {};
  if (validator.$jsonSchema && validator.$jsonSchema.properties) {
    const schema = validator.$jsonSchema;
    if (schema.properties.status && Array.isArray(schema.properties.status.enum)) {
      schema.properties.status.enum = schema.properties.status.enum.filter((status) => status !== "rejected");
    }

    const imageItem = schema.properties.images && schema.properties.images.items;
    if (imageItem) {
      if (Array.isArray(imageItem.required)) {
        imageItem.required = imageItem.required.filter((field) => field !== "image_id");
      }
      if (imageItem.properties) {
        delete imageItem.properties.image_id;
        delete imageItem.properties.alt_text;
        delete imageItem.properties.variant_ids;
        delete imageItem.properties.width;
        delete imageItem.properties.height;
      }
    }

    database.runCommand({
      collMod: "products",
      validator,
      validationLevel: "moderate",
      validationAction: "error"
    });
  }
}
