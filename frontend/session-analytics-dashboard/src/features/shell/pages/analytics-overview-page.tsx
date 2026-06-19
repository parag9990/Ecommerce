import { AlertTriangle, RefreshCcw } from "lucide-react";
import { useMemo, useState } from "react";

import { FiltersBar } from "../../../layout/filters-bar";
import {
  getDefaultDateRange,
  validateDateRange,
  type DateRange
} from "../../../lib/date-range";
import {
  isLiveMetricsEmpty,
  MetricCardGrid
} from "../components/metric-card-grid";
import { useLiveMetrics } from "../hooks/use-live-metrics";
import {
  defaultSegmentFilters,
  type SegmentFilters
} from "../types";

export function AnalyticsOverviewPage() {
  const [dateRange, setDateRange] = useState<DateRange>(() =>
    getDefaultDateRange()
  );
  const [filters, setFilters] = useState<SegmentFilters>(defaultSegmentFilters);

  const dateRangeValidation = validateDateRange(dateRange);
  const request = useMemo(
    () => ({
      dateRange,
      filters
    }),
    [dateRange, filters]
  );

  const metricsQuery = useLiveMetrics(request, {
    enabled: dateRangeValidation.ok
  });

  return (
    <div>
      <FiltersBar
        dateRange={dateRange}
        dateRangeError={
          dateRangeValidation.ok ? undefined : dateRangeValidation.message
        }
        filters={filters}
        onDateRangeChange={setDateRange}
        onFiltersChange={setFilters}
      />

      <div className="p-4 lg:p-6">
        {!dateRangeValidation.ok ? (
          <ValidationState message={dateRangeValidation.message} />
        ) : null}

        {dateRangeValidation.ok && metricsQuery.isPending ? (
          <MetricsLoadingState />
        ) : null}

        {dateRangeValidation.ok && metricsQuery.isError ? (
          <MetricsErrorState onRetry={() => void metricsQuery.refetch()} />
        ) : null}

        {dateRangeValidation.ok &&
        metricsQuery.data &&
        isLiveMetricsEmpty(metricsQuery.data) ? (
          <MetricsEmptyState />
        ) : null}

        {dateRangeValidation.ok &&
        metricsQuery.data &&
        !isLiveMetricsEmpty(metricsQuery.data) ? (
          <MetricCardGrid metrics={metricsQuery.data} />
        ) : null}
      </div>
    </div>
  );
}

function MetricsLoadingState() {
  return (
    <section
      aria-label="Loading metrics"
      className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3"
    >
      {Array.from({ length: 6 }).map((_, index) => (
        <div
          aria-hidden="true"
          className="h-32 animate-pulse rounded-lg border border-zinc-200 bg-white"
          key={index}
        />
      ))}
    </section>
  );
}

function MetricsErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <section className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-900">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-3">
          <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
          <div>
            <h3 className="text-sm font-semibold">Metrics unavailable</h3>
            <p className="mt-1 text-sm text-red-800">
              The analytics service could not return the current summary.
            </p>
          </div>
        </div>
        <button
          className="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-red-300 bg-white px-3 text-sm font-medium text-red-900 transition-colors hover:bg-red-100"
          onClick={onRetry}
          type="button"
        >
          <RefreshCcw className="h-4 w-4" aria-hidden="true" />
          Retry
        </button>
      </div>
    </section>
  );
}

function MetricsEmptyState() {
  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-6 shadow-panel">
      <h3 className="text-sm font-semibold text-zinc-950">No matching sessions</h3>
      <p className="mt-1 text-sm text-zinc-500">
        Try a wider range or a less restrictive segment.
      </p>
    </section>
  );
}

function ValidationState({ message }: { message: string }) {
  return (
    <section className="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
      {message}
    </section>
  );
}
