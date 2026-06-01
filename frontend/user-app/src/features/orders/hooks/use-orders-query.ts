import { keepPreviousData, useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { listOrders } from '../api/orders.api';

export function useOrdersQuery(page: number, pageSize: number) {
  return useQuery({
    placeholderData: keepPreviousData,
    queryFn: ({ signal }) => listOrders({ page, page_size: pageSize }, signal),
    queryKey: queryKeys.orders.list(page, pageSize),
    staleTime: 30 * 1000,
  });
}
