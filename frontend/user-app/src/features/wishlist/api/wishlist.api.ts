import { apiDelete, apiGet, apiPost } from '../../../lib/http';
import type {
  Wishlist,
  WishlistItemInput,
  WishlistMoveToCartResponse,
} from '../types';

export function getWishlist(signal?: AbortSignal) {
  return apiGet<Wishlist>('/api/v1/wishlist', { auth: true, signal });
}

export function addWishlistItem(body: WishlistItemInput) {
  return apiPost<Wishlist, WishlistItemInput>('/api/v1/wishlist/items', body, {
    auth: true,
  });
}

export function removeWishlistItem(productId: string) {
  return apiDelete<Wishlist>(
    `/api/v1/wishlist/items/${encodeURIComponent(productId)}`,
    { auth: true },
  );
}

export function moveWishlistItemToCart(productId: string) {
  return apiPost<WishlistMoveToCartResponse, Record<string, never>>(
    `/api/v1/wishlist/items/${encodeURIComponent(productId)}/move-to-cart`,
    {},
    { auth: true },
  );
}
