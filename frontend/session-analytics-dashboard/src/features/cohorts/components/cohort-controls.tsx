import { RefreshCcw, Repeat } from "lucide-react";

import {
  retentionIntervalOptions,
  type RetentionInterval
} from "../../../api/session-api";
import { FiltersBar } from "../../../layout/filters-bar";
import type { DateRange } from "../../../lib/date-range";
import type { SegmentFilters } from "../../shell/types";
import {
  formatWindowLabel,
  getRetentionWindowOptions
} from "../lib/retention-format";

type CohortControlsProps = {
  dateRange: DateRange;
  dateRangeError?: string;
  filters: SegmentFilters;
  interval: RetentionInterval;
  isRefreshing: boolean;
  window: number;
  onDateRangeChange: (range: DateRange) => void;
  onFiltersChange: (filters: SegmentFilters) => void;
  onIntervalChange: (interval: RetentionInterval) => void;
  onRefresh: () => void;
  onWindowChange: (window: number) => void;
};

export function CohortControls({
  dateRange,
  dateRangeError,
  filters,
  interval,
  isRefreshing,
  window,
  onDateRangeChange,
  onFiltersChange,
  onIntervalChange,
  onRefresh,
  onWindowChange
}: CohortControlsProps) {
  const windowOptions = getRetentionWindowOptions(interval);

  return (
    <div className="sticky top-0 z-10 border-b border-zinc-200 bg-white/95 backdrop-blur">
      <FiltersBar
        dateRange={dateRange}
        dateRangeError={dateRangeError}
        description="New, returning, and cohort-level aggregate retention."
        filters={filters}
        sticky={false}
        title="Retention Reports"
        onDateRangeChange={onDateRangeChange}
        onFiltersChange={onFiltersChange}
      />

      <div className="grid gap-3 px-4 pb-4 md:grid-cols-[1fr_1fr_auto] lg:px-6 xl:max-w-4xl">
        <IntervalSelect value={interval} onChange={onIntervalChange} />
        <WindowSelect
          interval={interval}
          options={windowOptions}
          value={window}
          onChange={onWindowChange}
        />
        <button
          className="inline-flex h-10 items-center justify-center gap-2 rounded-md border border-zinc-300 bg-white px-3 text-sm font-medium text-zinc-800 transition-colors hover:bg-zinc-100 disabled:cursor-not-allowed disabled:opacity-60"
          disabled={isRefreshing}
          onClick={onRefresh}
          type="button"
        >
          <RefreshCcw
            className={`h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`}
            aria-hidden="true"
          />
          {isRefreshing ? "Refreshing" : "Refresh"}
        </button>
      </div>
    </div>
  );
}

type IntervalSelectProps = {
  value: RetentionInterval;
  onChange: (value: RetentionInterval) => void;
};

function IntervalSelect({ value, onChange }: IntervalSelectProps) {
  return (
    <label className="block" htmlFor="retention-interval">
      <span className="mb-1 flex items-center gap-1.5 text-xs font-semibold uppercase text-zinc-500">
        <Repeat className="h-3.5 w-3.5" aria-hidden="true" />
        Interval
      </span>
      <select
        className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
        id="retention-interval"
        onChange={(event) => onChange(event.target.value as RetentionInterval)}
        value={value}
      >
        {retentionIntervalOptions.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

type WindowSelectProps = {
  interval: RetentionInterval;
  options: readonly number[];
  value: number;
  onChange: (value: number) => void;
};

function WindowSelect({
  interval,
  options,
  value,
  onChange
}: WindowSelectProps) {
  return (
    <label className="block" htmlFor="retention-window">
      <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
        Window
      </span>
      <select
        className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
        id="retention-window"
        onChange={(event) => onChange(Number(event.target.value))}
        value={value}
      >
        {options.map((option) => (
          <option key={option} value={option}>
            {formatWindowLabel(interval, option)}
          </option>
        ))}
      </select>
    </label>
  );
}
