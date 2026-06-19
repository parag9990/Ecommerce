import {
  CalendarDays,
  MousePointerClick,
  RefreshCcw,
  ScrollText
} from "lucide-react";

import {
  heatmapDeviceTypeOptions,
  heatmapModeOptions,
  type HeatmapDeviceType,
  type HeatmapMode
} from "../../../api/session-api";
import {
  DATE_PRESETS,
  getDateRangeForPreset,
  type DateRange
} from "../../../lib/date-range";
import {
  heatmapPagePathOptions,
  type HeatmapFilters
} from "../lib/heatmap-normalize";

type HeatmapControlsProps = {
  dateRange: DateRange;
  dateRangeError?: string;
  filters: HeatmapFilters;
  isRefreshing: boolean;
  onDateRangeChange: (range: DateRange) => void;
  onFiltersChange: (filters: HeatmapFilters) => void;
  onRefresh: () => void;
};

export function HeatmapControls({
  dateRange,
  dateRangeError,
  filters,
  isRefreshing,
  onDateRangeChange,
  onFiltersChange,
  onRefresh
}: HeatmapControlsProps) {
  return (
    <div className="sticky top-0 z-10 border-b border-zinc-200 bg-white/95 px-4 py-4 backdrop-blur lg:px-6">
      <div className="mb-4 flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase text-emerald-700">
            UX diagnostics
          </p>
          <h1 className="mt-1 text-2xl font-semibold text-zinc-950">
            Heatmaps
          </h1>
        </div>

        <div className="grid gap-3 md:grid-cols-[auto_1fr_auto] md:items-end">
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

      {dateRangeError ? (
        <p className="mb-3 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
          {dateRangeError}
        </p>
      ) : null}

      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-[1.4fr_1fr_1fr]">
        <PathSelect
          value={filters.path}
          onChange={(path) => onFiltersChange({ ...filters, path })}
        />
        <DeviceSelect
          value={filters.deviceType}
          onChange={(deviceType) =>
            onFiltersChange({ ...filters, deviceType })
          }
        />
        <ModeSegmentedControl
          value={filters.mode}
          onChange={(mode) => onFiltersChange({ ...filters, mode })}
        />
      </div>
    </div>
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

type PathSelectProps = {
  value: string;
  onChange: (value: string) => void;
};

function PathSelect({ value, onChange }: PathSelectProps) {
  return (
    <label className="block" htmlFor="heatmap-page-path">
      <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
        Page
      </span>
      <select
        className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
        id="heatmap-page-path"
        onChange={(event) => onChange(event.target.value)}
        value={value}
      >
        {heatmapPagePathOptions.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

type DeviceSelectProps = {
  value: HeatmapDeviceType;
  onChange: (value: HeatmapDeviceType) => void;
};

function DeviceSelect({ value, onChange }: DeviceSelectProps) {
  return (
    <label className="block" htmlFor="heatmap-device-type">
      <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
        Device
      </span>
      <select
        className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
        id="heatmap-device-type"
        onChange={(event) => onChange(event.target.value as HeatmapDeviceType)}
        value={value}
      >
        {heatmapDeviceTypeOptions.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

type ModeSegmentedControlProps = {
  value: HeatmapMode;
  onChange: (value: HeatmapMode) => void;
};

function ModeSegmentedControl({ value, onChange }: ModeSegmentedControlProps) {
  return (
    <div>
      <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
        Mode
      </span>
      <div className="grid h-10 grid-cols-2 rounded-md border border-zinc-300 bg-zinc-100 p-1">
        {heatmapModeOptions.map((option) => {
          const Icon = option.value === "click" ? MousePointerClick : ScrollText;

          return (
            <button
              aria-pressed={value === option.value}
              className={[
                "inline-flex min-w-0 items-center justify-center gap-2 rounded px-2 text-sm font-medium transition-colors",
                value === option.value
                  ? "bg-white text-zinc-950 shadow-panel"
                  : "text-zinc-600 hover:text-zinc-950"
              ].join(" ")}
              key={option.value}
              onClick={() => onChange(option.value)}
              type="button"
            >
              <Icon className="h-4 w-4 shrink-0" aria-hidden="true" />
              <span>{option.label}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
