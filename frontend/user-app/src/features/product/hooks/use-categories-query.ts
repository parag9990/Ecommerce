import { useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { listCategories } from '../api/product.api';

export function useCategoriesQuery() {
  return useQuery({
    queryFn: ({ signal }) => listCategories({ signal }),
    queryKey: queryKeys.products.categories(),
    staleTime: 5 * 60 * 1000,
  });
}
