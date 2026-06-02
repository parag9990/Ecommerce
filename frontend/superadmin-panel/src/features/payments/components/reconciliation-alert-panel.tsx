import { X } from "lucide-react";

import { AmountText } from "../../../components/ui/amount-text";
import { DataState } from "../../../components/ui/data-state";
import { RiskBadge } from "../../../components/ui/risk-badge";
import { StatusBadge } from "../../../components/ui/status-badge";
import { formatDateTime } from "../../../lib/format";
import { useReconciliationAlert } from "../hooks/use-reconciliation-alerts";

function riskForStatus(status: string): "high" | "medium" | "low" {
  if (status === "missing_local" || status === "missing_provider") {
    return "high";
  }

  if (status === "mismatch") {
    return "medium";
  }

  return "low";
}

export function ReconciliationAlertPanel({
  reconciliationId,
  onClose
}: {
  reconciliationId: string | null;
  onClose: () => void;
}) {
  const detailQuery = useReconciliationAlert(reconciliationId);

  if (!reconciliationId) {
    return null;
  }

  const alert = detailQuery.data?.alert;

  return (
    <div className="fixed inset-0 z-40">
      <button
        type="button"
        className="absolute inset-0 bg-slate-950/30"
        aria-label="Close reconciliation alert"
        onClick={onClose}
      />
      <aside className="absolute inset-y-0 right-0 flex w-full max-w-xl flex-col overflow-hidden border-l border-slate-200 bg-white shadow-xl">
        <header className="border-b border-slate-200 px-4 py-4">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h2 className="text-lg font-semibold text-slate-950">Reconciliation Alert</h2>
              <p className="mt-1 break-all text-sm text-slate-600">{reconciliationId}</p>
            </div>
            <button
              type="button"
              onClick={onClose}
              title="Close reconciliation alert"
              className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-100"
            >
              <X className="h-4 w-4" aria-hidden="true" />
              <span className="sr-only">Close reconciliation alert</span>
            </button>
          </div>
        </header>

        <div className="min-h-0 flex-1 overflow-auto p-4">
          {detailQuery.isLoading ? (
            <DataState title="Loading reconciliation detail" description="Fetching mismatch context." />
          ) : null}

          {detailQuery.error ? (
            <DataState
              tone="danger"
              title="Unable to load reconciliation detail"
              description={
                detailQuery.error instanceof Error
                  ? detailQuery.error.message
                  : "The reconciliation detail request failed."
              }
              action={
                <button
                  type="button"
                  onClick={() => void detailQuery.refetch()}
                  className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
                >
                  Retry
                </button>
              }
            />
          ) : null}

          {alert ? (
            <section className="rounded-lg border border-slate-200 bg-slate-50 p-4">
              <dl className="grid gap-3 text-sm">
                <div className="flex items-center justify-between gap-3">
                  <dt className="text-slate-600">Status</dt>
                  <dd>
                    <StatusBadge status={alert.status} />
                  </dd>
                </div>
                <div className="flex items-center justify-between gap-3">
                  <dt className="text-slate-600">Risk</dt>
                  <dd>
                    <RiskBadge level={riskForStatus(alert.status)} />
                  </dd>
                </div>
                <div className="flex justify-between gap-3">
                  <dt className="text-slate-600">Provider</dt>
                  <dd className="uppercase text-slate-700">{alert.provider}</dd>
                </div>
                <div className="flex justify-between gap-3">
                  <dt className="text-slate-600">Payment</dt>
                  <dd className="break-all text-right text-slate-700">{alert.payment_id ?? "Not linked"}</dd>
                </div>
                <div className="flex justify-between gap-3">
                  <dt className="text-slate-600">Settlement</dt>
                  <dd className="break-all text-right text-slate-700">
                    {alert.settlement_id ?? "Not linked"}
                  </dd>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div className="border-l border-slate-200 pl-3">
                    <dt className="text-slate-600">Local amount</dt>
                    <dd className="mt-1 font-medium text-slate-950">
                      <AmountText money={alert.local_amount} />
                    </dd>
                    <dt className="mt-3 text-slate-600">Local status</dt>
                    <dd className="mt-1 text-slate-700">{alert.local_status ?? "Unknown"}</dd>
                  </div>
                  <div className="border-l border-slate-200 pl-3">
                    <dt className="text-slate-600">Provider amount</dt>
                    <dd className="mt-1 font-medium text-slate-950">
                      <AmountText money={alert.provider_amount} />
                    </dd>
                    <dt className="mt-3 text-slate-600">Provider status</dt>
                    <dd className="mt-1 text-slate-700">{alert.provider_status ?? "Unknown"}</dd>
                  </div>
                </div>
                <div className="flex justify-between gap-3">
                  <dt className="text-slate-600">Detected</dt>
                  <dd className="text-right text-slate-700">{formatDateTime(alert.detected_at)}</dd>
                </div>
                <div className="grid gap-1">
                  <dt className="text-slate-600">Note</dt>
                  <dd className="rounded-lg border border-slate-200 bg-white px-3 py-2 text-slate-800">
                    {alert.note ?? "No reconciliation note was returned."}
                  </dd>
                </div>
              </dl>
            </section>
          ) : null}
        </div>
      </aside>
    </div>
  );
}
