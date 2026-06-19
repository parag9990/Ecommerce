import { CalendarClock, CheckCircle2 } from "lucide-react";
import type { FormEvent } from "react";
import { useMemo, useState } from "react";

import {
  reportDayOfWeekOptions,
  reportFrequencyOptions,
  type AnalyticsReportType,
  type ReportDayOfWeek,
  type ReportFilters,
  type ReportFrequency
} from "../../../api/session-api";
import { useCreateReportSchedule } from "../hooks/use-report-schedules";
import { ReportFormatBadge } from "./report-format-badge";
import { ReportTypeSelect } from "./report-type-select";

type ScheduleReportFormProps = {
  dateRangeValid: boolean;
  filters: ReportFilters;
};

export function ScheduleReportForm({
  dateRangeValid,
  filters
}: ScheduleReportFormProps) {
  const [name, setName] = useState("Weekly analytics report");
  const [reportType, setReportType] =
    useState<AnalyticsReportType>("retention");
  const [frequency, setFrequency] = useState<ReportFrequency>("weekly");
  const [timeOfDay, setTimeOfDay] = useState("09:00");
  const [dayOfWeek, setDayOfWeek] = useState<ReportDayOfWeek>("monday");
  const [dayOfMonth, setDayOfMonth] = useState(1);
  const [recipientsText, setRecipientsText] = useState("");
  const createSchedule = useCreateReportSchedule();
  const recipients = useMemo(
    () => parseRecipients(recipientsText),
    [recipientsText]
  );
  const recipientError =
    recipientsText.trim().length > 0 && !recipients.every(isLikelyEmail);
  const canSubmit =
    dateRangeValid &&
    name.trim().length > 0 &&
    recipients.length > 0 &&
    !recipientError &&
    !createSchedule.isPending;

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!canSubmit) {
      return;
    }

    createSchedule.mutate({
      dayOfMonth: frequency === "monthly" ? dayOfMonth : undefined,
      dayOfWeek: frequency === "weekly" ? dayOfWeek : undefined,
      filters,
      format: "csv",
      frequency,
      name,
      recipients,
      reportType,
      timeOfDay,
      timezone: filters.timezone
    });
  }

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase text-emerald-700">
            Scheduled export
          </p>
          <h2 className="mt-1 text-base font-semibold text-zinc-950">
            Create schedule
          </h2>
        </div>
        <ReportFormatBadge />
      </div>

      <form className="mt-4 grid gap-4" onSubmit={handleSubmit}>
        <div className="grid gap-4 lg:grid-cols-2">
          <label className="block" htmlFor="schedule-name">
            <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
              Name
            </span>
            <input
              className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
              id="schedule-name"
              maxLength={120}
              onChange={(event) => setName(event.target.value)}
              value={name}
            />
          </label>

          <ReportTypeSelect
            id="schedule-report-type"
            value={reportType}
            onChange={setReportType}
          />

          <label className="block" htmlFor="schedule-frequency">
            <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
              Frequency
            </span>
            <select
              className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
              id="schedule-frequency"
              onChange={(event) =>
                setFrequency(event.target.value as ReportFrequency)
              }
              value={frequency}
            >
              {reportFrequencyOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>

          <label className="block" htmlFor="schedule-time">
            <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
              Time
            </span>
            <input
              className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
              id="schedule-time"
              onChange={(event) => setTimeOfDay(event.target.value)}
              type="time"
              value={timeOfDay}
            />
          </label>

          {frequency === "weekly" ? (
            <label className="block" htmlFor="schedule-day-of-week">
              <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
                Day
              </span>
              <select
                className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
                id="schedule-day-of-week"
                onChange={(event) =>
                  setDayOfWeek(event.target.value as ReportDayOfWeek)
                }
                value={dayOfWeek}
              >
                {reportDayOfWeekOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
          ) : null}

          {frequency === "monthly" ? (
            <label className="block" htmlFor="schedule-day-of-month">
              <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
                Day
              </span>
              <input
                className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
                id="schedule-day-of-month"
                max={31}
                min={1}
                onChange={(event) => setDayOfMonth(Number(event.target.value))}
                type="number"
                value={dayOfMonth}
              />
            </label>
          ) : null}

          <label className="block lg:col-span-2" htmlFor="schedule-recipients">
            <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
              Recipients
            </span>
            <input
              className={[
                "h-10 w-full rounded-md border bg-white px-3 text-sm outline-none transition-colors focus:ring-2",
                recipientError
                  ? "border-red-300 focus:border-red-600 focus:ring-red-600/10"
                  : "border-zinc-300 focus:border-zinc-950 focus:ring-zinc-950/10"
              ].join(" ")}
              id="schedule-recipients"
              onChange={(event) => setRecipientsText(event.target.value)}
              placeholder="analytics@example.com, ops@example.com"
              type="text"
              value={recipientsText}
            />
            {recipientError ? (
              <span className="mt-1 block text-xs font-medium text-red-700">
                Enter valid email addresses.
              </span>
            ) : null}
          </label>
        </div>

        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-xs text-zinc-500">
            {filters.from} to {filters.to} in {filters.timezone}
          </p>
          <button
            className="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-zinc-950 px-4 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:cursor-not-allowed disabled:bg-zinc-400"
            disabled={!canSubmit}
            type="submit"
          >
            <CalendarClock className="h-4 w-4" aria-hidden="true" />
            {createSchedule.isPending ? "Saving" : "Create schedule"}
          </button>
        </div>
      </form>

      {createSchedule.isSuccess ? (
        <div className="mt-4 rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-800">
          <div className="flex gap-2">
            <CheckCircle2
              className="mt-0.5 h-4 w-4 shrink-0"
              aria-hidden="true"
            />
            <p>Schedule saved.</p>
          </div>
        </div>
      ) : null}

      {createSchedule.isError ? (
        <div className="mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          Schedule could not be saved. Check recipients and cadence.
        </div>
      ) : null}
    </section>
  );
}

function parseRecipients(value: string): string[] {
  return value
    .split(/[,\n]/)
    .map((recipient) => recipient.trim())
    .filter(Boolean);
}

function isLikelyEmail(value: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
}
