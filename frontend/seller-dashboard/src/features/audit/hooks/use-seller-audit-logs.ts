import { useInfiniteQuery } from "@tanstack/react-query";

import { listSellerAuditLogs } from "../api/seller-audit-api";
import type { AuditFilters } from "../types";
import { DEFAULT_AUDIT_PAGE_SIZE } from "../types";
import { auditQueryKeys } from "./query-keys";

export function useSellerAuditLogs(
  sellerId: string | undefined,
  filters: AuditFilters,
  enabled = true,
) {
  return useInfiniteQuery({
    queryKey: auditQueryKeys.logs(sellerId ?? "", filters),
    queryFn: ({ pageParam }) =>
      listSellerAuditLogs({
        ...filters,
        page: Number(pageParam),
        page_size: filters.page_size ?? DEFAULT_AUDIT_PAGE_SIZE,
      }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined,
    enabled: Boolean(sellerId) && enabled,
    staleTime: 30_000,
  });
}
