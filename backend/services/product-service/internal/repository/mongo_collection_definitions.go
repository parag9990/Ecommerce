package repository

import (
	"product-service/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoCollectionDefinition struct {
	Name      string
	Validator bson.D
	Indexes   []MongoIndexDefinition
}

type MongoIndexDefinition struct {
	Name  string
	Model mongo.IndexModel
}

func ProductMongoCollectionDefinitions() []MongoCollectionDefinition {
	return []MongoCollectionDefinition{
		productsCollectionDefinition(),
		categoriesCollectionDefinition(),
		brandsCollectionDefinition(),
		inventoryReservationsCollectionDefinition(),
		inventorySnapshotsCollectionDefinition(),
		priceBooksCollectionDefinition(),
		productEventOutboxCollectionDefinition(),
	}
}

func (d MongoCollectionDefinition) IndexModels() []mongo.IndexModel {
	models := make([]mongo.IndexModel, 0, len(d.Indexes))
	for _, index := range d.Indexes {
		models = append(models, index.Model)
	}
	return models
}

func (d MongoCollectionDefinition) IndexNames() []string {
	names := make([]string, 0, len(d.Indexes))
	for _, index := range d.Indexes {
		names = append(names, index.Name)
	}
	return names
}

func DescribeProductMongoCollections() []CollectionDescription {
	definitions := ProductMongoCollectionDefinitions()
	descriptions := make([]CollectionDescription, 0, len(definitions))
	for _, definition := range definitions {
		descriptions = append(descriptions, CollectionDescription{
			Name:         definition.Name,
			IndexNames:   definition.IndexNames(),
			HasValidator: len(definition.Validator) > 0,
		})
	}
	return descriptions
}

func productsCollectionDefinition() MongoCollectionDefinition {
	return MongoCollectionDefinition{
		Name: CollectionProducts,
		Validator: jsonSchema(
			bson.A{"_id", "seller_id", "title", "category_id", "status", "variants", "created_at", "updated_at"},
			bson.D{
				prop("_id", bson.D{e("bsonType", "string")}),
				prop("seller_id", bson.D{e("bsonType", "string")}),
				prop("title", bson.D{e("bsonType", "string"), e("minLength", 1), e("maxLength", 200)}),
				prop("slug", bson.D{e("bsonType", "string")}),
				prop("description", nullable("string")),
				prop("brand_id", nullable("string")),
				prop("brand_name", nullable("string")),
				prop("category_id", bson.D{e("bsonType", "string")}),
				prop("category_path", bson.D{
					e("bsonType", "array"),
					e("items", bson.D{
						e("bsonType", "string"),
					}),
				}),
				prop("status", enum("draft", "submitted", "rejected", "published", "unpublished", "archived")),
				prop("attributes", bson.D{e("bsonType", "object")}),
				prop("images", bson.D{
					e("bsonType", "array"),
					e("items", bson.D{
						e("bsonType", "object"),
						e("required", bson.A{"image_id", "url", "position", "is_primary", "status"}),
						e("properties", bson.D{
							prop("image_id", bson.D{e("bsonType", "string")}),
							prop("url", bson.D{e("bsonType", "string")}),
							prop("alt", nullable("string")),
							prop("alt_text", nullable("string")),
							prop("position", bson.D{e("bsonType", "int"), e("minimum", 1)}),
							prop("is_primary", bson.D{e("bsonType", "bool")}),
							prop("variant_ids", bson.D{e("bsonType", "array"), e("items", bson.D{e("bsonType", "string")})}),
							prop("width", bson.D{e("bsonType", bson.A{"int", "null"}), e("minimum", 0)}),
							prop("height", bson.D{e("bsonType", bson.A{"int", "null"}), e("minimum", 0)}),
							prop("status", enum("active", "hidden", "processing", "failed")),
						}),
					}),
				}),
				prop("variants", bson.D{
					e("bsonType", "array"),
					e("minItems", 1),
					e("items", bson.D{
						e("bsonType", "object"),
						e("required", bson.A{"variant_id", "sku", "price", "stock_quantity", "reserved_quantity", "status"}),
						e("properties", bson.D{
							prop("variant_id", bson.D{e("bsonType", "string")}),
							prop("sku", bson.D{e("bsonType", "string")}),
							prop("title", nullable("string")),
							prop("attributes", bson.D{e("bsonType", "object")}),
							prop("price", moneySchema()),
							prop("mrp", nullableObject(moneySchema())),
							prop("stock_quantity", bson.D{e("bsonType", bson.A{"int", "long"}), e("minimum", 0)}),
							prop("reserved_quantity", bson.D{e("bsonType", bson.A{"int", "long"}), e("minimum", 0)}),
							prop("safety_stock", bson.D{e("bsonType", bson.A{"int", "long", "null"}), e("minimum", 0)}),
							prop("status", enum("active", "inactive", "out_of_stock", "deleted")),
						}),
					}),
				}),
				prop("rating_summary", bson.D{
					e("bsonType", "object"),
					e("properties", bson.D{
						prop("average", bson.D{e("bsonType", bson.A{"double", "int", "long"}), e("minimum", 0), e("maximum", 5)}),
						prop("count", bson.D{e("bsonType", bson.A{"int", "long"}), e("minimum", 0)}),
					}),
				}),
				prop("created_by", nullable("string")),
				prop("updated_by", nullable("string")),
				prop("created_at", bson.D{e("bsonType", "date")}),
				prop("updated_at", bson.D{e("bsonType", "date")}),
				prop("published_at", bson.D{e("bsonType", bson.A{"date", "null"})}),
			},
		),
		Indexes: []MongoIndexDefinition{
			index("idx_products_seller_status_updated", bson.D{e("seller_id", 1), e("status", 1), e("updated_at", -1)}),
			index("idx_products_status_published", bson.D{e("status", 1), e("published_at", -1), e("_id", 1)}),
			index("idx_products_category_status_updated", bson.D{e("category_id", 1), e("status", 1), e("updated_at", -1)}),
			index("idx_products_category_status_published", bson.D{e("category_id", 1), e("status", 1), e("published_at", -1), e("_id", 1)}),
			index("idx_products_category_path_status_updated", bson.D{e("category_path.category_id", 1), e("status", 1), e("updated_at", -1)}),
			index("idx_products_category_path_ids_status_updated", bson.D{e("category_path", 1), e("status", 1), e("updated_at", -1)}),
			index("idx_products_brand_status_updated", bson.D{e("brand_id", 1), e("status", 1), e("updated_at", -1)}),
			index("idx_products_status_price_amount", bson.D{e("status", 1), e("variants.price.amount", 1), e("_id", 1)}),
			index("idx_products_status_rating", bson.D{e("status", 1), e("rating_summary.average", -1), e("rating_summary.count", -1), e("_id", 1)}),
			uniqueIndex("uniq_products_variant_sku", bson.D{e("variants.sku", 1)}),
			sparseUniqueIndex("uniq_products_slug", bson.D{e("slug", 1)}),
			index("idx_products_dev_text_search", bson.D{e("title", "text"), e("description", "text"), e("brand_name", "text")}),
		},
	}
}

