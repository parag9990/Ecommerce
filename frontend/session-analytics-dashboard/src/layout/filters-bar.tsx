import { CalendarDays } from "lucide-react";

import {
  DATE_PRESETS,
  getDateRangeForPreset,
  type DateRange
} from "../lib/date-range";
import {
  SegmentFilterPanel
} from "../features/shell/components/segment-filter-panel";
import type { SegmentFilters } from "../features/shell/types";

type FiltersBarProps = {
  dateRange: DateRange;
  dateRangeError?: string;
  description?: string;
  filters: SegmentFilters;
  onDateRangeChange: (range: DateRange) => void;
  onFiltersChange: (filters: SegmentFilters) => void;
  sticky?: boolean;
  title?: string;
};

export function FiltersBar({
  dateRange,
  dateRangeError,
  description = "Session metrics by date range and segment.",
  filters,
  onDateRangeChange,
  onFiltersChange,
  sticky = true,
  title = "Overview"
}: FiltersBarProps) {
  return (
    <section
      className={[
        "border-b border-zinc-200 bg-white/95 px-4 py-4 backdrop-blur lg:px-6",
        sticky ? "sticky top-0 z-10" : ""
      ].join(" ")}
    >
      <div className="mb-4 flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
        <div>
          <h2 className="text-xl font-semibold">{title}</h2>
          <p className="mt-1 text-sm text-zinc-500">{description}</p>
        </div>

        <div className="grid gap-3 md:grid-cols-[auto_1fr] md:items-end">
          <div>
            <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
              Range
            </span>
            <div className="flex rounded-md border border-zinc-300 bg-zinc-100 p-1">
              {DATE_PRESETS.map((preset) => (
                <button
                  aria-pressed={dateRange.preset === preset.value}
                  className={[
                    "h-9 min-w-14 rounded px-3 text-sm font-medium transition-colors",
                    dateRange.preset === preset.value
                      ? "bg-white text-zinc-950 shadow-panel"
                      : "text-zinc-600 hover:text-zinc-950"
                  ].join(" ")}
                  key={preset.value}
                  onClick={() =>
                    onDateRangeChange(getDateRangeForPreset(preset.value))
                  }
                  type="button"
                >
                  {preset.label}
                </button>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-2">
            <DateInput
              label="From"
              value={dateRange.from}
              onChange={(from) =>
                onDateRangeChange({ ...dateRange, from, preset: "custom" })
              }
            />
            <DateInput
              label="To"
              value={dateRange.to}
              onChange={(to) =>
                onDateRangeChange({ ...dateRange, preset: "custom", to })
              }
            />
          </div>
        </div>
      </div>

      {dateRangeError ? (
        <p className="mb-3 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {dateRangeError}
        </p>
      ) : null}

      <SegmentFilterPanel value={filters} onChange={onFiltersChange} />
    </section>
  );
}

type DateInputProps = {
  label: string;
  value: string;
  onChange: (value: string) => void;
};

function DateInput({ label, value, onChange }: DateInputProps) {
  return (
    <label className="block">
      <span className="mb-1 flex items-center gap-1.5 text-xs font-semibold uppercase text-zinc-500">
        <CalendarDays className="h-3.5 w-3.5" aria-hidden="true" />
        {label}
      </span>
      <input
        className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
        onChange={(event) => onChange(event.target.value)}
        type="date"
        value={value}
      />
    </label>
  );
}
