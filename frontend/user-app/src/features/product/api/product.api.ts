import { ApiError, apiGet } from '../../../lib/http';
import type {
  AutocompleteResponse,
  Category,
  CategoryListResponse,
  Money,
  Product,
  ProductFilters,
  ProductVariant,
  ProductListResponse,
  SearchResponse,
} from '../types';

type RequestOptions = {
  signal?: AbortSignal;
};

function appendIfPresent(
  params: URLSearchParams,
  key: string,
  value?: boolean | number | string,
) {
  if (value === undefined || value === null || value === '') {
    return;
  }

  params.set(key, String(value));
}

function appendSearchFilters(params: URLSearchParams, filters: ProductFilters) {
  appendIfPresent(params, 'category_id', filters.categoryId);
  appendIfPresent(params, 'brand', filters.brand);
  appendIfPresent(params, 'seller_id', filters.sellerId);
  appendIfPresent(params, 'min_price', filters.minPrice);
  appendIfPresent(params, 'max_price', filters.maxPrice);
  appendIfPresent(params, 'min_rating', filters.minRating);
  appendIfPresent(params, 'in_stock', filters.inStock ? true : undefined);
}

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

function readRecord(value: unknown): Record<string, unknown> | undefined {
  return isRecord(value) ? value : undefined;
}

function isDefined<TValue>(value: TValue | undefined): value is TValue {
  return value !== undefined;
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

function normalizeImageUrls(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return [];
  }

  const images = value
    .map((image, index) => {
      if (typeof image === 'string') {
        return {
          isPrimary: index === 0,
          position: index,
          status: 'active',
          url: image.trim(),
        };
      }
      if (!isRecord(image)) {
        return undefined;
      }
      const url = readString(image.url);
      if (!url) {
        return undefined;
      }
      return {
        isPrimary: image.is_primary === true,
        position: readNumber(image.position) ?? index,
        status: readString(image.status) ?? 'active',
        url,
      };
    })
    .filter((image): image is {
      isPrimary: boolean;
      position: number;
      status: string;
      url: string;
    } => Boolean(image?.url))
    .filter((image) => image.status === 'active' || image.status === '');

  return images
    .sort((left, right) => {
      if (left.isPrimary !== right.isPrimary) {
        return left.isPrimary ? -1 : 1;
      }
      return left.position - right.position;
    })
    .map((image) => image.url);
}

function normalizeVariant(value: unknown): ProductVariant | undefined {
  if (!isRecord(value)) {
    return undefined;
  }
  const variantId = readString(value.variant_id) ?? readString(value.id);
  const sku = readString(value.sku) ?? variantId;

  if (!sku) {
    return undefined;
  }

  const variant: ProductVariant = {
    sku,
  };
  if (variantId) {
    variant.variant_id = variantId;
  }
  if (isRecord(value.attributes)) {
    variant.attributes = value.attributes;
  }
  const price = readMoney(value.price) ?? readMoney(value.unit_price);
  if (price) {
    variant.price = price;
  }
  const stockQuantity = readNumber(value.stock_quantity);
  if (stockQuantity !== undefined) {
    variant.stock_quantity = stockQuantity;
  }
  const availableQuantity = readNumber(value.available_quantity);
  if (availableQuantity !== undefined) {
    variant.available_quantity = availableQuantity;
  }
  return variant;
}

function normalizeProduct(value: unknown): Product | undefined {
  if (!isRecord(value)) {
    return undefined;
  }

  const productId = readString(value.product_id) ?? readString(value.id);
  const title = readString(value.title) ?? readString(value.name);

  if (!productId || !title) {
    return undefined;
  }

  const categoryPath = Array.isArray(value.category_path)
    ? value.category_path.map(readString).filter(isDefined)
    : undefined;
  const variants = Array.isArray(value.variants)
    ? value.variants.map(normalizeVariant).filter(isDefined)
    : [];
  const ratingSummary = readRecord(value.rating_summary);
  const product: Product = {
    product_id: productId,
    title,
  };

  const brand = readString(value.brand);
  if (brand) product.brand = brand;
  const categoryId = readString(value.category_id);
  if (categoryId) product.category_id = categoryId;
  if (categoryPath?.length) product.category_ids = categoryPath;
  const description = readString(value.description);
  if (description) product.description = description;
  const images = normalizeImageUrls(value.images);
  if (images.length > 0) product.images = images;
  const rating = readNumber(value.rating) ?? readNumber(ratingSummary?.average);
  if (rating !== undefined) product.rating = rating;
  const sellerId = readString(value.seller_id);
  if (sellerId) product.seller_id = sellerId;
  const status = readString(value.status);
  if (status) product.status = status;
  product.variants = variants;
  return product;
}

function normalizeProductArray(value: unknown): Product[] {
  return Array.isArray(value)
    ? value
        .map(normalizeProduct)
        .filter((product): product is Product => Boolean(product))
    : [];
}

function normalizeProductListResponse(value: unknown): ProductListResponse {
  const record = readRecord(value);
  const products = normalizeProductArray(record?.products);
  return {
    products,
    total: readNumber(record?.total) ?? products.length,
  };
}

