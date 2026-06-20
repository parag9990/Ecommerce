import { useQuery } from "@tanstack/react-query";

import { getAdminOrder } from "../api/orders-api";

export function useAdminOrder(orderId: string) {
  return useQuery({
    queryKey: ["admin-order", orderId],
    queryFn: () => getAdminOrder(orderId),
    enabled: Boolean(orderId),
    staleTime: 30_000
  });
}
