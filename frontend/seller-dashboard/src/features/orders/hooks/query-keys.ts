import type { SellerOrderFilters } from "../types";

export const orderQueryKeys = {
  all: ["seller-orders"] as const,
  lists: () => [...orderQueryKeys.all, "list"] as const,
  list: (sellerId: string, filters: SellerOrderFilters) =>
    [...orderQueryKeys.lists(), sellerId, filters] as const,
  details: () => [...orderQueryKeys.all, "detail"] as const,
  detail: (orderId: string) => [...orderQueryKeys.details(), orderId] as const,
};