function normalizeSearchResponse(value: unknown): SearchResponse {
  const record = readRecord(value);
  const products = normalizeProductArray(record?.products);
  return {
    facets: isRecord(record?.facets) ? record.facets : {},
    products,
    total: readNumber(record?.total) ?? products.length,
  };
}

function normalizeCategoryListResponse(value: unknown): CategoryListResponse {
  const record = readRecord(value);
  const categories: Category[] = Array.isArray(record?.categories)
    ? record.categories
        .map((category) => {
          if (!isRecord(category)) {
            return undefined;
          }
          const categoryId = readString(category.category_id) ?? readString(category.id);
          const name = readString(category.name);
          if (!categoryId || !name) {
            return undefined;
          }
          const normalized: Category = {
            category_id: categoryId,
            name,
          };
          const parentId = readString(category.parent_id);
          if (parentId) {
            normalized.parent_id = parentId;
          }
          return normalized;
        })
        .filter((category): category is Category => Boolean(category))
    : [];

  return { categories };
}

function normalizeAutocompleteResponse(value: unknown): AutocompleteResponse {
  const record = readRecord(value);
  return {
    suggestions: Array.isArray(record?.suggestions)
      ? record.suggestions.map(readString).filter(isDefined)
      : [],
  };
}

export function listProducts(
  input: {
    categoryId?: string | undefined;
    page: number;
    pageSize: number;
    sort?: ProductFilters['sort'] | undefined;
  },
  options: RequestOptions = {},
) {
  const params = new URLSearchParams();

  appendIfPresent(params, 'page', input.page);
  appendIfPresent(params, 'page_size', input.pageSize);
  appendIfPresent(params, 'category_id', input.categoryId);
  appendIfPresent(params, 'sort', input.sort);
  appendIfPresent(params, 'status', 'published');

  return apiGet<unknown>(`/api/v1/products?${params.toString()}`, {
    signal: options.signal,
  }).then(normalizeProductListResponse);
}

export function getProduct(productId: string, options: RequestOptions = {}) {
  return apiGet<unknown>(`/api/v1/products/${encodeURIComponent(productId)}`, {
    signal: options.signal,
  }).then((product) => {
    const normalized = normalizeProduct(product);
    if (!normalized) {
      throw new ApiError(
        'The server returned an invalid product response.',
        'INVALID_PRODUCT_RESPONSE',
        200,
      );
    }
    return normalized;
  });
}

export function listCategories(options: RequestOptions = {}) {
  return apiGet<unknown>('/api/v1/categories', {
    signal: options.signal,
  }).then(normalizeCategoryListResponse);
}

export function searchProducts(
  filters: ProductFilters,
  options: RequestOptions = {},
) {
  const params = new URLSearchParams();

  appendIfPresent(params, 'q', filters.q);
  appendIfPresent(params, 'sort', filters.sort);
  appendIfPresent(params, 'page', filters.page);
  appendIfPresent(params, 'page_size', filters.pageSize);
  appendSearchFilters(params, filters);

  return apiGet<unknown>(`/api/v1/search?${params.toString()}`, {
    signal: options.signal,
  })
    .then(normalizeSearchResponse)
    .then((response) => hydrateSearchProducts(response, options));
}

export function autocompleteProducts(
  q: string,
  limit = 6,
  options: RequestOptions = {},
) {
  const params = new URLSearchParams();

  appendIfPresent(params, 'q', q.trim());
  appendIfPresent(params, 'limit', limit);

  return apiGet<unknown>(
    `/api/v1/search/autocomplete?${params.toString()}`,
    {
      signal: options.signal,
    },
  ).then(normalizeAutocompleteResponse);
}

function needsProductHydration(product: Product) {
  const primaryVariant = product.variants?.[0];

  return (
    !product.images?.length ||
    Boolean(primaryVariant && !primaryVariant.variant_id) ||
    primaryVariant?.available_quantity === undefined
  );
}

function mergeHydratedProduct(summary: Product, detail: Product): Product {
  return {
    ...summary,
    ...detail,
    product_id: summary.product_id,
    title: detail.title || summary.title,
  };
}

async function hydrateSearchProducts(
  response: SearchResponse,
  options: RequestOptions,
): Promise<SearchResponse> {
  const productsToHydrate = response.products.filter(needsProductHydration);

  if (productsToHydrate.length === 0) {
    return response;
  }

  const settledProducts = await Promise.allSettled(
    productsToHydrate.map((product) => getProduct(product.product_id, options)),
  );

  if (options.signal?.aborted) {
    throw new DOMException('The operation was aborted.', 'AbortError');
  }

  const hydratedById = new Map<string, Product>();
  settledProducts.forEach((result, index) => {
    const product = productsToHydrate[index];
    if (product && result.status === 'fulfilled') {
      hydratedById.set(product.product_id, result.value);
    }
  });

  if (hydratedById.size === 0) {
    return response;
  }

  return {
    ...response,
    products: response.products.map((product) => {
      const hydrated = hydratedById.get(product.product_id);

      return hydrated ? mergeHydratedProduct(product, hydrated) : product;
    }),
  };
}
