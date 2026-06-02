import { keepPreviousData, useQuery } from "@tanstack/react-query";

import { listSellerOrders } from "../api/seller-order-api";
import type { SellerOrderFilters } from "../types";
import { orderQueryKeys } from "./query-keys";

export function useSellerOrders(filters: SellerOrderFilters, sellerId?: string) {
  return useQuery({
    queryKey: orderQueryKeys.list(sellerId ?? "", filters),
    queryFn: () => listSellerOrders(filters),
    enabled: Boolean(sellerId),
    placeholderData: keepPreviousData,
  });
}
