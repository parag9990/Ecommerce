import { AlertTriangle } from "lucide-react";

import { DataState } from "../../../components/ui/data-state";
import { StatusBadge } from "../../../components/ui/status-badge";
import { formatDateTime } from "../../../lib/format";
import { useOrderDisputes } from "../hooks/use-order-disputes";

export function DisputePanel({ orderId, canView }: { orderId: string; canView: boolean }) {
  const disputesQuery = useOrderDisputes(orderId, canView);
  const disputes = disputesQuery.data?.disputes ?? [];

  if (!canView) {
    return (
      <DataState
        title="Disputes hidden"
        description="Your role can view the order, but dispute details are not available."
      />
    );
  }

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="flex items-center gap-2">
        <AlertTriangle className="h-4 w-4 text-orange-600" aria-hidden="true" />
        <h2 className="text-base font-semibold text-slate-950">Disputes</h2>
      </div>

      {disputesQuery.isLoading ? (
        <p className="mt-3 text-sm text-slate-600">Loading disputes.</p>
      ) : null}

      {disputesQuery.error ? (
        <DataState
          tone="danger"
          title="Unable to load disputes"
          description={
            disputesQuery.error instanceof Error
              ? disputesQuery.error.message
              : "The dispute request failed."
          }
          action={
            <button
              type="button"
              onClick={() => void disputesQuery.refetch()}
              className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
            >
              Retry
            </button>
          }
        />
      ) : null}

      {!disputesQuery.isLoading && !disputesQuery.error && disputes.length === 0 ? (
        <p className="mt-3 text-sm text-slate-600">No disputes found.</p>
      ) : null}

      {disputes.length > 0 ? (
        <div className="mt-4 space-y-3">
          {disputes.map((dispute) => (
            <article key={dispute.dispute_id} className="rounded-lg border border-slate-200 p-3 text-sm">
              <div className="flex flex-wrap items-start justify-between gap-2">
                <div className="min-w-0">
                  <div className="truncate font-medium capitalize text-slate-950">
                    {dispute.type.replace(/_/g, " ")}
                  </div>
                  <div className="mt-1 text-xs text-slate-500">{dispute.dispute_id}</div>
                </div>
                <StatusBadge status={dispute.status} />
              </div>
              <p className="mt-3 text-slate-700">{dispute.summary}</p>
              <p className="mt-3 text-xs text-slate-500">
                Opened by {dispute.opened_by} - {formatDateTime(dispute.created_at)}
              </p>
            </article>
          ))}
        </div>
      ) : null}
    </section>
  );
}
