import { AlertTriangle, RefreshCcw } from "lucide-react";
import { useMemo, useState } from "react";

import type { FunnelReportRequest } from "../../../api/session-api";
import {
  getDefaultDateRange,
  validateDateRange,
  type DateRange
} from "../../../lib/date-range";
import {
  defaultSegmentFilters,
  type SegmentFilters
} from "../../shell/types";
import { ConversionSummaryStrip } from "../components/conversion-summary-strip";
import { FunnelChart } from "../components/funnel-chart";
import { FunnelDropoffTable } from "../components/funnel-dropoff-table";
import { FunnelEmptyState } from "../components/funnel-empty-state";
import { FunnelFilters } from "../components/funnel-filters";
import { FunnelInsightPanel } from "../components/funnel-insight-panel";
import { FunnelStepList } from "../components/funnel-step-list";
import { useFunnelReport } from "../hooks/use-funnel-report";
import { defaultFunnelSteps } from "../lib/funnel-format";

export function FunnelAnalysisPage() {
  const [dateRange, setDateRange] = useState<DateRange>(() =>
    getDefaultDateRange()
  );
  const [filters, setFilters] = useState<SegmentFilters>(defaultSegmentFilters);

  const dateRangeValidation = validateDateRange(dateRange);
  const request = useMemo<FunnelReportRequest>(
    () => ({
      dateRange,
      filters,
      steps: defaultFunnelSteps
    }),
    [dateRange, filters]
  );

  const funnelQuery = useFunnelReport(request, {
    enabled: dateRangeValidation.ok
  });
  const report = funnelQuery.data;
  const hasData = report ? report.totalStarted > 0 : false;

  return (
    <div>
      <FunnelFilters
        dateRange={dateRange}
        dateRangeError={
          dateRangeValidation.ok ? undefined : dateRangeValidation.message
        }
        filters={filters}
        steps={defaultFunnelSteps}
        onDateRangeChange={setDateRange}
        onFiltersChange={setFilters}
      />

      <div className="space-y-5 p-4 lg:p-6">
        {!dateRangeValidation.ok ? (
          <ValidationState message={dateRangeValidation.message} />
        ) : null}

        {dateRangeValidation.ok && funnelQuery.isPending ? (
          <FunnelLoadingState />
        ) : null}

        {dateRangeValidation.ok && funnelQuery.isError ? (
          <FunnelErrorState onRetry={() => void funnelQuery.refetch()} />
        ) : null}

        {dateRangeValidation.ok && report && !hasData ? (
          <FunnelEmptyState />
        ) : null}

        {dateRangeValidation.ok && report && hasData ? (
          <>
            <div className="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
              <div>
                <p className="text-xs font-semibold uppercase text-emerald-700">
                  Conversion diagnostics
                </p>
                <h1 className="mt-1 text-2xl font-semibold text-zinc-950">
                  Product-to-paid funnel
                </h1>
              </div>
              <button
                className="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-zinc-300 bg-white px-3 text-sm font-medium text-zinc-800 transition-colors hover:bg-zinc-100"
                disabled={funnelQuery.isFetching}
                onClick={() => void funnelQuery.refetch()}
                type="button"
              >
                <RefreshCcw
                  className={`h-4 w-4 ${
                    funnelQuery.isFetching ? "animate-spin" : ""
                  }`}
                  aria-hidden="true"
                />
                {funnelQuery.isFetching ? "Refreshing" : "Refresh"}
              </button>
            </div>

            <ConversionSummaryStrip report={report} />

            <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_340px]">
              <FunnelChart report={report} />
              <FunnelInsightPanel report={report} />
            </div>

            <FunnelStepList report={report} />
            <FunnelDropoffTable report={report} />
          </>
        ) : null}
      </div>
    </div>
  );
}

function FunnelLoadingState() {
  return (
    <section
      aria-label="Loading funnel report"
      className="grid gap-4 md:grid-cols-2 xl:grid-cols-4"
    >
      {Array.from({ length: 4 }).map((_, index) => (
        <div
          aria-hidden="true"
          className="h-36 animate-pulse rounded-lg border border-zinc-200 bg-white"
          key={index}
        />
      ))}
    </section>
  );
}

function FunnelErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <section className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-900">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-3">
          <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
          <div>
            <h3 className="text-sm font-semibold">Funnel report unavailable</h3>
            <p className="mt-1 text-sm text-red-800">
              The analytics service could not return the funnel aggregate.
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
