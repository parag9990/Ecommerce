import { useMemo, useState } from "react";

import type {
  ActiveSessionDeviceType,
  ActiveSessionsRequest
} from "../../../api/session-api";
import { ActiveSessionSummary } from "../components/active-session-summary";
import { ActiveSessionsTable } from "../components/active-sessions-table";
import { ActiveUserMap } from "../components/active-user-map";
import { DeviceBreakdownPanel } from "../components/device-breakdown-panel";
import { EntryPagesPanel } from "../components/entry-pages-panel";
import { LiveSessionFilters } from "../components/live-session-filters";
import { useActiveSessions } from "../hooks/use-active-sessions";

export function LiveSessionsPage() {
  const [query, setQuery] = useState("");
  const [deviceType, setDeviceType] = useState<ActiveSessionDeviceType | "all">(
    "all"
  );
  const [country, setCountry] = useState("");
  const [entryPage, setEntryPage] = useState("");
  const [autoRefresh, setAutoRefresh] = useState(true);

  const request = useMemo<ActiveSessionsRequest>(
    () => ({
      country: country.trim() || undefined,
      deviceType,
      entryPage: entryPage.trim() || undefined,
      limit: 50,
      q: query.trim() || undefined
    }),
    [country, deviceType, entryPage, query]
  );

  const activeSessionsQuery = useActiveSessions({ autoRefresh, request });
  const data = activeSessionsQuery.data;

  return (
    <div className="p-4 lg:p-6">
      <div className="mb-5 flex flex-col gap-3 xl:flex-row xl:items-end xl:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase text-emerald-700">
            Live traffic
          </p>
          <h1 className="mt-1 text-2xl font-semibold text-zinc-950">
            Active sessions
          </h1>
        </div>
        <p className="text-sm text-zinc-500">
          {autoRefresh ? "Auto refresh on" : "Auto refresh paused"}
        </p>
      </div>

      <div className="space-y-5">
        <LiveSessionFilters
          autoRefresh={autoRefresh}
          country={country}
          deviceType={deviceType}
          entryPage={entryPage}
          isRefreshing={activeSessionsQuery.isFetching}
          query={query}
          onAutoRefreshChange={setAutoRefresh}
          onCountryChange={setCountry}
          onDeviceTypeChange={setDeviceType}
          onEntryPageChange={setEntryPage}
          onQueryChange={setQuery}
          onRefresh={() => void activeSessionsQuery.refetch()}
        />

        <ActiveSessionSummary
          data={data}
          isFetching={activeSessionsQuery.isFetching}
          isLoading={activeSessionsQuery.isPending}
        />

        <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_360px]">
          <ActiveSessionsTable
            error={activeSessionsQuery.error}
            isLoading={activeSessionsQuery.isPending}
            onRetry={() => void activeSessionsQuery.refetch()}
            sessions={data?.sessions ?? []}
          />

          <aside className="space-y-5">
            <ActiveUserMap locations={data?.locationBreakdown ?? []} />
            <DeviceBreakdownPanel devices={data?.deviceBreakdown ?? []} />
            <EntryPagesPanel pages={data?.entryPageBreakdown ?? []} />
          </aside>
        </div>
      </div>
    </div>
  );
}
