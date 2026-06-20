import { useQuery } from "@tanstack/react-query";

import { listAdminAuditLogs } from "../api/admin-audit-api";
import type { AuditLogFilters } from "../types";

export function useAdminAuditLogs(filters: AuditLogFilters, enabled = true) {
  return useQuery({
    queryKey: ["admin", "audit-logs", "list", filters],
    queryFn: () => listAdminAuditLogs(filters),
    enabled,
    placeholderData: (previous) => previous,
    staleTime: 30_000
  });
}
