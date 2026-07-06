import { apiDelete, apiGet, apiPatch, apiPost } from '../../../lib/http';
import type {
  AddCartItemRequest,
  Cart,
  CartItem,
  CouponPreviewRequest,
  CouponPreviewResponse,
  Money,
  UpdateCartItemRequest,
} from '../types';

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function readString(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim().length > 0
    ? value.trim()
    : undefined;
}

function readNumber(value: unknown): number | undefined {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === 'string' && value.trim().length > 0) {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : undefined;
  }
  return undefined;
}

function readMoney(value: unknown): Money | undefined {
  if (!isRecord(value)) {
    return undefined;
  }
  const amount = readNumber(value.amount);
  if (amount === undefined) {
    return undefined;
  }
  return {
    amount,
    currency: readString(value.currency),
  };
}

function variantLabel(value: unknown): string | undefined {
  if (!isRecord(value)) {
    return undefined;
  }
  const parts = Object.entries(value)
    .map(([key, entry]) => {
      if (!['boolean', 'number', 'string'].includes(typeof entry)) {
        return undefined;
      }
      return `${key}: ${String(entry)}`;
    })
    .filter((entry): entry is string => Boolean(entry));

  return parts.length > 0 ? parts.join(' / ') : undefined;
}

function normalizeCartItem(value: unknown): CartItem | undefined {
  if (!isRecord(value)) {
    return undefined;
  }
  const itemId = readString(value.item_id);
  const productId = readString(value.product_id);
  const variantId = readString(value.variant_id);
  const quantity = readNumber(value.quantity);

  if (!itemId || !productId || !variantId || quantity === undefined) {
    return undefined;
  }

  return {
    image_url: readString(value.image_url) ?? readString(value.image_url_snapshot),
    item_id: itemId,
    line_total: readMoney(value.line_total) ?? readMoney(value.line_subtotal),
    product_id: productId,
    quantity,
    stock_status: readString(value.stock_status) as CartItem['stock_status'],
    title: readString(value.title) ?? readString(value.title_snapshot),
    unit_price: readMoney(value.unit_price),
    variant_id: variantId,
    variant_label:
      readString(value.variant_label) ??
      readString(value.sku_snapshot) ??
      variantLabel(value.variant_snapshot),
  };
}

function normalizeCart(value: unknown): Cart {
  const record = isRecord(value) ? value : {};
  const items = Array.isArray(record.items)
    ? record.items
        .map(normalizeCartItem)
        .filter((item): item is CartItem => Boolean(item))
    : [];

  return {
    cart_id: readString(record.cart_id),
    discount: readMoney(record.discount),
    items,
    subtotal: readMoney(record.subtotal),
    total: readMoney(record.total),
    user_id: readString(record.user_id),
  };
}

function normalizeCouponPreview(value: unknown): CouponPreviewResponse {
  const record = isRecord(value) ? value : {};

  return {
    coupon_id: readString(record.coupon_id),
    discount: readMoney(record.discount),
    reason: readString(record.reason),
    valid: typeof record.valid === 'boolean' ? record.valid : undefined,
  };
}

export function getCart(signal?: AbortSignal) {
  return apiGet<unknown>('/api/v1/cart', { auth: true, signal }).then(
    normalizeCart,
  );
}

export function addCartItem(body: AddCartItemRequest) {
  return apiPost<unknown, AddCartItemRequest>('/api/v1/cart/items', body, {
    auth: true,
  }).then(normalizeCart);
}

export function updateCartItem(itemId: string, body: UpdateCartItemRequest) {
  return apiPatch<unknown, UpdateCartItemRequest>(
    `/api/v1/cart/items/${encodeURIComponent(itemId)}`,
    body,
    { auth: true },
  ).then(normalizeCart);
}

export function removeCartItem(itemId: string) {
  return apiDelete<unknown>(`/api/v1/cart/items/${encodeURIComponent(itemId)}`, {
    auth: true,
  }).then(normalizeCart);
}

export function previewCoupon(body: CouponPreviewRequest) {
  return apiPost<unknown, CouponPreviewRequest>(
    '/api/v1/cart/coupons/preview',
    body,
    { auth: true },
  ).then(normalizeCouponPreview);
}
