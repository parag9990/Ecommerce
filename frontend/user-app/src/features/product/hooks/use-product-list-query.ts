import { keepPreviousData, useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { listProducts, searchProducts } from '../api/product.api';
import type { ProductFilters } from '../types';

type ProductListMode = 'home' | 'search';

export function useProductListQuery(
  filters: ProductFilters,
  mode: ProductListMode,
) {
  return useQuery({
    placeholderData: keepPreviousData,
    queryFn: ({ signal }) =>
      mode === 'home'
        ? listProducts(
            {
              categoryId: filters.categoryId,
              page: filters.page,
              pageSize: filters.pageSize,
            },
            { signal },
          )
        : searchProducts(filters, { signal }),
    queryKey: queryKeys.products.list(filters, mode),
    staleTime: 60 * 1000,
  });
}