func categoriesCollectionDefinition() MongoCollectionDefinition {
	return MongoCollectionDefinition{
		Name: CollectionCategories,
		Validator: jsonSchema(
			bson.A{"_id", "name", "slug", "path", "depth", "sort_order", "is_active", "created_at", "updated_at"},
			bson.D{
				prop("_id", bson.D{e("bsonType", "string")}),
				prop("name", bson.D{e("bsonType", "string"), e("minLength", 1), e("maxLength", 120)}),
				prop("slug", bson.D{e("bsonType", "string")}),
				prop("parent_id", nullable("string")),
				prop("path", bson.D{e("bsonType", "array"), e("items", bson.D{e("bsonType", "string")})}),
				prop("depth", bson.D{e("bsonType", "int"), e("minimum", 0)}),
				prop("sort_order", bson.D{e("bsonType", "int")}),
				prop("is_active", bson.D{e("bsonType", "bool")}),
				prop("attribute_schema", bson.D{
					e("bsonType", "array"),
					e("items", bson.D{
						e("bsonType", "object"),
						e("required", bson.A{"key", "label", "type", "required", "filterable", "variant_axis"}),
						e("properties", bson.D{
							prop("key", bson.D{e("bsonType", "string")}),
							prop("label", bson.D{e("bsonType", "string")}),
							prop("type", bson.D{e("bsonType", "string")}),
							prop("required", bson.D{e("bsonType", "bool")}),
							prop("filterable", bson.D{e("bsonType", "bool")}),
							prop("variant_axis", bson.D{e("bsonType", "bool")}),
							prop("allowed_values", bson.D{e("bsonType", "array"), e("items", bson.D{e("bsonType", "string")})}),
							prop("unit", nullable("string")),
						}),
					}),
				}),
				prop("seo", nullableObject(bson.D{
					e("bsonType", "object"),
					e("properties", bson.D{
						prop("title", nullable("string")),
						prop("description", nullable("string")),
					}),
				})),
				prop("created_at", bson.D{e("bsonType", "date")}),
				prop("updated_at", bson.D{e("bsonType", "date")}),
			},
		),
		Indexes: []MongoIndexDefinition{
			index("idx_categories_parent_sort", bson.D{e("parent_id", 1), e("sort_order", 1)}),
			uniqueIndex("uniq_categories_slug", bson.D{e("slug", 1)}),
			index("idx_categories_path", bson.D{e("path", 1)}),
			index("idx_categories_active_sort", bson.D{e("is_active", 1), e("sort_order", 1)}),
			index("idx_categories_active_parent_sort_name", bson.D{e("is_active", 1), e("parent_id", 1), e("sort_order", 1), e("name", 1)}),
		},
	}
}

