package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	productv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"product-service/internal/transport/dto"
	"product-service/internal/transport/inventory"
	"product-service/internal/transport/productread"
	"product-service/internal/transport/sellerproduct"
	"product-service/internal/usecase"
)

type Server struct {
	productv1.UnimplementedProductServiceServer

	read      *productread.Handler
	seller    *sellerproduct.Handler
	inventory *inventory.Handler
}

func NewServer(read *productread.Handler, seller *sellerproduct.Handler, inventoryHandler *inventory.Handler) (*Server, error) {
	if read == nil {
		return nil, fmt.Errorf("product read handler is required")
	}
	if seller == nil {
		return nil, fmt.Errorf("seller product handler is required")
	}
	if inventoryHandler == nil {
		return nil, fmt.Errorf("inventory handler is required")
	}
	return &Server{read: read, seller: seller, inventory: inventoryHandler}, nil
}

func (s *Server) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ProductListResponse, error) {
	actor := actorFromRequest(ctx, req.GetActor())
	var result *dto.ProductListResponseDTO
	var err error
	if actor.SellerID != "" {
		result, err = s.read.ListSellerProducts(ctx, dto.ListSellerProductsRequestDTO{
			Actor: actor, CategoryID: req.GetCategoryId(), Status: req.GetStatus(),
			Page: int(req.GetPage()), PageSize: int(req.GetPageSize()), Sort: req.GetSort(),
		})
	} else {
		result, err = s.read.ListProducts(ctx, dto.ListProductsRequestDTO{
			CategoryID: req.GetCategoryId(), SellerID: req.GetSellerId(), Status: req.GetStatus(),
			Page: int(req.GetPage()), PageSize: int(req.GetPageSize()), Sort: req.GetSort(),
		})
	}
	if err != nil {
		return nil, mapError(err)
	}
	return productList(result)
}

func (s *Server) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.Product, error) {
	actor := actorFromRequest(ctx, req.GetActor())
	var result *dto.ProductReadDTO
	var err error
	if actor.SellerID != "" {
		result, err = s.read.GetSellerProduct(ctx, dto.GetSellerProductRequestDTO{Actor: actor, ProductID: req.GetProductId()})
	} else {
		result, err = s.read.GetProduct(ctx, dto.GetProductRequestDTO{ProductID: req.GetProductId()})
	}
	if err != nil {
		return nil, mapError(err)
	}
	return productFromRead(*result)
}

func (s *Server) ListCategories(ctx context.Context, req *productv1.ListCategoriesRequest) (*productv1.CategoryListResponse, error) {
	result, err := s.read.ListCategories(ctx, dto.ListCategoriesRequestDTO{ParentID: req.GetParentId()})
	if err != nil {
		return nil, mapError(err)
	}
	categories := make([]*productv1.Category, 0, len(result.Categories))
	for _, category := range result.Categories {
		parentID := ""
		if category.ParentID != nil {
			parentID = *category.ParentID
		}
		categories = append(categories, &productv1.Category{
			CategoryId: category.ID, Name: category.Name, Slug: category.Slug, ParentId: parentID,
			Path: category.Path, Level: int32(category.Level), SortOrder: int32(category.SortOrder),
		})
	}
	return &productv1.CategoryListResponse{Categories: categories}, nil
}

func (s *Server) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.Product, error) {
	input, err := productInput(req.GetProduct())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result, err := s.seller.CreateProduct(ctx, dto.CreateSellerProductRequestDTO{Actor: actorFromRequest(ctx, req.GetActor()), Product: input})
	if err != nil {
		return nil, mapError(err)
	}
	return productFromDTO(*result)
}

