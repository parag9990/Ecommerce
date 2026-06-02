import { useEffect, useMemo, useState } from "react";

import { PermissionDenied } from "../../../components/ui/permission-denied";
import { HighRiskSessionTable } from "../components/high-risk-session-table";
import { LiveTrafficPanel } from "../components/live-traffic-panel";
import { SessionDetailDrawer } from "../components/session-detail-drawer";
import { SessionFilterBar, type SessionFilterPatch } from "../components/session-filter-bar";
import { SessionTable } from "../components/session-table";
import {
  SuspiciousActivityFeed,
  type SuspiciousSessionItem
} from "../components/suspicious-activity-feed";
import { useAdminSessions } from "../hooks/use-admin-sessions";
import { useLiveMetrics } from "../hooks/use-live-metrics";
import { useSessionPermissions } from "../permissions";
import { calculateSessionRisk } from "../risk-rules";
import type { AdminSession, AdminSessionFilters } from "../types";

const SESSION_PAGE_SIZE = 25;

function initialFilters(): AdminSessionFilters {
  return {
    user_id: "",
    from: "",
    to: "",
    status: "all",
    risk_level: "all",
    page: 1,
    limit: SESSION_PAGE_SIZE
  };
}

function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedValue(value), delayMs);

    return () => window.clearTimeout(timer);
  }, [delayMs, value]);

  return debouncedValue;
}

function sessionStatus(session: AdminSession): string {
  return session.status ?? (session.revoked_at ? "revoked" : "active");
}

function sortByRisk(items: SuspiciousSessionItem[]): SuspiciousSessionItem[] {
  return [...items].sort((left, right) => right.risk.score - left.risk.score);
}

export function SessionOversightPage() {
  const permissions = useSessionPermissions();
  const [filters, setFilters] = useState<AdminSessionFilters>(() => initialFilters());
  const [selectedSessionId, setSelectedSessionId] = useState<string | null>(null);
  const debouncedUserId = useDebouncedValue(filters.user_id?.trim() ?? "", 350);
  const queryFilters = useMemo(
    () => ({
      ...filters,
      user_id: debouncedUserId,
      status: "all" as const,
      risk_level: "all" as const
    }),
    [debouncedUserId, filters]
  );
  const liveMetricsQuery = useLiveMetrics();
  const sessionsQuery = useAdminSessions(queryFilters);
  const sessionRows = sessionsQuery.data?.sessions ?? [];
  const riskItems = useMemo(
    () =>
      sessionRows.map((session) => ({
        session,
        risk: calculateSessionRisk(session)
      })),
    [sessionRows]
  );
  const visibleRows = useMemo(
    () =>
      riskItems
        .filter(({ session }) => filters.status === "all" || sessionStatus(session) === filters.status)
        .filter(({ risk }) => filters.risk_level === "all" || risk.level === filters.risk_level)
        .map(({ session }) => session),
    [filters.risk_level, filters.status, riskItems]
  );
  const visibleRiskItems = useMemo(
    () =>
      riskItems.filter(({ session, risk }) => {
        const matchesStatus = filters.status === "all" || sessionStatus(session) === filters.status;
        const matchesRisk = filters.risk_level === "all" || risk.level === filters.risk_level;

        return matchesStatus && matchesRisk;
      }),
    [filters.risk_level, filters.status, riskItems]
  );
  const highRiskItems = useMemo(
    () => sortByRisk(visibleRiskItems.filter(({ risk }) => risk.level === "high")).slice(0, 8),
    [visibleRiskItems]
  );
  const totalCount =
    filters.status === "all" && filters.risk_level === "all"
      ? sessionsQuery.data?.total
      : visibleRows.length;

  function updateFilters(patch: SessionFilterPatch) {
    setFilters((current) => ({
      ...current,
      ...patch,
      page: 1
    }));
  }

  if (!permissions.canViewSessions) {
    return <PermissionDenied compact />;
  }

  return (
    <section className="flex min-h-[calc(100vh-6.5rem)] flex-col overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
      <header className="border-b border-slate-200 bg-white px-4 py-4">
        <h1 className="text-xl font-semibold text-slate-950">Session Oversight</h1>
        <p className="mt-1 text-sm text-slate-600">
          Live traffic, high-risk sessions, suspicious activity, and journey inspection.
        </p>
      </header>

      <div className="border-b border-slate-200 bg-slate-50 p-4">
        <LiveTrafficPanel
          metrics={liveMetricsQuery.data}
          isLoading={liveMetricsQuery.isLoading}
          isFetching={liveMetricsQuery.isFetching}
          error={liveMetricsQuery.error}
          onRetry={() => void liveMetricsQuery.refetch()}
        />
      </div>

      <SessionFilterBar
        filters={filters}
        onChange={updateFilters}
        onReset={() => setFilters(initialFilters())}
      />

      <div className="grid min-h-0 flex-1 gap-4 p-4 xl:grid-cols-[minmax(0,1fr)_420px]">
        <main className="min-h-0">
          <SessionTable
            sessions={visibleRows}
            getRisk={(session) => calculateSessionRisk(session)}
            isLoading={sessionsQuery.isLoading}
            isFetching={sessionsQuery.isFetching}
            error={sessionsQuery.error}
            page={filters.page}
            limit={SESSION_PAGE_SIZE}
            totalCount={totalCount}
            onPageChange={(page) => setFilters((current) => ({ ...current, page }))}
            onRetry={() => void sessionsQuery.refetch()}
            onOpen={setSelectedSessionId}
          />
        </main>

        <aside className="min-h-0 space-y-4">
          <SuspiciousActivityFeed items={highRiskItems.slice(0, 5)} onOpen={setSelectedSessionId} />
          <HighRiskSessionTable items={highRiskItems} onOpen={setSelectedSessionId} />
        </aside>
      </div>

      <SessionDetailDrawer sessionId={selectedSessionId} onClose={() => setSelectedSessionId(null)} />
    </section>
  );
}
