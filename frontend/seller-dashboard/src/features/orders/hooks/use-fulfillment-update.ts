import { useMutation, useQueryClient } from "@tanstack/react-query";

import { updateOrderFulfillment } from "../api/seller-order-api";
import type { FulfillmentUpdateInput, Order } from "../types";
import { orderQueryKeys } from "./query-keys";

type FulfillmentUpdateOptions = {
  onSuccess?: (order: Order) => void;
};

export function useFulfillmentUpdate(
  orderId: string,
  options: FulfillmentUpdateOptions = {},
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: FulfillmentUpdateInput) =>
      updateOrderFulfillment(orderId, input),
    onSuccess: (order) => {
      queryClient.invalidateQueries({ queryKey: orderQueryKeys.lists() });
      queryClient.setQueryData(orderQueryKeys.detail(order.order_id), order);
      options.onSuccess?.(order);
    },
  });
}
