package grpc

import (
	"context"
	"testing"

	productv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/product/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestActorFromRequestUsesTrustedMetadataFallback(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-user-id", "user_1",
		"x-seller-id", "seller_1",
		"x-roles", "seller,seller_manager",
		"x-permissions", "product:write:own_seller",
	))
	actor := actorFromRequest(ctx, nil)
	if actor.UserID != "user_1" || actor.SellerID != "seller_1" {
		t.Fatalf("unexpected actor identity: %+v", actor)
	}
	if len(actor.Roles) != 2 || len(actor.Permissions) != 1 {
		t.Fatalf("unexpected actor grants: %+v", actor)
	}
}

func TestProductInputMapsStructuredMediaAndAttributes(t *testing.T) {
	attributes, err := structpb.NewStruct(map[string]any{"color": "black"})
	if err != nil {
		t.Fatal(err)
	}
	input, err := productInput(&productv1.ProductInput{
		Title: "Shoe", CategoryId: "cat_1", Attributes: attributes,
		Images: []*productv1.ProductImage{{
			ImageId: "img_1", Url: "https://cdn.example.com/shoe.jpg", AltText: "Black shoe",
			Position: 1, IsPrimary: true, Status: "active",
		}},
		Variants: []*productv1.ProductVariantInput{{
			Sku: "SHOE-BLK", Price: &productv1.Money{Amount: 10000, Currency: "INR"}, StockQuantity: 5,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if input.ImageDetails[0].AltText != "Black shoe" || !input.ImageDetails[0].IsPrimary {
		t.Fatalf("media metadata was not preserved: %+v", input.ImageDetails[0])
	}
	if input.Attributes["color"] != "black" || input.Variants[0].Price.Amount != 10000 {
		t.Fatalf("product input was not preserved: %+v", input)
	}
}
