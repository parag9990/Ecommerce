import { useQuery } from "@tanstack/react-query";

import { listRefunds } from "../api/payments-api";
import type { RefundFilters } from "../types";

export function useRefunds(filters: RefundFilters) {
  return useQuery({
    queryKey: ["admin-refunds", filters],
    queryFn: () => listRefunds(filters),
    staleTime: 30_000
  });
}