func brandsCollectionDefinition() MongoCollectionDefinition {
	return MongoCollectionDefinition{
		Name: CollectionBrands,
		Validator: jsonSchema(
			bson.A{"_id", "name", "slug", "status", "created_at", "updated_at"},
			bson.D{
				prop("_id", bson.D{e("bsonType", "string")}),
				prop("name", bson.D{e("bsonType", "string"), e("minLength", 1), e("maxLength", 120)}),
				prop("slug", bson.D{e("bsonType", "string")}),
				prop("description", nullable("string")),
				prop("logo_url", nullable("string")),
				prop("website_url", nullable("string")),
				prop("status", enum("active", "inactive", "archived")),
				prop("created_by", nullable("string")),
				prop("created_at", bson.D{e("bsonType", "date")}),
				prop("updated_at", bson.D{e("bsonType", "date")}),
			},
		),
		Indexes: []MongoIndexDefinition{
			uniqueIndex("uniq_brands_slug", bson.D{e("slug", 1)}),
			index("idx_brands_status_name", bson.D{e("status", 1), e("name", 1)}),
			index("idx_brands_name_text", bson.D{e("name", "text")}),
		},
	}
}

func inventoryReservationsCollectionDefinition() MongoCollectionDefinition {
	return MongoCollectionDefinition{
		Name: CollectionInventoryReservations,
		Validator: jsonSchema(
			bson.A{"_id", "order_id", "status", "items", "expires_at", "created_at", "updated_at"},
			bson.D{
				prop("_id", bson.D{e("bsonType", "string")}),
				prop("order_id", bson.D{e("bsonType", "string")}),
				prop("idempotency_key", nullable("string")),
				prop("status", enum("reserved", "committed", "released", "expired")),
				prop("items", bson.D{
					e("bsonType", "array"),
					e("minItems", 1),
					e("items", bson.D{
						e("bsonType", "object"),
						e("required", bson.A{"product_id", "variant_id", "sku", "seller_id", "quantity"}),
						e("properties", bson.D{
							prop("product_id", bson.D{e("bsonType", "string")}),
							prop("variant_id", bson.D{e("bsonType", "string")}),
							prop("sku", bson.D{e("bsonType", "string")}),
							prop("seller_id", bson.D{e("bsonType", "string")}),
							prop("quantity", bson.D{e("bsonType", bson.A{"int", "long"}), e("minimum", 1)}),
						}),
					}),
				}),
				prop("expires_at", bson.D{e("bsonType", "date")}),
				prop("created_at", bson.D{e("bsonType", "date")}),
				prop("updated_at", bson.D{e("bsonType", "date")}),
				prop("committed_at", bson.D{e("bsonType", bson.A{"date", "null"})}),
				prop("released_at", bson.D{e("bsonType", bson.A{"date", "null"})}),
				prop("expired_at", bson.D{e("bsonType", bson.A{"date", "null"})}),
				prop("reason", nullable("string")),
			},
		),
		Indexes: []MongoIndexDefinition{
			uniqueIndex("uq_inventory_reservations_order", bson.D{e("order_id", 1)}),
			sparseUniqueIndex("uq_inventory_reservations_idempotency", bson.D{e("idempotency_key", 1)}),
			index("idx_inventory_reservations_status_expires", bson.D{e("status", 1), e("expires_at", 1)}),
			ttlIndex("ttl_inventory_reservations_cleanup", bson.D{e("expires_at", 1)}, 86400),
		},
	}
}

