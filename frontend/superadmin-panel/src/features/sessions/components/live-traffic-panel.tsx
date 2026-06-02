import { RefreshCw } from "lucide-react";

import { DataState } from "../../../components/ui/data-state";
import { formatDateTime } from "../../../lib/format";
import { LiveMetricsCards } from "./live-metrics-cards";
import type { LiveMetrics } from "../types";

export function LiveTrafficPanel({
  metrics,
  isLoading,
  isFetching,
  error,
  onRetry
}: {
  metrics?: LiveMetrics;
  isLoading: boolean;
  isFetching: boolean;
  error: unknown;
  onRetry: () => void;
}) {
  if (error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load live traffic"
        description={error instanceof Error ? error.message : "The live metrics request failed."}
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

  return (
    <section className="space-y-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-slate-950">Live Traffic</h2>
          <p className="text-sm text-slate-600">Updated {formatDateTime(metrics?.updated_at)}</p>
        </div>
        {isFetching && !isLoading ? (
          <span className="inline-flex h-8 items-center gap-2 rounded-lg border border-slate-200 bg-white px-2 text-xs font-medium text-slate-600">
            <RefreshCw className="h-3.5 w-3.5" aria-hidden="true" />
            Refreshing
          </span>
        ) : null}
      </div>
      <LiveMetricsCards metrics={metrics} isLoading={isLoading} />
    </section>
  );
}
