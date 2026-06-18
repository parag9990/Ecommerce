import type { ProductFilters } from '../types';
import { useProductListQuery } from './use-product-list-query';

type ProductListMode = 'home' | 'search';

export function useProductList(
  filters: ProductFilters,
  mode: ProductListMode,
) {
  const query = useProductListQuery(filters, mode);

  return {
    data: query.data,
    error:
      query.error instanceof Error
        ? query.error.message
        : query.isError
          ? 'Products could not be loaded.'
          : undefined,
    isFetching: query.isFetching,
    isLoading: query.isLoading,
  };
}
