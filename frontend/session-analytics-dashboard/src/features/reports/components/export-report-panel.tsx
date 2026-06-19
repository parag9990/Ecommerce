import { AlertTriangle, Download, ShieldCheck } from "lucide-react";
import { useState } from "react";

import type {
  AnalyticsReportType,
  ReportFilters
} from "../../../api/session-api";
import { useExportReport } from "../hooks/use-export-report";
import { ReportFormatBadge } from "./report-format-badge";
import { ReportTypeSelect } from "./report-type-select";

type ExportReportPanelProps = {
  dateRangeValid: boolean;
  filters: ReportFilters;
};

export function ExportReportPanel({
  dateRangeValid,
  filters
}: ExportReportPanelProps) {
  const [reportType, setReportType] =
    useState<AnalyticsReportType>("funnel");
  const exportReport = useExportReport();

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase text-emerald-700">
            Immediate export
          </p>
          <h2 className="mt-1 text-base font-semibold text-zinc-950">
            Download report
          </h2>
        </div>
        <ReportFormatBadge />
      </div>

      <div className="mt-4 grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
        <ReportTypeSelect
          value={reportType}
          onChange={setReportType}
        />
        <button
          className="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-zinc-950 px-4 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:cursor-not-allowed disabled:bg-zinc-400"
          disabled={!dateRangeValid || exportReport.isPending}
          onClick={() =>
            exportReport.mutate({
              ...filters,
              format: "csv",
              reportType
            })
          }
          type="button"
        >
          <Download className="h-4 w-4" aria-hidden="true" />
          {exportReport.isPending ? "Preparing CSV" : "Download CSV"}
        </button>
      </div>

      <div className="mt-4 rounded-md border border-sky-200 bg-sky-50 px-3 py-2 text-sm text-sky-900">
        <div className="flex gap-2">
          <ShieldCheck className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
          <p>
            Exports use aggregate and summary data only; raw event payloads and
            direct personal identifiers stay out of the file.
          </p>
        </div>
      </div>

      {exportReport.isError ? (
        <div className="mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          <div className="flex gap-2">
            <AlertTriangle
              className="mt-0.5 h-4 w-4 shrink-0"
              aria-hidden="true"
            />
            <p>Report export failed. Adjust the range or filters and retry.</p>
          </div>
        </div>
      ) : null}
    </section>
  );
}
