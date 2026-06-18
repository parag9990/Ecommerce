import { CalendarDays } from "lucide-react";

import { cn } from "../../../lib/cn";
import { ANALYTICS_PRESETS } from "../types";
import type { AnalyticsDateRange, AnalyticsPreset } from "../types";
import {
  createCustomRange,
  createPresetRange,
  toDateInputValue,
} from "../utils/analytics-date-range";

type DateRangeFilterProps = {
  value: AnalyticsDateRange;
  onChange: (range: AnalyticsDateRange) => void;
};

const presetLabels: Record<AnalyticsPreset, string> = {
  "7d": "7D",
  "30d": "30D",
  "90d": "90D",
};

export function DateRangeFilter({ value, onChange }: DateRangeFilterProps) {
  const fromInput = toDateInputValue(value.from);
  const toInput = toDateInputValue(value.to);

  function updateCustomRange(nextFrom: string, nextTo: string) {
    if (!nextFrom && !nextTo) {
      return;
    }

    onChange(createCustomRange(nextFrom || nextTo, nextTo || nextFrom));
  }

  return (
    <div className="flex flex-wrap items-end gap-2 rounded-md border border-slate-200 bg-white p-2 shadow-sm">
      <div className="flex h-9 rounded-md border border-slate-200 bg-slate-50 p-0.5">
        {ANALYTICS_PRESETS.map((preset) => (
          <button
            key={preset}
            type="button"
            aria-pressed={value.preset === preset}
            onClick={() => onChange(createPresetRange(preset))}
            className={cn(
              "h-8 min-w-12 rounded px-2.5 text-xs font-semibold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600",
              value.preset === preset
                ? "bg-blue-600 text-white shadow-sm"
                : "text-slate-600 hover:bg-white hover:text-slate-950",
            )}
          >
            {presetLabels[preset]}
          </button>
        ))}
      </div>

      <label className="space-y-1">
        <span className="sr-only">Analytics from date</span>
        <div className="relative">
          <CalendarDays
            className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400"
            aria-hidden="true"
          />
          <input
            type="date"
            value={fromInput}
            onChange={(event) => {
              const nextFrom = event.target.value;
              const nextTo = toInput && toInput < nextFrom ? nextFrom : toInput;
              updateCustomRange(nextFrom, nextTo);
            }}
            className="h-9 w-36 rounded-md border border-slate-300 bg-white pl-8 pr-2 text-sm text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          />
        </div>
      </label>

      <label className="space-y-1">
        <span className="sr-only">Analytics to date</span>
        <input
          type="date"
          value={toInput}
          onChange={(event) => {
            const nextTo = event.target.value;
            const nextFrom = fromInput && fromInput > nextTo ? nextTo : fromInput;
            updateCustomRange(nextFrom, nextTo);
          }}
          className="h-9 w-36 rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
        />
      </label>
    </div>
  );
}
