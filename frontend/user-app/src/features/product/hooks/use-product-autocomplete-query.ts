import { useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { autocompleteProducts } from '../api/product.api';

export function useProductAutocompleteQuery(q: string, limit = 6) {
  const normalizedQuery = q.trim();

  return useQuery({
    enabled: normalizedQuery.length >= 2,
    queryFn: ({ signal }) =>
      autocompleteProducts(normalizedQuery, limit, { signal }),
    queryKey: queryKeys.products.autocomplete(normalizedQuery, limit),
    staleTime: 60 * 1000,
  });
}
