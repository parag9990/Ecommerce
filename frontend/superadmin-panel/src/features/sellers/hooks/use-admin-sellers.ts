import { useQuery } from "@tanstack/react-query";

import { listAdminSellers } from "../api/sellers-api";
import type { AdminSellerFilters } from "../types";

export function useAdminSellers(filters: AdminSellerFilters) {
  return useQuery({
    queryKey: ["admin-sellers", filters],
    queryFn: () => listAdminSellers(filters),
    staleTime: 30_000
  });
}
