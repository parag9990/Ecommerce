import { History } from "lucide-react";

import type { StatusHistoryEntry } from "../types";
import { formatDateTime, formatOrderStatus } from "../utils/order-formatters";
import { OrderStatusBadge } from "./order-status-badge";

type StatusHistoryListProps = {
  history: StatusHistoryEntry[];
};

export function StatusHistoryList({ history }: StatusHistoryListProps) {
  return (
    <section className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex items-center gap-2">
        <History className="h-4 w-4 text-slate-500" aria-hidden="true" />
        <h2 className="text-sm font-semibold text-slate-950">Status history</h2>
      </div>

      {history.length === 0 ? (
        <p className="mt-3 text-sm text-slate-500">No status history available.</p>
      ) : (
        <ol className="mt-3 divide-y divide-slate-100">
          {history.map((entry, index) => (
            <li
              key={`${entry.status}-${entry.created_at || index}`}
              className="flex items-start justify-between gap-3 py-3 text-sm"
            >
              <div className="min-w-0">
                <OrderStatusBadge status={entry.status} />
                {entry.note ? (
                  <p className="mt-1 text-slate-600">{entry.note}</p>
                ) : (
                  <p className="mt-1 text-slate-500 capitalize">
                    {formatOrderStatus(entry.status)}
                  </p>
                )}
              </div>
              <time className="shrink-0 text-xs text-slate-500">
                {formatDateTime(entry.created_at)}
              </time>
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}
