import { useQuery } from "@tanstack/react-query";

import { listAdminPayments } from "../api/payments-api";
import type { PaymentFilters } from "../types";

export function useAdminPayments(filters: PaymentFilters) {
  return useQuery({
    queryKey: ["admin-payments", filters],
    queryFn: () => listAdminPayments(filters),
    staleTime: 30_000
  });
}
