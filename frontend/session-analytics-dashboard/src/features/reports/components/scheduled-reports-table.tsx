import { AlertTriangle, Pause, Play, RefreshCcw, Trash2 } from "lucide-react";

import type { ReportSchedule } from "../../../api/session-api";
import {
  useDeleteReportSchedule,
  useReportSchedules,
  useUpdateReportScheduleStatus
} from "../hooks/use-report-schedules";
import { reportLabels } from "../lib/report-labels";
import {
  formatReportFrequency,
  formatReportRunTime
} from "../lib/report-schedule-format";
import { ReportEmptyState } from "./report-empty-state";
import { ScheduleStatusBadge } from "./schedule-status-badge";

export function ScheduledReportsTable() {
  const schedulesQuery = useReportSchedules();
  const updateStatus = useUpdateReportScheduleStatus();
  const deleteSchedule = useDeleteReportSchedule();
  const isMutating = updateStatus.isPending || deleteSchedule.isPending;

  if (schedulesQuery.isPending) {
    return <ScheduledReportsLoadingState />;
  }

  if (schedulesQuery.isError) {
    return (
      <section className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-900">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex gap-3">
            <AlertTriangle
              className="mt-0.5 h-5 w-5 shrink-0"
              aria-hidden="true"
            />
            <div>
              <h3 className="text-sm font-semibold">Schedules unavailable</h3>
              <p className="mt-1 text-sm text-red-800">
                The analytics service could not load saved report schedules.
              </p>
            </div>
          </div>
          <button
            className="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-red-300 bg-white px-3 text-sm font-medium text-red-900 transition-colors hover:bg-red-100"
            onClick={() => void schedulesQuery.refetch()}
            type="button"
          >
            <RefreshCcw className="h-4 w-4" aria-hidden="true" />
            Retry
          </button>
        </div>
      </section>
    );
  }

  const schedules = schedulesQuery.data.items;

  if (!schedules.length) {
    return (
      <ReportEmptyState
        title="No scheduled reports"
        description="Create a daily, weekly, or monthly CSV schedule from the form above."
      />
    );
  }

  function handleToggle(schedule: ReportSchedule) {
    updateStatus.mutate({
      id: schedule.id,
      status: schedule.status === "active" ? "paused" : "active"
    });
  }

  function handleDelete(schedule: ReportSchedule) {
    if (
      window.confirm(`Delete scheduled report "${schedule.name}"? This cannot be undone.`)
    ) {
      deleteSchedule.mutate(schedule.id);
    }
  }

  return (
    <section className="overflow-hidden rounded-lg border border-zinc-200 bg-white shadow-panel">
      <div className="flex flex-col gap-3 border-b border-zinc-200 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase text-zinc-500">
            Automation
          </p>
          <h2 className="mt-1 text-base font-semibold text-zinc-950">
            Scheduled reports
          </h2>
        </div>
        <button
          className="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-zinc-300 bg-white px-3 text-sm font-medium text-zinc-800 transition-colors hover:bg-zinc-100 disabled:cursor-not-allowed disabled:text-zinc-400"
          disabled={schedulesQuery.isFetching}
          onClick={() => void schedulesQuery.refetch()}
          type="button"
        >
          <RefreshCcw
            className={`h-4 w-4 ${schedulesQuery.isFetching ? "animate-spin" : ""}`}
            aria-hidden="true"
          />
          {schedulesQuery.isFetching ? "Refreshing" : "Refresh"}
        </button>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-zinc-200 text-sm">
          <thead className="bg-zinc-50 text-left text-xs font-semibold uppercase text-zinc-500">
            <tr>
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Report</th>
              <th className="px-4 py-3">Frequency</th>
              <th className="px-4 py-3">Recipients</th>
              <th className="px-4 py-3">Next run</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-100">
            {schedules.map((schedule) => (
              <tr key={schedule.id}>
                <td className="px-4 py-3">
                  <div className="max-w-56">
                    <p className="truncate font-medium text-zinc-950">
                      {schedule.name}
                    </p>
                    {schedule.lastRunAt ? (
                      <p className="mt-1 text-xs text-zinc-500">
                        Last {formatReportRunTime(schedule.lastRunAt)}
                      </p>
                    ) : null}
                  </div>
                </td>
                <td className="px-4 py-3 text-zinc-700">
                  {reportLabels[schedule.reportType]}
                </td>
                <td className="px-4 py-3 text-zinc-700">
                  {formatReportFrequency(schedule.frequency)}
                </td>
                <td className="px-4 py-3 text-zinc-700">
                  {schedule.recipients.length}
                </td>
                <td className="px-4 py-3 text-zinc-700">
                  {formatReportRunTime(schedule.nextRunAt)}
                </td>
                <td className="px-4 py-3">
                  <ScheduleStatusBadge status={schedule.status} />
                </td>
                <td className="px-4 py-3">
                  <div className="flex justify-end gap-2">
                    <button
                      aria-label={
                        schedule.status === "active"
                          ? "Pause schedule"
                          : "Resume schedule"
                      }
                      className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-zinc-300 text-zinc-700 transition-colors hover:bg-zinc-100 disabled:cursor-not-allowed disabled:text-zinc-400"
                      disabled={isMutating}
                      onClick={() => handleToggle(schedule)}
                      type="button"
                    >
                      {schedule.status === "active" ? (
                        <Pause className="h-4 w-4" aria-hidden="true" />
                      ) : (
                        <Play className="h-4 w-4" aria-hidden="true" />
                      )}
                    </button>
                    <button
                      aria-label="Delete schedule"
                      className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-red-200 text-red-700 transition-colors hover:bg-red-50 disabled:cursor-not-allowed disabled:text-red-300"
                      disabled={isMutating}
                      onClick={() => handleDelete(schedule)}
                      type="button"
                    >
                      <Trash2 className="h-4 w-4" aria-hidden="true" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {updateStatus.isError || deleteSchedule.isError ? (
        <div className="border-t border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          Schedule action failed. Refresh and retry.
        </div>
      ) : null}
    </section>
  );
}

function ScheduledReportsLoadingState() {
  return (
    <section
      aria-label="Loading scheduled reports"
      className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel"
    >
      <div className="h-5 w-44 animate-pulse rounded bg-zinc-200" />
      <div className="mt-4 space-y-3">
        {Array.from({ length: 3 }).map((_, index) => (
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
