import { useQuery } from "@tanstack/react-query";

import { listAdminOrders } from "../api/orders-api";
import type { AdminOrderFilters } from "../types";

export function useAdminOrders(filters: AdminOrderFilters) {
  return useQuery({
    queryKey: ["admin-orders", filters],
    queryFn: () => listAdminOrders(filters),
    staleTime: 30_000
  });
}
