import { AlertTriangle, RefreshCcw } from "lucide-react";

import type { DeletionRequest } from "../../../api/session-api";
import { formatDateTime, formatNumber } from "../../../lib/format";
import { useDeletionRequests } from "../hooks/use-deletion-requests";

export function DeletionStatusTable() {
  const requests = useDeletionRequests();

  if (requests.isPending) {
    return (
      <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, index) => (
            <div
              aria-hidden="true"
              className="h-12 animate-pulse rounded-md bg-zinc-100"
              key={index}
            />
          ))}
        </div>
      </section>
    );
  }

  if (requests.isError) {
    return (
      <section className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-900">
        <div className="flex gap-3">
          <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
          <div className="min-w-0">
            <h2 className="text-sm font-semibold">Deletion status unavailable</h2>
            <p className="mt-1 text-sm text-red-800">
              Recent deletion requests could not be loaded.
            </p>
            <button
              className="mt-3 inline-flex h-9 items-center gap-2 rounded-md border border-red-300 bg-white px-3 text-sm font-medium text-red-900 transition-colors hover:bg-red-100"
              onClick={() => void requests.refetch()}
              type="button"
            >
              <RefreshCcw className="h-4 w-4" aria-hidden="true" />
              Retry
            </button>
          </div>
        </div>
      </section>
    );
  }

  const items = requests.data.items;

  return (
    <section className="overflow-hidden rounded-lg border border-zinc-200 bg-white shadow-panel">
      <div className="border-b border-zinc-200 px-4 py-3">
        <h2 className="text-sm font-semibold text-zinc-950">
          Recent deletion requests
        </h2>
        <p className="mt-1 text-xs text-zinc-500">
          Raw target values are never echoed back into this table.
        </p>
      </div>

      {items.length === 0 ? (
        <div className="p-4 text-sm text-zinc-500">
          No deletion requests have been submitted yet.
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-zinc-200 text-left">
            <thead className="bg-zinc-50 text-xs uppercase text-zinc-500">
              <tr>
                <th className="px-4 py-3 font-semibold" scope="col">
                  Target
                </th>
                <th className="px-4 py-3 font-semibold" scope="col">
                  Status
                </th>
                <th className="px-4 py-3 font-semibold" scope="col">
                  Impact
                </th>
                <th className="px-4 py-3 font-semibold" scope="col">
                  Requested
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-200">
              {items.map((item) => (
                <DeletionStatusRow item={item} key={item.requestId} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

function DeletionStatusRow({ item }: { item: DeletionRequest }) {
  return (
    <tr className="hover:bg-zinc-50">
      <td className="min-w-48 px-4 py-3 align-top">
        <div className="font-mono text-xs font-medium text-zinc-950">
          {item.targetValueMasked}
        </div>
        <div className="mt-1 text-xs text-zinc-500">{item.targetType}</div>
      </td>
      <td className="min-w-32 px-4 py-3 align-top">
        <StatusBadge status={item.status} />
        {item.error ? (
          <div className="mt-1 max-w-48 truncate text-xs text-red-600">
            {item.error}
          </div>
        ) : null}
      </td>
      <td className="min-w-44 px-4 py-3 align-top text-xs text-zinc-600">
        <div>{formatNumber(item.matchedSessions)} sessions</div>
        <div>{formatNumber(item.matchedEvents)} events</div>
      </td>
      <td className="min-w-44 px-4 py-3 align-top text-xs text-zinc-600">
        <div>{formatDateTime(item.createdAt)}</div>
        <div className="mt-1 truncate">by {item.requestedBy}</div>
      </td>
    </tr>
  );
}

function StatusBadge({ status }: { status: DeletionRequest["status"] }) {
  const classes: Record<DeletionRequest["status"], string> = {
    completed: "border-emerald-200 bg-emerald-50 text-emerald-700",
    failed: "border-red-200 bg-red-50 text-red-700",
    processing: "border-sky-200 bg-sky-50 text-sky-700",
    queued: "border-amber-200 bg-amber-50 text-amber-700"
  };

  return (
    <span
      className={[
        "inline-flex h-7 items-center rounded-md border px-2.5 text-xs font-medium capitalize",
        classes[status]
      ].join(" ")}
    >
      {status}
    </span>
  );
}
