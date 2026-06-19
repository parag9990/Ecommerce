import { Activity, Clock3, Radio, Users } from "lucide-react";

import type { ActiveSessionsResponse } from "../../../api/session-api";
import { formatNumber, formatRelativeTime } from "../../../lib/format";

type ActiveSessionSummaryProps = {
  data?: ActiveSessionsResponse;
  isFetching: boolean;
  isLoading: boolean;
};

export function ActiveSessionSummary({
  data,
  isFetching,
  isLoading
}: ActiveSessionSummaryProps) {
  const items = [
    {
      helper: "Unique active identities",
      icon: Users,
      label: "Active users",
      value:
        data?.activeUsers === undefined ? undefined : formatNumber(data.activeUsers)
    },
    {
      helper: "Open active session records",
      icon: Activity,
      label: "Active sessions",
      value:
        data?.activeSessions === undefined
          ? undefined
          : formatNumber(data.activeSessions)
    },
    {
      helper: "Recent event velocity",
      icon: Radio,
      label: "Events/min",
      value:
        data?.eventsPerMinute === undefined
          ? undefined
          : formatNumber(data.eventsPerMinute)
    },
    {
      helper: isFetching ? "Refreshing in background" : "Latest API snapshot",
      icon: Clock3,
      label: "Last refreshed",
      value: data?.refreshedAt ? formatRelativeTime(data.refreshedAt) : undefined
    }
  ];

  return (
    <section
      aria-label="Active session summary"
      className="grid gap-4 md:grid-cols-2 xl:grid-cols-4"
    >
      {items.map((item) => {
        const Icon = item.icon;

        return (
          <div
            className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel"
            key={item.label}
          >
            <div className="mb-3 flex items-start justify-between gap-3">
              <div>
                <p className="text-sm font-medium text-zinc-500">{item.label}</p>
                <p className="mt-1 text-xs text-zinc-400">{item.helper}</p>
              </div>
              <Icon className="h-5 w-5 text-emerald-600" aria-hidden="true" />
            </div>
            <div className="text-2xl font-semibold tracking-normal text-zinc-950">
              {isLoading ? "..." : item.value ?? "-"}
            </div>
          </div>
        );
      })}
    </section>
  );
}