func inventorySnapshotsCollectionDefinition() MongoCollectionDefinition {
	return MongoCollectionDefinition{
		Name: CollectionInventorySnapshots,
		Validator: jsonSchema(
			bson.A{
				"_id",
				"product_id",
				"variant_id",
				"sku",
				"seller_id",
				"snapshot_type",
				"stock_quantity",
				"reserved_quantity",
				"available_quantity",
				"created_at",
			},
			bson.D{
				prop("_id", bson.D{e("bsonType", "string")}),
				prop("product_id", bson.D{e("bsonType", "string")}),
				prop("variant_id", bson.D{e("bsonType", "string")}),
				prop("sku", bson.D{e("bsonType", "string")}),
				prop("seller_id", bson.D{e("bsonType", "string")}),
				prop("snapshot_type", enum("periodic", "manual_adjustment", "reservation", "release", "order_commit", "correction")),
				prop("stock_quantity", bson.D{e("bsonType", bson.A{"int", "long"}), e("minimum", 0)}),
				prop("reserved_quantity", bson.D{e("bsonType", bson.A{"int", "long"}), e("minimum", 0)}),
				prop("safety_stock", bson.D{e("bsonType", bson.A{"int", "long", "null"}), e("minimum", 0)}),
				prop("available_quantity", bson.D{e("bsonType", bson.A{"int", "long"})}),
				prop("reason", nullable("string")),
				prop("reference", nullableObject(bson.D{
					e("bsonType", "object"),
					e("properties", bson.D{
						prop("type", nullable("string")),
						prop("id", nullable("string")),
						prop("order_id", nullable("string")),
					}),
				})),
				prop("created_at", bson.D{e("bsonType", "date")}),
			},
		),
		Indexes: []MongoIndexDefinition{
			index("idx_inventory_product_variant_created", bson.D{e("product_id", 1), e("variant_id", 1), e("created_at", -1)}),
			index("idx_inventory_seller_created", bson.D{e("seller_id", 1), e("created_at", -1)}),
			index("idx_inventory_sku_created", bson.D{e("sku", 1), e("created_at", -1)}),
			index("idx_inventory_type_created", bson.D{e("snapshot_type", 1), e("created_at", -1)}),
		},
	}
}

func priceBooksCollectionDefinition() MongoCollectionDefinition {
	return MongoCollectionDefinition{
		Name: CollectionPriceBooks,
		Validator: jsonSchema(
			bson.A{"_id", "seller_id", "name", "currency", "status", "priority", "entries", "created_at", "updated_at"},
			bson.D{
				prop("_id", bson.D{e("bsonType", "string")}),
				prop("seller_id", bson.D{e("bsonType", "string")}),
				prop("name", bson.D{e("bsonType", "string"), e("minLength", 1), e("maxLength", 160)}),
				prop("currency", bson.D{e("bsonType", "string")}),
				prop("status", enum("draft", "active", "expired", "archived")),
				prop("priority", bson.D{e("bsonType", "int")}),
				prop("starts_at", bson.D{e("bsonType", bson.A{"date", "null"})}),
				prop("ends_at", bson.D{e("bsonType", bson.A{"date", "null"})}),
				prop("entries", bson.D{
					e("bsonType", "array"),
					e("items", bson.D{
						e("bsonType", "object"),
						e("required", bson.A{"entry_id", "product_id", "variant_id", "sku", "price"}),
						e("properties", bson.D{
							prop("entry_id", bson.D{e("bsonType", "string")}),
							prop("product_id", bson.D{e("bsonType", "string")}),
							prop("variant_id", bson.D{e("bsonType", "string")}),
							prop("sku", bson.D{e("bsonType", "string")}),
							prop("price", moneySchema()),
							prop("mrp", nullableObject(moneySchema())),
							prop("min_quantity", bson.D{e("bsonType", bson.A{"int", "long", "null"}), e("minimum", 1)}),
						}),
					}),
				}),
				prop("created_by", nullable("string")),
				prop("updated_by", nullable("string")),
				prop("created_at", bson.D{e("bsonType", "date")}),
				prop("updated_at", bson.D{e("bsonType", "date")}),
			},
		),
		Indexes: []MongoIndexDefinition{
			index("idx_price_books_seller_status_starts", bson.D{e("seller_id", 1), e("status", 1), e("starts_at", -1)}),
			index("idx_price_books_active_window_priority", bson.D{e("status", 1), e("starts_at", 1), e("ends_at", 1), e("priority", -1)}),
			index("idx_price_books_entry_product_variant_status", bson.D{e("entries.product_id", 1), e("entries.variant_id", 1), e("status", 1)}),
			index("idx_price_books_entry_sku_status", bson.D{e("entries.sku", 1), e("status", 1)}),
		},
	}
}

