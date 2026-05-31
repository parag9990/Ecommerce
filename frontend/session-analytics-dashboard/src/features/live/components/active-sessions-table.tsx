import { AlertTriangle, RefreshCcw } from "lucide-react";

import type { ActiveSession } from "../../../api/session-api";
import { ActiveSessionRow } from "./active-session-row";

type ActiveSessionsTableProps = {
  error: unknown;
  isLoading: boolean;
  onRetry: () => void;
  sessions: ActiveSession[];
};

export function ActiveSessionsTable({
  error,
  isLoading,
  onRetry,
  sessions
}: ActiveSessionsTableProps) {
  if (isLoading) {
    return <TableLoadingState />;
  }

  if (error) {
    return <TableErrorState onRetry={onRetry} />;
  }

  if (sessions.length === 0) {
    return <TableEmptyState />;
  }

  return (
    <section className="overflow-hidden rounded-lg border border-zinc-200 bg-white shadow-panel">
      <div className="flex flex-col gap-1 border-b border-zinc-200 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
        <h2 className="text-sm font-semibold text-zinc-950">
          Live active sessions
        </h2>
        <span className="text-xs text-zinc-500">{sessions.length} shown</span>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-zinc-200 text-left">
          <thead className="bg-zinc-50 text-xs uppercase text-zinc-500">
            <tr>
              <th className="px-4 py-3 font-semibold" scope="col">
                Session
              </th>
              <th className="px-4 py-3 font-semibold" scope="col">
                Device
              </th>
              <th className="px-4 py-3 font-semibold" scope="col">
                Location
              </th>
              <th className="px-4 py-3 font-semibold" scope="col">
                Entry page
              </th>
              <th className="px-4 py-3 font-semibold" scope="col">
                Current page
              </th>
              <th className="px-4 py-3 font-semibold" scope="col">
                Activity
              </th>
              <th className="px-4 py-3 font-semibold" scope="col">
                Last seen
              </th>
              <th className="px-4 py-3 font-semibold" scope="col">
                Journey
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-200">
            {sessions.map((session) => (
              <ActiveSessionRow key={session.sessionId} session={session} />
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function TableLoadingState() {
  return (
    <section
      aria-label="Loading active sessions"
      className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel"
    >
      <div className="space-y-3">
        {Array.from({ length: 5 }).map((_, index) => (
          <div
            aria-hidden="true"
            className="h-14 animate-pulse rounded-md bg-zinc-100"
            key={index}
          />
        ))}
      </div>
    </section>
  );
}

function TableErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <section className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-900">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-3">
          <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
          <div>
            <h2 className="text-sm font-semibold">Active sessions unavailable</h2>
            <p className="mt-1 text-sm text-red-800">
              The analytics service could not return the live session snapshot.
            </p>
          </div>
        </div>
        <button
          className="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-red-300 bg-white px-3 text-sm font-medium text-red-900 transition-colors hover:bg-red-100"
          onClick={onRetry}
          type="button"
        >
          <RefreshCcw className="h-4 w-4" aria-hidden="true" />
          Retry
        </button>
      </div>
    </section>
  );
}

function TableEmptyState() {
  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-6 shadow-panel">
      <h2 className="text-sm font-semibold text-zinc-950">
        No active sessions match these filters
      </h2>
      <p className="mt-1 text-sm text-zinc-500">
        Live traffic will appear here when the current filters match active users.
      </p>
    </section>
  );
}
