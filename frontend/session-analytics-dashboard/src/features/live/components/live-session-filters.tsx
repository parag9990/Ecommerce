import { PauseCircle, PlayCircle, RefreshCw, Search } from "lucide-react";

import {
  activeSessionDeviceTypeOptions,
  type ActiveSessionDeviceType
} from "../../../api/session-api";

type LiveSessionFiltersProps = {
  autoRefresh: boolean;
  country: string;
  deviceType: ActiveSessionDeviceType | "all";
  entryPage: string;
  isRefreshing: boolean;
  query: string;
  onAutoRefreshChange: (value: boolean) => void;
  onCountryChange: (value: string) => void;
  onDeviceTypeChange: (value: ActiveSessionDeviceType | "all") => void;
  onEntryPageChange: (value: string) => void;
  onQueryChange: (value: string) => void;
  onRefresh: () => void;
};

export function LiveSessionFilters({
  autoRefresh,
  country,
  deviceType,
  entryPage,
  isRefreshing,
  query,
  onAutoRefreshChange,
  onCountryChange,
  onDeviceTypeChange,
  onEntryPageChange,
  onQueryChange,
  onRefresh
}: LiveSessionFiltersProps) {
  const AutoRefreshIcon = autoRefresh ? PlayCircle : PauseCircle;

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="grid gap-3 lg:grid-cols-[minmax(220px,1fr)_160px_180px_220px_auto_auto] lg:items-end">
        <label className="block" htmlFor="live-session-search">
          <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
            Search
          </span>
          <span className="relative block">
            <Search
              className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-zinc-400"
              aria-hidden="true"
            />
            <input
              className="h-10 w-full rounded-md border border-zinc-300 bg-white pl-9 pr-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
              id="live-session-search"
              maxLength={128}
              onChange={(event) => onQueryChange(event.target.value)}
              placeholder="Session or masked user"
              value={query}
            />
          </span>
        </label>

        <label className="block" htmlFor="live-session-device">
          <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
            Device
          </span>
          <select
            className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
            id="live-session-device"
            onChange={(event) =>
              onDeviceTypeChange(
                event.target.value as ActiveSessionDeviceType | "all"
              )
            }
            value={deviceType}
          >
            {activeSessionDeviceTypeOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>

        <label className="block" htmlFor="live-session-country">
          <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
            Country
          </span>
          <input
            className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
            id="live-session-country"
            maxLength={80}
            onChange={(event) => onCountryChange(event.target.value)}
            placeholder="Any country"
            value={country}
          />
        </label>

        <label className="block" htmlFor="live-session-entry-page">
          <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
            Entry page
          </span>
          <input
            className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
            id="live-session-entry-page"
            maxLength={512}
            onChange={(event) => onEntryPageChange(event.target.value)}
            placeholder="/products"
            value={entryPage}
          />
        </label>

        <label
          className="flex h-10 items-center gap-2 rounded-md border border-zinc-300 bg-zinc-50 px-3 text-sm font-medium text-zinc-700"
          htmlFor="live-session-auto-refresh"
        >
          <input
            checked={autoRefresh}
            className="h-4 w-4 accent-zinc-950"
            id="live-session-auto-refresh"
            onChange={(event) => onAutoRefreshChange(event.target.checked)}
            type="checkbox"
          />
          <AutoRefreshIcon className="h-4 w-4" aria-hidden="true" />
          Auto
        </label>

        <button
          className="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-zinc-950 px-4 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:cursor-not-allowed disabled:bg-zinc-400"
          disabled={isRefreshing}
          onClick={onRefresh}
          title="Refresh active sessions"
          type="button"
        >
          <RefreshCw
            className={["h-4 w-4", isRefreshing ? "animate-spin" : ""].join(" ")}
            aria-hidden="true"
          />
          Refresh
        </button>
      </div>
    </section>
  );
}