func productEventOutboxCollectionDefinition() MongoCollectionDefinition {
	return MongoCollectionDefinition{
		Name: CollectionProductEventOutbox,
		Validator: jsonSchema(
			bson.A{
				"_id",
				"topic",
				"event_type",
				"version",
				"source",
				"request_id",
				"trace_id",
				"payload",
				"status",
				"attempts",
				"next_attempt_at",
				"occurred_at",
			},
			bson.D{
				prop("_id", bson.D{e("bsonType", "string")}),
				prop("topic", bson.D{e("bsonType", "string")}),
				prop("event_type", enum(
					string(domain.ProductEventCreated),
					string(domain.ProductEventUpdated),
					string(domain.ProductEventPublished),
					string(domain.ProductEventUnpublished),
					string(domain.ProductEventInventoryChanged),
				)),
				prop("version", bson.D{e("bsonType", "int"), e("minimum", 1)}),
				prop("source", bson.D{e("bsonType", "string")}),
				prop("request_id", bson.D{e("bsonType", "string")}),
				prop("trace_id", bson.D{e("bsonType", "string")}),
				prop("payload", bson.D{e("bsonType", "object")}),
				prop("status", enum(
					string(domain.OutboxStatusPending),
					string(domain.OutboxStatusPublishing),
					string(domain.OutboxStatusPublished),
					string(domain.OutboxStatusDeadLettered),
				)),
				prop("attempts", bson.D{e("bsonType", "int"), e("minimum", 0)}),
				prop("next_attempt_at", bson.D{e("bsonType", "date")}),
				prop("occurred_at", bson.D{e("bsonType", "date")}),
				prop("published_at", bson.D{e("bsonType", bson.A{"date", "null"})}),
				prop("last_error", nullable("string")),
			},
		),
		Indexes: []MongoIndexDefinition{
			index("idx_product_event_outbox_status_next_occurred", bson.D{
				e("status", 1),
				e("next_attempt_at", 1),
				e("occurred_at", 1),
			}),
			index("idx_product_event_outbox_type_occurred", bson.D{e("event_type", 1), e("occurred_at", -1)}),
			ttlIndex("ttl_product_event_outbox_published", bson.D{e("published_at", 1)}, 2592000),
		},
	}
}

func jsonSchema(required bson.A, properties bson.D) bson.D {
	return bson.D{
		e("$jsonSchema", bson.D{
			e("bsonType", "object"),
			e("required", required),
			e("properties", properties),
		}),
	}
}

func moneySchema() bson.D {
	return bson.D{
		e("bsonType", "object"),
		e("required", bson.A{"amount", "currency"}),
		e("properties", bson.D{
			prop("amount", bson.D{e("bsonType", bson.A{"int", "long"}), e("minimum", 1)}),
			prop("currency", bson.D{e("bsonType", "string")}),
		}),
	}
}

func nullable(bsonType string) bson.D {
	return bson.D{e("bsonType", bson.A{bsonType, "null"})}
}

func nullableObject(schema bson.D) bson.D {
	nullableSchema := append(bson.D{}, schema...)
	nullableSchema[0] = e("bsonType", bson.A{"object", "null"})
	return nullableSchema
}

func enum(values ...string) bson.D {
	items := make(bson.A, 0, len(values))
	for _, value := range values {
		items = append(items, value)
	}
	return bson.D{e("enum", items)}
}

func index(name string, keys bson.D) MongoIndexDefinition {
	return MongoIndexDefinition{
		Name: name,
		Model: mongo.IndexModel{
			Keys:    keys,
			Options: options.Index().SetName(name),
		},
	}
}

func uniqueIndex(name string, keys bson.D) MongoIndexDefinition {
	definition := index(name, keys)
	definition.Model.Options = definition.Model.Options.SetUnique(true)
	return definition
}

func sparseUniqueIndex(name string, keys bson.D) MongoIndexDefinition {
	definition := uniqueIndex(name, keys)
	definition.Model.Options = definition.Model.Options.SetSparse(true)
	return definition
}

func ttlIndex(name string, keys bson.D, expireAfterSeconds int32) MongoIndexDefinition {
	definition := index(name, keys)
	definition.Model.Options = definition.Model.Options.SetExpireAfterSeconds(expireAfterSeconds)
	return definition
}

func prop(key string, value any) bson.E {
	return e(key, value)
}

func e(key string, value any) bson.E {
	return bson.E{Key: key, Value: value}
}
