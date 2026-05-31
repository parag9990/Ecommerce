import { AlertTriangle } from "lucide-react";
import { useMemo, useState } from "react";

import { FiltersBar } from "../../../layout/filters-bar";
import {
  getDefaultDateRange,
  validateDateRange,
  type DateRange
} from "../../../lib/date-range";
import {
  defaultSegmentFilters,
  type SegmentFilters
} from "../../shell/types";
import { ExportReportPanel } from "../components/export-report-panel";
import { ScheduleReportForm } from "../components/schedule-report-form";
import { ScheduledReportsTable } from "../components/scheduled-reports-table";
import type { ReportFilters } from "../../../api/session-api";

export function ReportsExportPage() {
  const [dateRange, setDateRange] = useState<DateRange>(() =>
    getDefaultDateRange()
  );
  const [segmentFilters, setSegmentFilters] = useState<SegmentFilters>(
    defaultSegmentFilters
  );
  const dateRangeValidation = validateDateRange(dateRange);
  const filters = useMemo<ReportFilters>(
    () => ({
      channel: segmentFilters.channel,
      deviceType: segmentFilters.deviceType,
      from: dateRange.from,
      source: segmentFilters.source,
      timezone: getBrowserTimezone(),
      to: dateRange.to,
      userType: segmentFilters.userType
    }),
    [dateRange.from, dateRange.to, segmentFilters]
  );

  return (
    <div>
      <FiltersBar
        dateRange={dateRange}
        dateRangeError={
          dateRangeValidation.ok ? undefined : dateRangeValidation.message
        }
        description="CSV exports and scheduled analytics reports."
        filters={segmentFilters}
        title="Reports"
        onDateRangeChange={setDateRange}
        onFiltersChange={setSegmentFilters}
      />

      <div className="space-y-5 p-4 lg:p-6">
        {!dateRangeValidation.ok ? (
          <ValidationState message={dateRangeValidation.message} />
        ) : null}

        <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_420px]">
          <ExportReportPanel
            dateRangeValid={dateRangeValidation.ok}
            filters={filters}
          />
          <ScheduleReportForm
            dateRangeValid={dateRangeValidation.ok}
            filters={filters}
          />
        </div>

        <ScheduledReportsTable />
      </div>
    </div>
  );
}

function ValidationState({ message }: { message: string }) {
  return (
    <section className="rounded-lg border border-amber-200 bg-amber-50 p-4 text-amber-900">
      <div className="flex gap-3">
        <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
        <p className="text-sm">{message}</p>
      </div>
    </section>
  );
}

function getBrowserTimezone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
}
