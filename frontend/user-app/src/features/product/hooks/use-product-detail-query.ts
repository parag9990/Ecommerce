import { useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { getProduct } from '../api/product.api';

export function useProductDetailQuery(productId?: string) {
  return useQuery({
    enabled: Boolean(productId),
    queryFn: ({ signal }) => getProduct(productId ?? '', { signal }),
    queryKey: queryKeys.products.detail(productId ?? ''),
    staleTime: 2 * 60 * 1000,
  });
}
