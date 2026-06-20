import { AlertTriangle, Monitor, Smartphone, Tablet } from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { DataState } from "../../../components/ui/data-state";
import { RiskBadge } from "../../../components/ui/risk-badge";
import { StatusBadge } from "../../../components/ui/status-badge";
import { TablePagination } from "../../../components/ui/table-pagination";
import { cn } from "../../../lib/classnames";
import { formatDateTime, formatSessionDuration } from "../../../lib/format";
import { MaskedIdentity } from "./masked-identity";
import type { AdminSession, AdminSessionStatus, SessionRisk } from "../types";

function resolveSessionStatus(session: AdminSession): AdminSessionStatus | string {
  return session.status ?? (session.revoked_at ? "revoked" : "active");
}

function deviceIcon(session: AdminSession): LucideIcon {
  const deviceType = session.device?.device_type?.toLowerCase();

  if (deviceType === "mobile") {
    return Smartphone;
  }

  if (deviceType === "tablet") {
    return Tablet;
  }

  return Monitor;
}

function deviceSummary(session: AdminSession): string {
  const browser = session.device?.browser || "Unknown browser";
  const operatingSystem = session.device?.operating_system || session.device?.os || "Unknown OS";

  return `${browser} on ${operatingSystem}`;
}

export function SessionTable({
  sessions,
  getRisk,
  isLoading,
  isFetching,
  error,
  page,
  limit,
  totalCount,
  onPageChange,
  onRetry,
  onOpen
}: {
  sessions: AdminSession[];
  getRisk: (session: AdminSession) => SessionRisk;
  isLoading: boolean;
  isFetching: boolean;
  error: unknown;
  page: number;
  limit: number;
  totalCount?: number;
  onPageChange: (page: number) => void;
  onRetry: () => void;
  onOpen: (sessionId: string) => void;
}) {
  if (isLoading) {
    return <DataState title="Loading sessions" description="Fetching admin-visible session records." />;
  }

  if (error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load sessions"
        description={error instanceof Error ? error.message : "The session list request failed."}
        action={
          <button
            type="button"
            onClick={onRetry}
            className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
          >
            Retry
          </button>
        }
      />
    );
  }

  if (sessions.length === 0) {
    return <DataState title="No sessions found" description="Try a different date, user, status, or risk filter." />;
  }

  return (
    <div className="min-h-0 flex-1 overflow-hidden rounded-lg border border-slate-200 bg-white">
      <div className="min-h-0 overflow-auto">
        <table className="min-w-full border-separate border-spacing-0 text-left text-sm">
          <thead className="sticky top-0 bg-slate-100 text-xs uppercase text-slate-600">
            <tr>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Session</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">User</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Device</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Status</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Risk</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Last Seen</th>
              <th className="border-b border-slate-200 px-4 py-3 text-right font-semibold">Action</th>
            </tr>
          </thead>
          <tbody>
            {sessions.map((session) => {
              const risk = getRisk(session);
              const Icon = risk.level === "high" ? AlertTriangle : deviceIcon(session);

              return (
                <tr
                  key={session.session_id}
                  className={cn(
                    "hover:bg-slate-50",
                    risk.level === "high" && "bg-red-50/40 hover:bg-red-50"
                  )}
                >
                  <td className="border-b border-slate-100 px-4 py-3">
                    <div className="flex min-w-0 items-center gap-3">
                      <span
                        className={cn(
                          "inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600",
                          risk.level === "high" && "bg-red-100 text-red-700"
                        )}
                      >
                        <Icon size={17} aria-hidden="true" />
                      </span>
                      <div className="min-w-0">
                        <MaskedIdentity value={session.session_id} label="Session id" />
                        <div className="mt-1 text-xs text-slate-500">
                          Started {formatDateTime(session.started_at)}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3">
                    <div className="space-y-1">
                      <MaskedIdentity value={session.user_id} label="User id" />
                      <div>
                        <MaskedIdentity
                          value={session.anonymous_id}
                          emptyLabel="No anonymous id"
                          label="Anonymous id"
                        />
                      </div>
                    </div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                    <div className="max-w-[220px] truncate">{deviceSummary(session)}</div>
                    <div className="text-xs capitalize text-slate-500">
                      {session.device?.device_type || "unknown"}
                    </div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3">
                    <StatusBadge status={resolveSessionStatus(session)} />
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3">
                    <div className="flex flex-wrap items-center gap-2">
                      <RiskBadge level={risk.level} />
                      <span className="text-xs font-medium text-slate-600">{risk.score}/100</span>
                    </div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                    <div>{formatDateTime(session.last_seen_at)}</div>
                    <div className="text-xs text-slate-500">
                      Duration {formatSessionDuration(session.started_at, session.last_seen_at)}
                    </div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3 text-right">
                    <button
                      type="button"
                      onClick={() => onOpen(session.session_id)}
                      className="font-medium text-blue-700 hover:text-blue-900"
                    >
                      View journey
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <TablePagination
        page={page}
        limit={limit}
        itemCount={sessions.length}
        totalCount={totalCount}
        isFetching={isFetching}
        onPageChange={onPageChange}
      />
    </div>
  );
}
