import type { AuditFilters } from "../types";

export const auditQueryKeys = {
  all: ["seller-audit"] as const,
  logs: (sellerId: string, filters: AuditFilters) =>
    [...auditQueryKeys.all, sellerId, filters] as const,
};
