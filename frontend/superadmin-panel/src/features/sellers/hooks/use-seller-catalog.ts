import { useQuery } from "@tanstack/react-query";

import { listSellerCatalog } from "../api/sellers-api";

export function useSellerCatalog(sellerId: string, enabled = true) {
  return useQuery({
    queryKey: ["seller-catalog", sellerId],
    queryFn: () => listSellerCatalog(sellerId),
    enabled: Boolean(sellerId) && enabled,
    staleTime: 30_000
  });
}
