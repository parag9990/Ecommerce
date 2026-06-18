import { apiGet } from '../../../lib/http';
import type {
  AutocompleteResponse,
  CategoryListResponse,
  Product,
  ProductFilters,
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

function appendSearchFilter(
  params: URLSearchParams,
  key: string,
  value?: boolean | number | string,
) {
  if (value === undefined || value === null || value === '') {
    return;
  }

  params.append('filter', `${key}:${String(value)}`);
}

function appendSearchFilters(params: URLSearchParams, filters: ProductFilters) {
  appendSearchFilter(params, 'category_ids', filters.categoryId);
  appendSearchFilter(params, 'brand', filters.brand);
  appendSearchFilter(params, 'seller_id', filters.sellerId);

  if (filters.minPrice) {
    params.append('filter', `price:>=${filters.minPrice}`);
  }

  if (filters.maxPrice) {
    params.append('filter', `price:<=${filters.maxPrice}`);
  }

  if (filters.minRating) {
    params.append('filter', `rating:>=${filters.minRating}`);
  }

  if (filters.inStock) {
    appendSearchFilter(params, 'in_stock', true);
  }
}

export function listProducts(
  input: { categoryId?: string | undefined; page: number; pageSize: number },
  options: RequestOptions = {},
) {
  const params = new URLSearchParams();

  appendIfPresent(params, 'page', input.page);
  appendIfPresent(params, 'page_size', input.pageSize);
  appendIfPresent(params, 'category_id', input.categoryId);
  appendIfPresent(params, 'status', 'published');

  return apiGet<ProductListResponse>(`/api/v1/products?${params.toString()}`, {
    signal: options.signal,
  });
}

export function getProduct(productId: string, options: RequestOptions = {}) {
  return apiGet<Product>(`/api/v1/products/${encodeURIComponent(productId)}`, {
    signal: options.signal,
  });
}

export function listCategories(options: RequestOptions = {}) {
  return apiGet<CategoryListResponse>('/api/v1/categories', {
    signal: options.signal,
  });
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

  return apiGet<SearchResponse>(`/api/v1/search?${params.toString()}`, {
    signal: options.signal,
  });
}

export function autocompleteProducts(
  q: string,
  limit = 6,
  options: RequestOptions = {},
) {
  const params = new URLSearchParams();

  appendIfPresent(params, 'q', q.trim());
  appendIfPresent(params, 'limit', limit);

  return apiGet<AutocompleteResponse>(
    `/api/v1/search/autocomplete?${params.toString()}`,
    {
      signal: options.signal,
    },
  );
}
