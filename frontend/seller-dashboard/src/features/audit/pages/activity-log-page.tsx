import { useMemo, useState } from "react";

import { PermissionDeniedState } from "../../../components/state/permission-denied-state";
import { RefreshingNotice } from "../../../components/state/refreshing-notice";
import { getSafeErrorMessage } from "../../../lib/api-error";
import { useSellerStore } from "../../../stores/seller-store";
import { AuditEmptyState } from "../components/audit-empty-state";
import { AuditErrorState } from "../components/audit-error-state";
import { AuditFilterBar } from "../components/audit-filter-bar";
import { AuditPageHeader } from "../components/audit-page-header";
import { ActivityTimeline } from "../components/activity-timeline";
import { useSellerAuditLogs } from "../hooks/use-seller-audit-logs";
import type { AuditFilters } from "../types";
import { DEFAULT_AUDIT_PAGE_SIZE } from "../types";
import { useSellerPermissions } from "../../team/hooks/use-seller-permissions";

const defaultFilters: AuditFilters = {
  page_size: DEFAULT_AUDIT_PAGE_SIZE,
};

export function ActivityLogPage() {
  const activeSeller = useSellerStore((state) => state.activeSeller);
  const permissions = useSellerPermissions();
  const [filters, setFilters] = useState<AuditFilters>(defaultFilters);
  const canViewAudit = permissions.can("audit:view");
  const auditQuery = useSellerAuditLogs(
    activeSeller?.seller_id,
    filters,
    canViewAudit,
  );

  const logs = useMemo(
    () => auditQuery.data?.pages.flatMap((page) => page.logs) ?? [],
    [auditQuery.data],
  );
  const hasFilters = useMemo(
    () =>
      Boolean(
        filters.actor_id ||
          filters.action ||
          filters.resource_type ||
          filters.resource_id ||
          filters.from ||
          filters.to,
      ),
    [filters],
  );

  if (!activeSeller) {
    return (
      <PermissionDeniedState
        title="Active seller unavailable"
        description="Audit activity load karne ke liye active seller context required hai."
      />
    );
  }

  if (!canViewAudit) {
    return (
      <PermissionDeniedState
        title="Audit access unavailable"
        description="Aapke current seller role ke paas audit activity dekhne ka permission nahi hai."
      />
    );
  }

  return (
    <section className="space-y-4">
      <AuditPageHeader
        sellerName={activeSeller.display_name}
        isRefreshing={auditQuery.isFetching && !auditQuery.isFetchingNextPage}
        onRefresh={() => auditQuery.refetch()}
      />

      <AuditFilterBar
        filters={filters}
        onApply={setFilters}
        onReset={() => setFilters(defaultFilters)}
      />

      <RefreshingNotice
        show={auditQuery.isFetching && !auditQuery.isPending && !auditQuery.isError}
      />

      {auditQuery.isError && logs.length > 0 ? (
        <RefreshingNotice
          show
          failed
          message={getSafeErrorMessage(auditQuery.error)}
          onRetry={() => auditQuery.refetch()}
        />
      ) : null}

      {auditQuery.isError && logs.length === 0 ? (
        <AuditErrorState error={auditQuery.error} onRetry={() => auditQuery.refetch()} />
      ) : null}

      {!auditQuery.isError && auditQuery.isPending ? (
        <ActivityTimeline
          logs={[]}
          isLoading
          hasNextPage={false}
          isFetchingNextPage={false}
          onLoadMore={() => undefined}
        />
      ) : null}

      {!auditQuery.isError && !auditQuery.isPending && logs.length === 0 ? (
        <AuditEmptyState hasFilters={hasFilters} />
      ) : null}

      {logs.length > 0 ? (
        <ActivityTimeline
          logs={logs}
          isLoading={false}
          hasNextPage={Boolean(auditQuery.hasNextPage)}
          isFetchingNextPage={auditQuery.isFetchingNextPage}
          onLoadMore={() => auditQuery.fetchNextPage()}
        />
      ) : null}
    </section>
  );
}