func (s *Server) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.Product, error) {
	input, err := productInput(req.GetProduct())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result, err := s.seller.UpdateProduct(ctx, dto.UpdateSellerProductRequestDTO{
		Actor: actorFromRequest(ctx, req.GetActor()), ProductID: req.GetProductId(), Product: input,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return productFromDTO(*result)
}

func (s *Server) PublishProduct(ctx context.Context, req *productv1.ProductLifecycleRequest) (*productv1.Product, error) {
	result, err := s.seller.PublishProduct(ctx, dto.ProductLifecycleRequestDTO{Actor: actorFromRequest(ctx, req.GetActor()), ProductID: req.GetProductId()})
	if err != nil {
		return nil, mapError(err)
	}
	return productFromDTO(*result)
}

func (s *Server) UnpublishProduct(ctx context.Context, req *productv1.ProductLifecycleRequest) (*productv1.Product, error) {
	result, err := s.seller.UnpublishProduct(ctx, dto.ProductLifecycleRequestDTO{Actor: actorFromRequest(ctx, req.GetActor()), ProductID: req.GetProductId()})
	if err != nil {
		return nil, mapError(err)
	}
	return productFromDTO(*result)
}

func (s *Server) BatchGetProducts(ctx context.Context, req *productv1.BatchGetProductsRequest) (*productv1.ProductListResponse, error) {
	result, err := s.read.BatchGetProducts(ctx, dto.BatchGetProductsRequestDTO{ProductIDs: req.GetProductIds()})
	if err != nil {
		return nil, mapError(err)
	}
	return productList(result)
}

func (s *Server) ReserveInventory(ctx context.Context, req *productv1.InventoryReservationRequest) (*productv1.InventoryReservationResponse, error) {
	items := make([]dto.InventoryReservationItemDTO, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		items = append(items, dto.InventoryReservationItemDTO{ProductID: item.GetProductId(), VariantID: item.GetVariantId(), Quantity: item.GetQuantity()})
	}
	result, err := s.inventory.ReserveInventory(ctx, dto.InventoryReservationRequestDTO{
		OrderID: req.GetOrderId(), Items: items, TTLSeconds: int(req.GetTtlSeconds()), IdempotencyKey: req.GetIdempotencyKey(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &productv1.InventoryReservationResponse{ReservationId: result.ReservationID, ExpiresAt: timestamppb.New(result.ExpiresAt)}, nil
}

func (s *Server) ReleaseInventory(ctx context.Context, req *productv1.InventoryReservationActionRequest) (*productv1.SuccessResponse, error) {
	_, err := s.inventory.ReleaseInventory(ctx, dto.InventoryReservationActionRequestDTO{ReservationID: req.GetReservationId(), Reason: req.GetReason()})
	if err != nil {
		return nil, mapError(err)
	}
	return &productv1.SuccessResponse{Success: true}, nil
}

func (s *Server) CommitInventory(ctx context.Context, req *productv1.InventoryReservationActionRequest) (*productv1.SuccessResponse, error) {
	_, err := s.inventory.CommitInventory(ctx, dto.InventoryReservationActionRequestDTO{ReservationID: req.GetReservationId(), Reason: req.GetReason()})
	if err != nil {
		return nil, mapError(err)
	}
	return &productv1.SuccessResponse{Success: true}, nil
}

func actorFromRequest(ctx context.Context, actor *productv1.ActorContext) dto.ActorContextDTO {
	result := dto.ActorContextDTO{}
	if actor != nil {
		result = dto.ActorContextDTO{UserID: actor.GetUserId(), SellerID: actor.GetSellerId(), Roles: actor.GetRoles(), Permissions: actor.GetPermissions()}
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if strings.TrimSpace(result.UserID) == "" {
			result.UserID = firstMetadata(md, "x-user-id")
		}
		if strings.TrimSpace(result.SellerID) == "" {
			result.SellerID = firstMetadata(md, "x-seller-id")
		}
		if len(result.Roles) == 0 {
			result.Roles = splitCSV(firstMetadata(md, "x-roles"))
		}
		if len(result.Permissions) == 0 {
			result.Permissions = splitCSV(firstMetadata(md, "x-permissions"))
		}
	}
	return result
}

func firstMetadata(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func splitCSV(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}

func productInput(input *productv1.ProductInput) (dto.ProductInputDTO, error) {
	if input == nil {
		return dto.ProductInputDTO{}, fmt.Errorf("product is required")
	}
	images := make([]dto.ProductImageDTO, 0, len(input.GetImages()))
	for _, image := range input.GetImages() {
		images = append(images, dto.ProductImageDTO{
			ID: image.GetImageId(), URL: image.GetUrl(), AltText: image.GetAltText(), Position: int(image.GetPosition()),
			IsPrimary: image.GetIsPrimary(), VariantIDs: image.GetVariantIds(), Width: int(image.GetWidth()), Height: int(image.GetHeight()), Status: image.GetStatus(),
		})
	}
	variants := make([]dto.ProductVariantInputDTO, 0, len(input.GetVariants()))
	for _, variant := range input.GetVariants() {
		if variant.GetPrice() == nil {
			return dto.ProductInputDTO{}, fmt.Errorf("variant price is required")
		}
		var mrp *dto.MoneyDTO
		if variant.GetMrp() != nil {
			mrp = &dto.MoneyDTO{Amount: variant.GetMrp().GetAmount(), Currency: variant.GetMrp().GetCurrency()}
		}
		variants = append(variants, dto.ProductVariantInputDTO{
			SKU: variant.GetSku(), Attributes: structMap(variant.GetAttributes()),
			Price: dto.MoneyDTO{Amount: variant.GetPrice().GetAmount(), Currency: variant.GetPrice().GetCurrency()},
			MRP:   mrp, StockQuantity: variant.GetStockQuantity(),
		})
	}
	return dto.ProductInputDTO{
		Title: input.GetTitle(), Description: input.GetDescription(), Brand: input.GetBrand(), CategoryID: input.GetCategoryId(),
		Attributes: structMap(input.GetAttributes()), ImageDetails: images, Variants: variants,
	}, nil
}

func productList(input *dto.ProductListResponseDTO) (*productv1.ProductListResponse, error) {
	products := make([]*productv1.Product, 0, len(input.Products))
	for _, item := range input.Products {
		product, err := productFromRead(item)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return &productv1.ProductListResponse{Products: products, Total: input.Total}, nil
}

func productFromDTO(input dto.ProductDTO) (*productv1.Product, error) {
	read := dto.ProductReadDTO{
		ID: input.ID, SellerID: input.SellerID, Title: input.Title, Slug: input.Slug, Description: input.Description,
		Brand: input.Brand, CategoryID: input.CategoryID, CategoryPath: input.CategoryPath, Status: input.Status,
		Attributes: input.Attributes, RatingSummary: input.RatingSummary, UpdatedAt: input.UpdatedAt, PublishedAt: input.PublishedAt,
	}
	read.Images = make([]dto.ProductImageReadDTO, 0, len(input.Images))
	for _, image := range input.Images {
		read.Images = append(read.Images, dto.ProductImageReadDTO(image))
	}
	read.Variants = make([]dto.ProductVariantReadDTO, 0, len(input.Variants))
	for _, variant := range input.Variants {
		read.Variants = append(read.Variants, dto.ProductVariantReadDTO{
			ID: variant.ID, SKU: variant.SKU, Title: variant.Title, Attributes: variant.Attributes, Price: variant.Price,
			MRP: variant.MRP, StockQuantity: variant.StockQuantity,
			AvailableQuantity: variant.StockQuantity - variant.ReservedQuantity - variant.SafetyStock, Status: variant.Status,
		})
	}
	product, err := productFromRead(read)
	if err != nil {
		return nil, err
	}
	product.CreatedAt = timestamppb.New(input.CreatedAt)
	return product, nil
}

func productFromRead(input dto.ProductReadDTO) (*productv1.Product, error) {
	attributes, err := structpb.NewStruct(input.Attributes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "encode product attributes: %v", err)
	}
	images := make([]*productv1.ProductImage, 0, len(input.Images))
	for _, image := range input.Images {
		images = append(images, &productv1.ProductImage{
			ImageId: image.ID, Url: image.URL, AltText: image.AltText, Position: int32(image.Position), IsPrimary: image.IsPrimary,
			VariantIds: image.VariantIDs, Width: int32(image.Width), Height: int32(image.Height), Status: image.Status,
		})
	}
	variants := make([]*productv1.ProductVariant, 0, len(input.Variants))
	for _, variant := range input.Variants {
		variantAttributes, err := structpb.NewStruct(variant.Attributes)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "encode variant attributes: %v", err)
		}
		var mrp *productv1.Money
		if variant.MRP != nil {
			mrp = &productv1.Money{Amount: variant.MRP.Amount, Currency: variant.MRP.Currency}
		}
		variants = append(variants, &productv1.ProductVariant{
			VariantId: variant.ID, Sku: variant.SKU, Title: variant.Title, Attributes: variantAttributes,
			Price: &productv1.Money{Amount: variant.Price.Amount, Currency: variant.Price.Currency}, Mrp: mrp,
			StockQuantity: variant.StockQuantity, AvailableQuantity: variant.AvailableQuantity, Status: variant.Status,
		})
	}
	result := &productv1.Product{
		ProductId: input.ID, SellerId: input.SellerID, Title: input.Title, Slug: input.Slug, Description: input.Description,
		Brand: input.Brand, CategoryId: input.CategoryID, CategoryPath: input.CategoryPath, Status: input.Status,
		Attributes: attributes, Images: images, Variants: variants,
		RatingSummary: &productv1.RatingSummary{Average: input.RatingSummary.Average, Count: input.RatingSummary.Count},
		UpdatedAt:     timestamppb.New(input.UpdatedAt),
	}
	if input.PublishedAt != nil {
		result.PublishedAt = timestamppb.New(*input.PublishedAt)
	}
	return result, nil
}

func structMap(value *structpb.Struct) map[string]any {
	if value == nil {
		return nil
	}
	return value.AsMap()
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, "request canceled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	}
	var serviceErr *usecase.ServiceError
	if !errors.As(err, &serviceErr) {
		return status.Error(codes.Internal, "internal product service error")
	}
	code := codes.Internal
	switch serviceErr.Kind {
	case usecase.ErrorKindInvalidArgument:
		code = codes.InvalidArgument
	case usecase.ErrorKindUnauthenticated:
		code = codes.Unauthenticated
	case usecase.ErrorKindPermissionDenied:
		code = codes.PermissionDenied
	case usecase.ErrorKindNotFound:
		code = codes.NotFound
	case usecase.ErrorKindAlreadyExists:
		code = codes.AlreadyExists
	case usecase.ErrorKindConflict:
		code = codes.Aborted
	case usecase.ErrorKindFailedPrecondition:
		code = codes.FailedPrecondition
	case usecase.ErrorKindUnavailable:
		code = codes.Unavailable
	}
	return status.Error(code, serviceErr.Message)
}
