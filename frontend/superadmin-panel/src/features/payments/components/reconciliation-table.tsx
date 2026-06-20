import { AlertTriangle } from "lucide-react";

import { AmountText } from "../../../components/ui/amount-text";
import { DataState } from "../../../components/ui/data-state";
import { RiskBadge } from "../../../components/ui/risk-badge";
import { StatusBadge } from "../../../components/ui/status-badge";
import { TablePagination } from "../../../components/ui/table-pagination";
import { formatDateTime } from "../../../lib/format";
import type { ReconciliationAlert } from "../types";

function alertRisk(alert: ReconciliationAlert): "high" | "medium" | "low" {
  if (alert.status === "missing_provider" || alert.status === "missing_local") {
    return "high";
  }

  if (alert.status === "mismatch") {
    return "medium";
  }

  return "low";
}

export function ReconciliationTable({
  alerts,
  isLoading,
  isFetching,
  error,
  page,
  limit,
  totalCount,
  compact = false,
  onPageChange,
  onRetry,
  onSelect
}: {
  alerts: ReconciliationAlert[];
  isLoading: boolean;
  isFetching: boolean;
  error: unknown;
  page: number;
  limit: number;
  totalCount?: number;
  compact?: boolean;
  onPageChange: (page: number) => void;
  onRetry: () => void;
  onSelect: (alert: ReconciliationAlert) => void;
}) {
  if (isLoading) {
    return <DataState title="Loading reconciliation alerts" description="Fetching provider mismatch records." />;
  }

  if (error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load reconciliation alerts"
        description={error instanceof Error ? error.message : "The reconciliation request failed."}
        action={
          <button
            type="button"
            onClick={onRetry}
            className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
          >
            Retry
          </button>
        }
      />
    );
  }

  if (alerts.length === 0) {
    return <DataState title="No reconciliation alerts" description="No provider/local mismatches are visible." />;
  }

  return (
    <div className="overflow-hidden rounded-lg border border-slate-200 bg-white">
      <div className="overflow-auto">
        <table className="min-w-full border-separate border-spacing-0 text-left text-sm">
          <thead className="sticky top-0 bg-slate-100 text-xs uppercase text-slate-600">
            <tr>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Alert</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Status</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Risk</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Provider</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Payment</th>
              {!compact ? (
                <>
                  <th className="border-b border-slate-200 px-4 py-3 font-semibold">Local</th>
                  <th className="border-b border-slate-200 px-4 py-3 font-semibold">Provider Amount</th>
                </>
              ) : null}
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Detected</th>
              <th className="border-b border-slate-200 px-4 py-3 text-right font-semibold">Action</th>
            </tr>
          </thead>
          <tbody>
            {alerts.map((alert) => (
              <tr key={alert.reconciliation_id} className="hover:bg-slate-50">
                <td className="border-b border-slate-100 px-4 py-3">
                  <div className="flex min-w-0 items-center gap-3">
                    <span className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                      <AlertTriangle size={17} aria-hidden="true" />
                    </span>
                    <div className="min-w-0">
                      <div className="max-w-[220px] truncate font-medium text-slate-950">
                        {alert.reconciliation_id}
                      </div>
                      <div className="max-w-[220px] truncate text-xs text-slate-500">
                        {alert.settlement_id ?? "No settlement id"}
                      </div>
                    </div>
                  </div>
                </td>
                <td className="border-b border-slate-100 px-4 py-3">
                  <StatusBadge status={alert.status} />
                </td>
                <td className="border-b border-slate-100 px-4 py-3">
                  <RiskBadge level={alertRisk(alert)} />
                </td>
                <td className="border-b border-slate-100 px-4 py-3 uppercase text-slate-700">
                  {alert.provider}
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                  {alert.payment_id ?? "Not linked"}
                </td>
                {!compact ? (
                  <>
                    <td className="border-b border-slate-100 px-4 py-3 font-medium text-slate-950">
                      <AmountText money={alert.local_amount} />
                    </td>
                    <td className="border-b border-slate-100 px-4 py-3 font-medium text-slate-950">
                      <AmountText money={alert.provider_amount} />
                    </td>
                  </>
                ) : null}
                <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                  {formatDateTime(alert.detected_at)}
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-right">
                  <button
                    type="button"
                    onClick={() => onSelect(alert)}
                    className="font-medium text-blue-700 hover:text-blue-900"
                  >
                    Inspect
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {!compact ? (
        <TablePagination
          page={page}
          limit={limit}
          itemCount={alerts.length}
          totalCount={totalCount}
          isFetching={isFetching}
          onPageChange={onPageChange}
        />
      ) : null}
    </div>
  );
}
