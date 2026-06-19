import { AlertTriangle, RefreshCcw, ShieldCheck } from "lucide-react";
import { useMemo, useState } from "react";

import type {
  RetentionInterval,
  RetentionReportRequest
} from "../../../api/session-api";
import {
  getDefaultDateRange,
  validateDateRange,
  type DateRange
} from "../../../lib/date-range";
import { formatNumber } from "../../../lib/format";
import {
  defaultSegmentFilters,
  type SegmentFilters
} from "../../shell/types";
import { CohortControls } from "../components/cohort-controls";
import { CohortEmptyState } from "../components/cohort-empty-state";
import { CohortRetentionMatrix } from "../components/cohort-retention-matrix";
import { CohortSummaryCards } from "../components/cohort-summary-cards";
import { NewReturningChart } from "../components/new-returning-chart";
import { RetentionLegend } from "../components/retention-legend";
import { useRetentionReport } from "../hooks/use-retention-report";
import { getRetentionWindowOptions } from "../lib/retention-format";
import {
  hasRetentionData,
  hasSuppressedBuckets
} from "../lib/retention-math";

export function CohortRetentionPage() {
  const [dateRange, setDateRange] = useState<DateRange>(() =>
    getDefaultDateRange()
  );
  const [filters, setFilters] = useState<SegmentFilters>(defaultSegmentFilters);
  const [interval, setInterval] = useState<RetentionInterval>("week");
  const [window, setWindow] = useState(8);

  const dateRangeValidation = validateDateRange(dateRange);
  const request = useMemo<RetentionReportRequest>(
    () => ({
      dateRange,
      filters,
      interval,
      window
    }),
    [dateRange, filters, interval, window]
  );

  const retentionQuery = useRetentionReport(request, {
    enabled: dateRangeValidation.ok
  });
  const report = retentionQuery.data;
  const hasData = report ? hasRetentionData(report.cohorts) : false;
  const showPrivacyNotice =
    Boolean(report?.meta.partial) ||
    Boolean(report?.meta.suppressed) ||
    report?.meta.smallCountThreshold !== undefined ||
    Boolean(report && hasSuppressedBuckets(report.cohorts));

  function handleIntervalChange(nextInterval: RetentionInterval) {
    const options = getRetentionWindowOptions(nextInterval);
    setInterval(nextInterval);
    setWindow((current) =>
      options.includes(current) ? current : options[0]
    );
  }

  return (
    <div>
      <CohortControls
        dateRange={dateRange}
        dateRangeError={
          dateRangeValidation.ok ? undefined : dateRangeValidation.message
        }
        filters={filters}
        interval={interval}
        isRefreshing={retentionQuery.isFetching}
        window={window}
        onDateRangeChange={setDateRange}
        onFiltersChange={setFilters}
        onIntervalChange={handleIntervalChange}
        onRefresh={() => void retentionQuery.refetch()}
        onWindowChange={setWindow}
      />

      <div className="space-y-5 p-4 lg:p-6">
        {!dateRangeValidation.ok ? (
          <ValidationState message={dateRangeValidation.message} />
        ) : null}

        {dateRangeValidation.ok && retentionQuery.isPending ? (
          <RetentionLoadingState />
        ) : null}

        {dateRangeValidation.ok && retentionQuery.isError ? (
          <RetentionErrorState onRetry={() => void retentionQuery.refetch()} />
        ) : null}

        {dateRangeValidation.ok && report && showPrivacyNotice ? (
          <RetentionPrivacyNotice
            smallCountThreshold={report.meta.smallCountThreshold}
          />
        ) : null}

        {dateRangeValidation.ok && report && !hasData ? (
          <CohortEmptyState />
        ) : null}

        {dateRangeValidation.ok && report && hasData ? (
          <>
            <CohortSummaryCards summary={report.summary} />
            <NewReturningChart data={report.newVsReturning} />
            <div className="flex justify-end">
              <RetentionLegend />
            </div>
            <CohortRetentionMatrix
              cohorts={report.cohorts}
              interval={interval}
              window={window}
            />
          </>
        ) : null}
      </div>
    </div>
  );
}

function RetentionLoadingState() {
  return (
    <section
      aria-label="Loading retention report"
      className="grid gap-4 md:grid-cols-2 xl:grid-cols-4"
    >
      {Array.from({ length: 4 }).map((_, index) => (
        <div
          aria-hidden="true"
          className="h-32 animate-pulse rounded-lg border border-zinc-200 bg-white"
          key={index}
        />
      ))}
    </section>
  );
}

function RetentionErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <section className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-900">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-3">
          <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
          <div>
            <h3 className="text-sm font-semibold">Retention report unavailable</h3>
            <p className="mt-1 text-sm text-red-800">
              The analytics service could not return the retention aggregate.
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

function RetentionPrivacyNotice({
  smallCountThreshold
}: {
  smallCountThreshold?: number;
}) {
  return (
    <section className="rounded-lg border border-sky-200 bg-sky-50 p-4 text-sky-950">
      <div className="flex gap-3">
        <ShieldCheck className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
        <div>
          <h3 className="text-sm font-semibold">Privacy-safe aggregate</h3>
          <p className="mt-1 text-sm text-sky-800">
            Low-count cohorts are masked before rendering
            {smallCountThreshold
              ? ` below ${formatNumber(smallCountThreshold)} identities`
              : ""}.
          </p>
        </div>
      </div>
    </section>
  );
}
