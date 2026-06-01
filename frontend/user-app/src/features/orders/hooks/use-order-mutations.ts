import { useMutation, useQueryClient } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { cancelOrder } from '../api/orders.api';

export function useOrderMutations(orderId?: string) {
  const queryClient = useQueryClient();

  const cancel = useMutation({
    mutationFn: (reason: string) => cancelOrder(orderId ?? '', reason),
    onSuccess: (order) => {
      queryClient.setQueryData(
        queryKeys.orders.detail(order.order_id ?? orderId ?? ''),
        order,
      );
      void queryClient.invalidateQueries({ queryKey: queryKeys.orders.all });
    },
  });

  return { cancel };
}
