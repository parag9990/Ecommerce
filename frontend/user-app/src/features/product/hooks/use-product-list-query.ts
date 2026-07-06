import { keepPreviousData, useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { listProducts, searchProducts } from '../api/product.api';
import type { ProductFilters } from '../types';

type ProductListMode = 'home' | 'search';

function hasSearchOnlyFilters(filters: ProductFilters) {
  return Boolean(
    filters.q ||
      filters.brand ||
      filters.sellerId ||
      filters.minPrice ||
      filters.maxPrice ||
      filters.minRating ||
      filters.inStock,
  );
}

function canUseCatalogList(filters: ProductFilters, mode: ProductListMode) {
  if (mode === 'home') {
    return true;
  }

  if (hasSearchOnlyFilters(filters)) {
    return false;
  }

  return filters.sort !== 'popular';
}

export function useProductListQuery(
  filters: ProductFilters,
  mode: ProductListMode,
) {
  return useQuery({
    placeholderData: keepPreviousData,
    queryFn: ({ signal }) =>
      canUseCatalogList(filters, mode)
        ? listProducts(
            {
              categoryId: filters.categoryId,
              page: filters.page,
              pageSize: filters.pageSize,
              sort: filters.sort,
            },
            { signal },
          )
        : searchProducts(filters, { signal }),
    queryKey: queryKeys.products.list(filters, mode),
    staleTime: 60 * 1000,
  });
}
