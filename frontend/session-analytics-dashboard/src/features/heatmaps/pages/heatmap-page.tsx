import { AlertTriangle, RefreshCcw, ShieldCheck } from "lucide-react";
import { useMemo, useState } from "react";

import type { HeatmapRequest } from "../../../api/session-api";
import {
  getDefaultDateRange,
  validateDateRange,
  type DateRange
} from "../../../lib/date-range";
import { formatNumber } from "../../../lib/format";
import { HeatmapControls } from "../components/heatmap-controls";
import { HeatmapEmptyState } from "../components/heatmap-empty-state";
import { HeatmapLegend } from "../components/heatmap-legend";
import { HeatmapSummaryStrip } from "../components/heatmap-summary-strip";
import { PagePreviewFrame } from "../components/page-preview-frame";
import { ScrollDepthPanel } from "../components/scroll-depth-panel";
import { useHeatmap } from "../hooks/use-heatmap";
import {
  defaultHeatmapFilters,
  type HeatmapFilters
} from "../lib/heatmap-normalize";

export function HeatmapPage() {
  const [dateRange, setDateRange] = useState<DateRange>(() =>
    getDefaultDateRange()
  );
  const [filters, setFilters] = useState<HeatmapFilters>(
    defaultHeatmapFilters
  );

  const dateRangeValidation = validateDateRange(dateRange);
  const request = useMemo<HeatmapRequest>(
    () => ({
      deviceType: filters.deviceType,
      from: dateRange.from,
      mode: filters.mode,
      path: filters.path,
      to: dateRange.to
    }),
    [dateRange.from, dateRange.to, filters]
  );

  const heatmapQuery = useHeatmap(request, {
    enabled: dateRangeValidation.ok
  });
  const response = heatmapQuery.data;
  const points = response?.points ?? [];
  const hasData = points.length > 0;
  const showPrivacyNotice =
    Boolean(response?.partial) ||
    Boolean(response?.suppressed) ||
    response?.minBucketSize !== undefined;

  return (
    <div>
      <HeatmapControls
        dateRange={dateRange}
        dateRangeError={
          dateRangeValidation.ok ? undefined : dateRangeValidation.message
        }
        filters={filters}
        isRefreshing={heatmapQuery.isFetching}
        onDateRangeChange={setDateRange}
        onFiltersChange={setFilters}
        onRefresh={() => void heatmapQuery.refetch()}
      />

      <div className="space-y-5 p-4 lg:p-6">
        {!dateRangeValidation.ok ? (
          <ValidationState message={dateRangeValidation.message} />
        ) : null}

        {dateRangeValidation.ok && heatmapQuery.isPending ? (
          <HeatmapLoadingState />
        ) : null}

        {dateRangeValidation.ok && heatmapQuery.isError ? (
          <HeatmapErrorState onRetry={() => void heatmapQuery.refetch()} />
        ) : null}

        {dateRangeValidation.ok && response && showPrivacyNotice ? (
          <AggregatePrivacyNotice minBucketSize={response.minBucketSize} />
        ) : null}

        {dateRangeValidation.ok && response && !hasData ? (
          <HeatmapEmptyState />
        ) : null}

        {dateRangeValidation.ok && response && hasData ? (
          <>
            <HeatmapSummaryStrip
              mode={filters.mode}
              points={points}
              response={response}
            />

            <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_340px]">
              <PagePreviewFrame
                deviceType={filters.deviceType}
                mode={filters.mode}
                path={filters.path}
                points={points}
              />

              <aside className="space-y-5">
                <HeatmapLegend compact />
                {filters.mode === "scroll" ? (
                  <ScrollDepthPanel points={points} />
                ) : (
                  <TopBucketsPanel points={points} />
                )}
              </aside>
            </div>
          </>
        ) : null}
      </div>
    </div>
  );
}

function HeatmapLoadingState() {
  return (
    <section
      aria-label="Loading heatmap"
      className="grid gap-4 md:grid-cols-2 xl:grid-cols-4"
    >
      {Array.from({ length: 4 }).map((_, index) => (
        <div
          aria-hidden="true"
          className="h-28 animate-pulse rounded-lg border border-zinc-200 bg-white"
          key={index}
        />
      ))}
    </section>
  );
}

function HeatmapErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <section className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-900">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-3">
          <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
          <div>
            <h3 className="text-sm font-semibold">Heatmap unavailable</h3>
            <p className="mt-1 text-sm text-red-800">
              The analytics service could not return the selected aggregate.
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

function ValidationState({ message }: { message: string }) {
  return (
    <section className="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
      {message}
    </section>
  );
}

function AggregatePrivacyNotice({
  minBucketSize
}: {
  minBucketSize?: number;
}) {
  return (
    <section className="rounded-lg border border-sky-200 bg-sky-50 p-4 text-sky-950">
      <div className="flex gap-3">
        <ShieldCheck className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
        <div>
          <h3 className="text-sm font-semibold">Privacy-safe aggregate</h3>
          <p className="mt-1 text-sm text-sky-800">
            Low-count buckets are hidden or rounded before rendering
            {minBucketSize ? ` below ${formatNumber(minBucketSize)} events` : ""}.
          </p>
        </div>
      </div>
    </section>
  );
}

function TopBucketsPanel({
  points
}: {
  points: Array<{ weight: number; x: number; y: number }>;
}) {
  const topBuckets = [...points]
    .sort((left, right) => right.weight - left.weight)
    .slice(0, 5);

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase text-zinc-500">
            Click buckets
          </p>
          <h2 className="mt-1 text-sm font-semibold text-zinc-950">
            Top aggregate areas
          </h2>
        </div>
      </div>

      <div className="mt-4 divide-y divide-zinc-100">
        {topBuckets.map((point, index) => (
          <div
            className="grid grid-cols-[auto_1fr_auto] items-center gap-3 py-3 text-sm"
            key={`${point.x}-${point.y}-${point.weight}-${index}`}
          >
            <span className="flex h-7 w-7 items-center justify-center rounded-md bg-zinc-100 text-xs font-semibold text-zinc-700">
              {index + 1}
            </span>
            <span className="min-w-0 text-zinc-700">
              {point.x.toFixed(0)}%, {point.y.toFixed(0)}%
            </span>
            <span className="font-semibold text-zinc-950">
              {formatNumber(point.weight)}
            </span>
          </div>
        ))}
      </div>
    </section>
  );
}
