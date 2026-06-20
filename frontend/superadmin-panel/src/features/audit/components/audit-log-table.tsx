import { Eye, RefreshCw } from "lucide-react";

import { DataState } from "../../../components/ui/data-state";
import { TablePagination } from "../../../components/ui/table-pagination";
import { cn } from "../../../lib/classnames";
import { formatDateTime } from "../../../lib/format";
import { AuditStatusBadge, isHighRiskAuditAction } from "./audit-status-badge";
import type { AdminAuditLog } from "../types";

function utcTitle(value: string | null): string | undefined {
  if (!value) {
    return undefined;
  }

  const date = new Date(value);

  return Number.isNaN(date.getTime()) ? undefined : date.toISOString();
}

export function AuditLogTable({
  logs,
  isLoading,
  isFetching,
  error,
  page,
  pageSize,
  totalCount,
  onPageChange,
  onRetry,
  onSelect
}: {
  logs: AdminAuditLog[];
  isLoading: boolean;
  isFetching: boolean;
  error: unknown;
  page: number;
  pageSize: number;
  totalCount?: number;
  onPageChange: (page: number) => void;
  onRetry: () => void;
  onSelect: (log: AdminAuditLog) => void;
}) {
  if (isLoading) {
    return <DataState title="Loading audit logs" description="Fetching admin action records." />;
  }

  if (error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load audit logs"
        description={error instanceof Error ? error.message : "The audit log request failed."}
        action={
          <button
            type="button"
            onClick={onRetry}
            className="inline-flex h-9 items-center gap-2 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
          >
            <RefreshCw className="h-4 w-4" aria-hidden="true" />
            Retry
          </button>
        }
      />
    );
  }

  if (logs.length === 0) {
    return <DataState title="No audit logs found" description="Try a different actor, resource, request, or date filter." />;
  }

  return (
    <div className="min-h-0 flex-1 overflow-hidden rounded-lg border border-slate-200 bg-white">
      <div className="min-h-0 overflow-auto">
        <table className="min-w-full border-separate border-spacing-0 text-left text-sm">
          <thead className="sticky top-0 bg-slate-100 text-xs uppercase text-slate-600">
            <tr>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Time</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Actor</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Action</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Resource</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Request ID</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">IP Hash</th>
              <th className="border-b border-slate-200 px-4 py-3 text-right font-semibold">Details</th>
            </tr>
          </thead>
          <tbody>
            {logs.map((log) => {
              const highRisk = isHighRiskAuditAction(log.action);

              return (
                <tr
                  key={log.id}
                  className={cn(
                    "hover:bg-slate-50",
                    highRisk && "bg-red-50/40 hover:bg-red-50"
                  )}
                >
                  <td className="whitespace-nowrap border-b border-slate-100 px-4 py-3 text-slate-700">
                    <time dateTime={log.created_at ?? undefined} title={utcTitle(log.created_at)}>
                      {formatDateTime(log.created_at)}
                    </time>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3">
                    <div className="max-w-[220px] truncate font-medium text-slate-950" title={log.actor_admin_id}>
                      {log.actor_admin_id}
                    </div>
                    <div className="mt-1 text-xs text-slate-500">{log.actor_role}</div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3">
                    <AuditStatusBadge action={log.action} />
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3">
                    <div className="font-medium text-slate-950">{log.resource_type}</div>
                    <div
                      className="mt-1 max-w-[220px] truncate font-mono text-xs text-slate-500"
                      title={log.resource_id}
                    >
                      {log.resource_id}
                    </div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3">
                    <div className="max-w-[220px] select-all truncate font-mono text-xs text-slate-600" title={log.request_id}>
                      {log.request_id}
                    </div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3">
                    <div className="max-w-[180px] select-all truncate font-mono text-xs text-slate-600" title={log.ip_hash}>
                      {log.ip_hash}
                    </div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3 text-right">
                    <button
                      type="button"
                      onClick={() => onSelect(log)}
                      className="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-slate-300 px-3 font-medium text-slate-700 hover:bg-slate-50"
                    >
                      <Eye className="h-4 w-4" aria-hidden="true" />
                      View
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <TablePagination
        page={page}
        limit={pageSize}
        itemCount={logs.length}
        totalCount={totalCount}
        isFetching={isFetching}
        onPageChange={onPageChange}
      />
    </div>
  );
}
