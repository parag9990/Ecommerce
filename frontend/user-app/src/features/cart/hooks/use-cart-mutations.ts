import { useMutation, useQueryClient } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import {
  addCartItem,
  previewCoupon,
  removeCartItem,
  updateCartItem,
} from '../api/cart.api';
import type {
  AddCartItemRequest,
  CouponPreviewRequest,
} from '../types';

export function useCartMutations() {
  const queryClient = useQueryClient();

  const addItem = useMutation({
    mutationFn: (body: AddCartItemRequest) => addCartItem(body),
    onSuccess: (cart) => {
      queryClient.setQueryData(queryKeys.cart.detail(), cart);
      void queryClient.invalidateQueries({ queryKey: queryKeys.wishlist.all });
    },
  });

  const updateItem = useMutation({
    mutationFn: (input: { itemId: string; quantity: number }) =>
      updateCartItem(input.itemId, { quantity: input.quantity }),
    onSuccess: (cart) => {
      queryClient.setQueryData(queryKeys.cart.detail(), cart);
    },
  });

  const removeItem = useMutation({
    mutationFn: (itemId: string) => removeCartItem(itemId),
    onSuccess: (cart) => {
      queryClient.setQueryData(queryKeys.cart.detail(), cart);
    },
  });

  const previewCouponCode = useMutation({
    mutationFn: (body: CouponPreviewRequest) => previewCoupon(body),
  });

  return {
    addItem,
    previewCouponCode,
    removeItem,
    updateItem,
  };
}
