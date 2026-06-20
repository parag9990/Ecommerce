import { useQuery } from "@tanstack/react-query";

import { listOrderDisputes } from "../api/orders-api";

export function useOrderDisputes(orderId: string, enabled = true) {
  return useQuery({
    queryKey: ["order-disputes", orderId],
    queryFn: () => listOrderDisputes(orderId),
    enabled: Boolean(orderId) && enabled,
    staleTime: 30_000
  });
}
