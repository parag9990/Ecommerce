package usecase

import (
	"context"
	"errors"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

type CartMutationUsecase struct {
	addItem            *AddItemUsecase
	removeItem         *RemoveItemUsecase
	applyCouponPreview *ApplyCouponPreviewUsecase
	mergeGuestCart     *MergeGuestCartUsecase
}

func NewCartMutationUsecase(addItem *AddItemUsecase, removeItem *RemoveItemUsecase, applyCouponPreview *ApplyCouponPreviewUsecase, mergeGuestCart *MergeGuestCartUsecase) (*CartMutationUsecase, error) {
	if addItem == nil {
		return nil, errors.New("add item usecase is required")
	}
	if removeItem == nil {
		return nil, errors.New("remove item usecase is required")
	}
	if applyCouponPreview == nil {
		return nil, errors.New("apply coupon preview usecase is required")
	}
	if mergeGuestCart == nil {
		return nil, errors.New("merge guest cart usecase is required")
	}
	return &CartMutationUsecase{addItem: addItem, removeItem: removeItem, applyCouponPreview: applyCouponPreview, mergeGuestCart: mergeGuestCart}, nil
}

func (u *CartMutationUsecase) AddItem(ctx context.Context, cmd AddItemCommand) (*domain.Cart, error) {
	return u.addItem.AddItem(ctx, cmd)
}

func (u *CartMutationUsecase) RemoveItem(ctx context.Context, cmd RemoveItemCommand) (*domain.Cart, error) {
	return u.removeItem.RemoveItem(ctx, cmd)
}

func (u *CartMutationUsecase) ApplyCouponPreview(ctx context.Context, cmd ApplyCouponPreviewCommand) (*CouponPreviewResult, error) {
	return u.applyCouponPreview.ApplyCouponPreview(ctx, cmd)
}

func (u *CartMutationUsecase) MergeGuestCart(ctx context.Context, cmd MergeGuestCartCommand) (*domain.Cart, error) {
	return u.mergeGuestCart.MergeGuestCart(ctx, cmd)
}
