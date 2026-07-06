import { apiDelete, apiGet, apiPost } from '../../../lib/http';
import { getProduct } from '../../product/api/product.api';
import type { Product } from '../../product/types';
import type {
  Wishlist,
  WishlistItem,
  WishlistItemAvailability,
  WishlistItemInput,
  WishlistMoveToCartResponse,
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

function readMoney(value: unknown): WishlistItem['price'] {
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

function readAvailability(value: unknown): WishlistItemAvailability | undefined {
  const availability = readString(value);

  switch (availability) {
    case 'deleted':
    case 'in_stock':
    case 'out_of_stock':
    case 'unknown':
      return availability;
    default:
      return undefined;
  }
}

function normalizeWishlistItem(value: unknown): WishlistItem | undefined {
  if (!isRecord(value)) {
    return undefined;
  }
  const productId = readString(value.product_id);

  if (!productId) {
    return undefined;
  }

  return {
    availability: readAvailability(value.availability),
    image_url: readString(value.image_url),
    price: readMoney(value.price) ?? readMoney(value.last_known_price),
    product_id: productId,
    title: readString(value.title),
    variant_id: readString(value.variant_id),
  };
}

function normalizeWishlist(value: unknown): Wishlist {
  const record = isRecord(value) ? value : {};
  const items = Array.isArray(record.items)
    ? record.items
        .map(normalizeWishlistItem)
        .filter((item): item is WishlistItem => Boolean(item))
    : [];

  return {
    items,
    user_id: readString(record.user_id),
    wishlist_id: readString(record.wishlist_id),
  };
}

function itemNeedsHydration(item: WishlistItem) {
  return !item.title || !item.image_url || !item.price || !item.variant_id;
}

function itemAvailabilityFromProduct(
  product: Product,
  variantId?: string,
): WishlistItemAvailability {
  if (product.status && product.status !== 'published') {
    return 'deleted';
  }

  const variant =
    product.variants?.find(
      (candidate) =>
        candidate.variant_id === variantId || candidate.sku === variantId,
    ) ?? product.variants?.[0];
  const stock = variant?.available_quantity ?? variant?.stock_quantity;

  if (stock === undefined) {
    return 'unknown';
  }

  return stock > 0 ? 'in_stock' : 'out_of_stock';
}

function hydrateWishlistItem(item: WishlistItem, product: Product): WishlistItem {
  const variant =
    product.variants?.find(
      (candidate) =>
        candidate.variant_id === item.variant_id || candidate.sku === item.variant_id,
    ) ?? product.variants?.[0];

  return {
    ...item,
    availability:
      item.availability === 'deleted'
        ? item.availability
        : itemAvailabilityFromProduct(product, item.variant_id),
    image_url: item.image_url ?? product.images?.[0],
    price: item.price ?? variant?.price,
    title: item.title ?? product.title,
    variant_id: item.variant_id ?? variant?.variant_id ?? variant?.sku,
  };
}

async function hydrateWishlist(
  wishlist: Wishlist,
  signal?: AbortSignal,
): Promise<Wishlist> {
  const itemsToHydrate = (wishlist.items ?? []).filter(itemNeedsHydration);

  if (itemsToHydrate.length === 0) {
    return wishlist;
  }

  const settledProducts = await Promise.allSettled(
    itemsToHydrate.map((item) =>
      getProduct(item.product_id, signal ? { signal } : {}),
    ),
  );

  if (signal?.aborted) {
    throw new DOMException('The operation was aborted.', 'AbortError');
  }

  const productsById = new Map<string, Product>();
  settledProducts.forEach((result, index) => {
    const item = itemsToHydrate[index];
    if (item && result.status === 'fulfilled') {
      productsById.set(item.product_id, result.value);
    }
  });

  if (productsById.size === 0) {
    return wishlist;
  }

  return {
    ...wishlist,
    items: (wishlist.items ?? []).map((item) => {
      const product = productsById.get(item.product_id);

      return product ? hydrateWishlistItem(item, product) : item;
    }),
  };
}

export function getWishlist(signal?: AbortSignal) {
  return apiGet<unknown>('/api/v1/wishlist', { auth: true, signal })
    .then(normalizeWishlist)
    .then((wishlist) => hydrateWishlist(wishlist, signal));
}

export function addWishlistItem(body: WishlistItemInput) {
  return apiPost<unknown, WishlistItemInput>('/api/v1/wishlist/items', body, {
    auth: true,
  })
    .then(normalizeWishlist)
    .then((wishlist) => hydrateWishlist(wishlist));
}

export function removeWishlistItem(productId: string) {
  return apiDelete<unknown>(
    `/api/v1/wishlist/items/${encodeURIComponent(productId)}`,
    { auth: true },
  )
    .then(normalizeWishlist)
    .then((wishlist) => hydrateWishlist(wishlist));
}

export function moveWishlistItemToCart(productId: string) {
  return apiPost<WishlistMoveToCartResponse, Record<string, never>>(
    `/api/v1/wishlist/items/${encodeURIComponent(productId)}/move-to-cart`,
    {},
    { auth: true },
  );
}
