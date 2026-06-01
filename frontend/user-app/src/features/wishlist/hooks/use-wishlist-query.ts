import { useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { getWishlist } from '../api/wishlist.api';

type UseWishlistQueryOptions = {
  enabled?: boolean | undefined;
};

export function useWishlistQuery(options: UseWishlistQueryOptions = {}) {
  return useQuery({
    enabled: options.enabled ?? true,
    queryFn: ({ signal }) => getWishlist(signal),
    queryKey: queryKeys.wishlist.detail(),
    staleTime: 60 * 1000,
  });
}
