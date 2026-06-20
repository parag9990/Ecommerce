import { useMemo, useState } from "react";

import { DataState } from "../../../components/ui/data-state";
import { PermissionDenied } from "../../../components/ui/permission-denied";
import { auditExportFilename, downloadBlob } from "../audit-csv";
import { AuditExportDialog } from "../components/audit-export-dialog";
import { AuditFilterPanel, type AuditFilterPatch } from "../components/audit-filter-panel";
import { AuditLogDetailDrawer } from "../components/audit-log-detail-drawer";
import { AuditLogTable } from "../components/audit-log-table";
import { AuditPageHeader } from "../components/audit-page-header";
import { useAdminAuditLogs } from "../hooks/use-admin-audit-logs";
import { useExportAuditLogs } from "../hooks/use-export-audit-logs";
import { useAuditPermissions } from "../permissions";
import type { AdminAuditLog, AuditLogFilters } from "../types";
import { createInitialAuditLogFilters, validateAuditLogFilters } from "../validators";

function pageErrorFromValidation(filters: AuditLogFilters): Error | null {
  const validation = validateAuditLogFilters(filters);

  return validation.valid ? null : new Error(validation.errors.join(" "));
}

export function AdminAuditLogPage() {
  const permissions = useAuditPermissions();
  const [filters, setFilters] = useState<AuditLogFilters>(() => createInitialAuditLogFilters());
  const [selectedLog, setSelectedLog] = useState<AdminAuditLog | null>(null);
  const [exportOpen, setExportOpen] = useState(false);
  const filterError = useMemo(() => pageErrorFromValidation(filters), [filters]);
  const logsQuery = useAdminAuditLogs(filters, permissions.canViewAuditLogs && filterError === null);
  const exportMutation = useExportAuditLogs();
  const logs = logsQuery.data?.logs ?? [];

  function updateFilters(patch: AuditFilterPatch) {
    setFilters((current) => ({
      ...current,
      ...patch,
      page: 1
    }));
  }

  function resetFilters() {
    setFilters(createInitialAuditLogFilters());
  }

  function changePage(page: number) {
    setFilters((current) => ({
      ...current,
      page: Math.max(page, 1)
    }));
  }

  if (!permissions.canViewAuditLogs) {
    return <PermissionDenied compact />;
  }

  return (
    <section className="flex min-h-[calc(100vh-6.5rem)] flex-col overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
      <AuditPageHeader
        canExport={permissions.canExportAuditLogs}
        isExporting={exportMutation.isPending}
        onExport={() => setExportOpen(true)}
      />

      <AuditFilterPanel filters={filters} onChange={updateFilters} onReset={resetFilters} />

      {filterError ? (
        <div className="border-b border-slate-200 bg-red-50 p-4">
          <DataState tone="danger" title="Audit filters need attention" description={filterError.message} />
        </div>
      ) : null}

      <main className="min-h-0 flex-1 p-4">
        <AuditLogTable
          logs={logs}
          isLoading={logsQuery.isLoading}
          isFetching={logsQuery.isFetching}
          error={filterError ?? logsQuery.error}
          page={logsQuery.data?.page ?? filters.page}
          pageSize={logsQuery.data?.page_size ?? filters.page_size}
          totalCount={logsQuery.data?.total}
          onPageChange={changePage}
          onRetry={() => void logsQuery.refetch()}
          onSelect={setSelectedLog}
        />
      </main>

      <AuditLogDetailDrawer log={selectedLog} onClose={() => setSelectedLog(null)} />

      <AuditExportDialog
        open={exportOpen}
        filters={filters}
        canExport={permissions.canExportAuditLogs}
        isExporting={exportMutation.isPending}
        error={exportMutation.error}
        onClose={() => setExportOpen(false)}
        onExport={(reason) => {
          exportMutation.mutate(
            { filters, reason },
            {
              onSuccess: (blob) => {
                downloadBlob(auditExportFilename(), blob);
                setExportOpen(false);
              }
            }
          );
        }}
      />
    </section>
  );
}
