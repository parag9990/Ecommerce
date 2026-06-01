import { productSortOptions, type ProductFilters, type SortOption } from './types';

export const DEFAULT_PAGE = 1;
export const DEFAULT_PAGE_SIZE = 24;

const MAX_PAGE_SIZE = 60;
const sortOptionSet = new Set<SortOption>(productSortOptions);

function readPositiveInt(
  value: string | null,
  fallback: number,
  maximum = Number.MAX_SAFE_INTEGER,
) {
  const parsed = Number(value);

  return Number.isInteger(parsed) && parsed > 0
    ? Math.min(parsed, maximum)
    : fallback;
}

function readTrimmedValue(value: string | null) {
  const trimmedValue = value?.trim();

  return trimmedValue && trimmedValue.length > 0 ? trimmedValue : undefined;
}

function readSort(value: string | null): SortOption | undefined {
  if (!value || !sortOptionSet.has(value as SortOption)) {
    return undefined;
  }

  return value as SortOption;
}

export function readProductFilters(searchParams: URLSearchParams): ProductFilters {
  return {
    brand: readTrimmedValue(searchParams.get('brand')),
    categoryId: readTrimmedValue(searchParams.get('category_id')),
    inStock: searchParams.get('in_stock') === 'true',
    maxPrice: readTrimmedValue(searchParams.get('max_price')),
    minPrice: readTrimmedValue(searchParams.get('min_price')),
    minRating: readTrimmedValue(searchParams.get('min_rating')),
    page: readPositiveInt(searchParams.get('page'), DEFAULT_PAGE),
    pageSize: readPositiveInt(
      searchParams.get('page_size'),
      DEFAULT_PAGE_SIZE,
      MAX_PAGE_SIZE,
    ),
    q: readTrimmedValue(searchParams.get('q')),
    sellerId: readTrimmedValue(searchParams.get('seller_id')),
    sort: readSort(searchParams.get('sort')),
  };
}

export function writeProductFilters(filters: ProductFilters) {
  const params = new URLSearchParams();

  if (filters.q) params.set('q', filters.q);
  if (filters.categoryId) params.set('category_id', filters.categoryId);
  if (filters.brand) params.set('brand', filters.brand);
  if (filters.sellerId) params.set('seller_id', filters.sellerId);
  if (filters.minPrice) params.set('min_price', filters.minPrice);
  if (filters.maxPrice) params.set('max_price', filters.maxPrice);
  if (filters.minRating) params.set('min_rating', filters.minRating);
  if (filters.inStock) params.set('in_stock', 'true');
  if (filters.sort) params.set('sort', filters.sort);
  if (filters.page > DEFAULT_PAGE) params.set('page', String(filters.page));
  if (filters.pageSize !== DEFAULT_PAGE_SIZE) {
    params.set('page_size', String(filters.pageSize));
  }

  return params;
}
