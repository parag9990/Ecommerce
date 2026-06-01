import { useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { getOrder } from '../api/orders.api';

export function useOrderDetailQuery(orderId?: string) {
  return useQuery({
    enabled: Boolean(orderId),
    queryFn: ({ signal }) => getOrder(orderId ?? '', signal),
    queryKey: queryKeys.orders.detail(orderId ?? ''),
    staleTime: 30 * 1000,
  });
}
