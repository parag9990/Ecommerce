package repository

import (
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func cartValidator() bson.D {
	moneySchema := bson.D{
		{Key: "bsonType", Value: "object"},
		{Key: "required", Value: bson.A{"amount", "currency"}},
		{Key: "additionalProperties", Value: false},
		{Key: "properties", Value: bson.D{
			{Key: "amount", Value: bson.D{
				{Key: "bsonType", Value: bson.A{"int", "long"}},
				{Key: "minimum", Value: int64(0)},
			}},
			{Key: "currency", Value: bson.D{
				{Key: "bsonType", Value: "string"},
				{Key: "pattern", Value: "^[A-Z]{3}$"},
			}},
		}},
	}

	itemSchema := bson.D{
		{Key: "bsonType", Value: "object"},
		{Key: "required", Value: bson.A{
			"item_id",
			"product_id",
			"variant_id",
			"seller_id",
			"title_snapshot",
			"unit_price",
			"quantity",
			"line_subtotal",
			"price_snapshot_at",
			"added_at",
			"updated_at",
		}},
		{Key: "properties", Value: bson.D{
			{Key: "item_id", Value: bson.D{{Key: "bsonType", Value: "string"}, {Key: "minLength", Value: 1}}},
			{Key: "product_id", Value: bson.D{{Key: "bsonType", Value: "string"}, {Key: "minLength", Value: 1}}},
			{Key: "variant_id", Value: bson.D{{Key: "bsonType", Value: "string"}, {Key: "minLength", Value: 1}}},
			{Key: "seller_id", Value: bson.D{{Key: "bsonType", Value: "string"}, {Key: "minLength", Value: 1}}},
			{Key: "sku_snapshot", Value: bson.D{{Key: "bsonType", Value: bson.A{"string", "null"}}}},
			{Key: "title_snapshot", Value: bson.D{{Key: "bsonType", Value: "string"}, {Key: "minLength", Value: 1}}},
			{Key: "image_url_snapshot", Value: bson.D{{Key: "bsonType", Value: bson.A{"string", "null"}}}},
			{Key: "variant_snapshot", Value: bson.D{{Key: "bsonType", Value: "object"}}},
			{Key: "unit_price", Value: moneySchema},
			{Key: "quantity", Value: bson.D{
				{Key: "bsonType", Value: bson.A{"int", "long"}},
				{Key: "minimum", Value: domain.MinItemQuantity},
				{Key: "maximum", Value: domain.MaxItemQuantity},
			}},
			{Key: "line_subtotal", Value: moneySchema},
			{Key: "price_snapshot_at", Value: bson.D{{Key: "bsonType", Value: "date"}}},
			{Key: "added_at", Value: bson.D{{Key: "bsonType", Value: "date"}}},
			{Key: "updated_at", Value: bson.D{{Key: "bsonType", Value: "date"}}},
		}},
	}

	totalsSchema := bson.D{
		{Key: "bsonType", Value: "object"},
		{Key: "required", Value: bson.A{"subtotal", "discount", "total", "currency", "item_count", "unique_item_count"}},
		{Key: "additionalProperties", Value: false},
		{Key: "properties", Value: bson.D{
			{Key: "subtotal", Value: moneySchema},
			{Key: "discount", Value: moneySchema},
			{Key: "total", Value: moneySchema},
			{Key: "currency", Value: bson.D{{Key: "bsonType", Value: "string"}, {Key: "pattern", Value: "^[A-Z]{3}$"}}},
			{Key: "item_count", Value: bson.D{{Key: "bsonType", Value: bson.A{"int", "long"}}, {Key: "minimum", Value: 0}}},
			{Key: "unique_item_count", Value: bson.D{{Key: "bsonType", Value: bson.A{"int", "long"}}, {Key: "minimum", Value: 0}}},
		}},
	}

	couponPreviewSchema := bson.D{
		{Key: "bsonType", Value: bson.A{"object", "null"}},
		{Key: "properties", Value: bson.D{
			{Key: "coupon_id", Value: bson.D{{Key: "bsonType", Value: "string"}}},
			{Key: "code", Value: bson.D{{Key: "bsonType", Value: "string"}}},
			{Key: "valid", Value: bson.D{{Key: "bsonType", Value: "bool"}}},
			{Key: "discount", Value: moneySchema},
			{Key: "reason", Value: bson.D{{Key: "bsonType", Value: "string"}}},
		}},
	}

	return bson.D{
		{Key: "$jsonSchema", Value: bson.D{
			{Key: "bsonType", Value: "object"},
			{Key: "required", Value: bson.A{
				"_id",
				"status",
				"items",
				"totals",
				"version",
				"created_at",
				"updated_at",
				"expires_at",
			}},
			{Key: "properties", Value: bson.D{
				{Key: "_id", Value: bson.D{{Key: "bsonType", Value: "string"}, {Key: "minLength", Value: 1}}},
				{Key: "user_id", Value: bson.D{{Key: "bsonType", Value: bson.A{"string", "null"}}}},
				{Key: "guest_session_id", Value: bson.D{{Key: "bsonType", Value: bson.A{"string", "null"}}}},
				{Key: "status", Value: bson.D{{Key: "enum", Value: bson.A{"active", "merged", "checked_out", "expired", "abandoned"}}}},
				{Key: "items", Value: bson.D{
					{Key: "bsonType", Value: "array"},
					{Key: "maxItems", Value: domain.MaxUniqueItemsPerCart},
					{Key: "items", Value: itemSchema},
				}},
				{Key: "coupon_code", Value: bson.D{{Key: "bsonType", Value: bson.A{"string", "null"}}}},
				{Key: "coupon_preview", Value: couponPreviewSchema},
				{Key: "totals", Value: totalsSchema},
				{Key: "version", Value: bson.D{{Key: "bsonType", Value: bson.A{"int", "long"}}, {Key: "minimum", Value: 1}}},
				{Key: "created_at", Value: bson.D{{Key: "bsonType", Value: "date"}}},
				{Key: "updated_at", Value: bson.D{{Key: "bsonType", Value: "date"}}},
				{Key: "expires_at", Value: bson.D{{Key: "bsonType", Value: "date"}}},
				{Key: "merged_into_cart_id", Value: bson.D{{Key: "bsonType", Value: bson.A{"string", "null"}}}},
				{Key: "checked_out_order_id", Value: bson.D{{Key: "bsonType", Value: bson.A{"string", "null"}}}},
			}},
		}},
	}
}

func createCollectionOptions() *options.CreateCollectionOptions {
	return options.CreateCollection().
		SetValidator(cartValidator()).
		SetValidationAction("error").
		SetValidationLevel("strict")
}

func cartIndexModels() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_carts_user_status"),
		},
		{
			Keys:    bson.D{{Key: "guest_session_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_carts_guest_status"),
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetName("idx_carts_expires_at_ttl").SetExpireAfterSeconds(0),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().
				SetName("uniq_active_cart_per_user").
				SetUnique(true).
				SetPartialFilterExpression(bson.D{
					{Key: "status", Value: string(domain.CartStatusActive)},
					{Key: "user_id", Value: bson.D{{Key: "$type", Value: "string"}}},
				}),
		},
		{
			Keys: bson.D{{Key: "guest_session_id", Value: 1}},
			Options: options.Index().
				SetName("uniq_active_cart_per_guest_session").
				SetUnique(true).
				SetPartialFilterExpression(bson.D{
					{Key: "status", Value: string(domain.CartStatusActive)},
					{Key: "guest_session_id", Value: bson.D{{Key: "$type", Value: "string"}}},
				}),
		},
		{
			Keys:    bson.D{{Key: "items.product_id", Value: 1}, {Key: "items.variant_id", Value: 1}},
			Options: options.Index().SetName("idx_carts_item_product_variant"),
		},
		{
			Keys:    bson.D{{Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("idx_carts_updated_at"),
		},
	}
}
