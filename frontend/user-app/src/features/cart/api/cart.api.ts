import { apiDelete, apiGet, apiPatch, apiPost } from '../../../lib/http';
import type {
  AddCartItemRequest,
  Cart,
  CouponPreviewRequest,
  CouponPreviewResponse,
  UpdateCartItemRequest,
} from '../types';

export function getCart(signal?: AbortSignal) {
  return apiGet<Cart>('/api/v1/cart', { auth: true, signal });
}

export function addCartItem(body: AddCartItemRequest) {
  return apiPost<Cart, AddCartItemRequest>('/api/v1/cart/items', body, {
    auth: true,
  });
}

export function updateCartItem(itemId: string, body: UpdateCartItemRequest) {
  return apiPatch<Cart, UpdateCartItemRequest>(
    `/api/v1/cart/items/${encodeURIComponent(itemId)}`,
    body,
    { auth: true },
  );
}

export function removeCartItem(itemId: string) {
  return apiDelete<Cart>(`/api/v1/cart/items/${encodeURIComponent(itemId)}`, {
    auth: true,
  });
}

export function previewCoupon(body: CouponPreviewRequest) {
  return apiPost<CouponPreviewResponse, CouponPreviewRequest>(
    '/api/v1/cart/coupons/preview',
    body,
    { auth: true },
  );
}
