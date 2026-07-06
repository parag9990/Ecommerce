import { useMutation, useQueryClient } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import {
  addWishlistItem,
  moveWishlistItemToCart,
  removeWishlistItem,
} from '../api/wishlist.api';

export function useWishlistMutations() {
  const queryClient = useQueryClient();

  const addItem = useMutation({
    mutationFn: (input: { productId: string; variantId?: string | undefined }) =>
      addWishlistItem({
        product_id: input.productId,
        variant_id: input.variantId,
      }),
    onSuccess: (wishlist) => {
      queryClient.setQueryData(queryKeys.wishlist.detail(), wishlist);
    },
  });

  const removeItem = useMutation({
    mutationFn: (productId: string) => removeWishlistItem(productId),
    onSuccess: (wishlist) => {
      queryClient.setQueryData(queryKeys.wishlist.detail(), wishlist);
    },
  });

  const moveToCart = useMutation({
    mutationFn: (productId: string) => moveWishlistItemToCart(productId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.cart.all });
      void queryClient.invalidateQueries({ queryKey: queryKeys.wishlist.all });
    },
  });

  return {
    addItem,
    moveToCart,
    removeItem,
  };
}
