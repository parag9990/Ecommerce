import { useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { getCart } from '../api/cart.api';

type UseCartQueryOptions = {
  enabled?: boolean | undefined;
};

export function useCartQuery(options: UseCartQueryOptions = {}) {
  return useQuery({
    enabled: options.enabled ?? true,
    queryFn: ({ signal }) => getCart(signal),
    queryKey: queryKeys.cart.detail(),
    staleTime: 30 * 1000,
  });
}
