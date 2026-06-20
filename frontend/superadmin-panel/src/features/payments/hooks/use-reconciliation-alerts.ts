import { useQuery } from "@tanstack/react-query";

import { getReconciliationAlert, listReconciliationAlerts } from "../api/payments-api";
import type { ReconciliationFilters } from "../types";

export function useReconciliationAlerts(filters: ReconciliationFilters) {
  return useQuery({
    queryKey: ["admin-payment-reconciliations", filters],
    queryFn: () => listReconciliationAlerts(filters),
    staleTime: 30_000
  });
}

export function useReconciliationAlert(reconciliationId?: string | null) {
  return useQuery({
    queryKey: ["admin-payment-reconciliation", reconciliationId],
    queryFn: () => getReconciliationAlert(reconciliationId ?? ""),
    enabled: Boolean(reconciliationId),
    staleTime: 30_000
  });
}
