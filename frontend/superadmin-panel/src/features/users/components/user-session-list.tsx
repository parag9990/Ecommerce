import { Monitor, Smartphone, X } from "lucide-react";
import { useState } from "react";

import { DataState } from "../../../components/ui/data-state";
import { formatDateTime, formatSessionDuration } from "../../../lib/format";
import { useSessionJourney } from "../hooks/use-session-journey";
import type { UserSession } from "../types";
import { UserSessionTimeline } from "./user-session-timeline";

function deviceSummary(session: UserSession): string {
  const browser = session.device?.browser || "Unknown browser";
  const os = session.device?.os || "Unknown OS";

  return `${browser} on ${os}`;
}

function isMobileSession(session: UserSession): boolean {
  return session.device?.device_type?.toLowerCase() === "mobile";
}

function SessionJourneyDrawer({
  sessionId,
  onClose
}: {
  sessionId: string;
  onClose: () => void;
}) {
  const journeyQuery = useSessionJourney(sessionId);

  return (
    <div className="fixed inset-0 z-40 flex justify-end bg-slate-950/40">
      <button
        type="button"
        className="absolute inset-0 cursor-default"
        aria-label="Close journey drawer"
        onClick={onClose}
      />
      <aside className="relative flex h-full w-full max-w-xl flex-col bg-white shadow-xl">
        <header className="flex items-start justify-between gap-3 border-b border-slate-200 px-4 py-4">
          <div className="min-w-0">
            <h2 className="text-base font-semibold text-slate-950">Session journey</h2>
            <p className="truncate text-sm text-slate-500">{sessionId}</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            title="Close journey"
            className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-100"
          >
            <X size={17} aria-hidden="true" />
            <span className="sr-only">Close journey</span>
          </button>
        </header>

        <div className="min-h-0 flex-1 overflow-auto p-4">
          {journeyQuery.isLoading ? (
            <DataState title="Loading journey" description="Fetching the ordered session events." />
          ) : journeyQuery.error ? (
            <DataState
              tone="danger"
              title="Unable to load journey"
              description={
                journeyQuery.error instanceof Error
                  ? journeyQuery.error.message
                  : "The journey request failed."
              }
              action={
                <button
                  type="button"
                  onClick={() => void journeyQuery.refetch()}
                  className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
                >
                  Retry
                </button>
              }
            />
          ) : (
            <UserSessionTimeline events={journeyQuery.data?.events ?? []} />
          )}
        </div>
      </aside>
    </div>
  );
}

export function UserSessionList({
  sessions,
  isLoading,
  error,
  onRetry
}: {
  sessions: UserSession[];
  isLoading: boolean;
  error: unknown;
  onRetry: () => void;
}) {
  const [selectedSessionId, setSelectedSessionId] = useState<string | null>(null);

  if (isLoading) {
    return <DataState title="Loading sessions" description="Fetching recent sessions for this user." />;
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
    return <DataState title="No sessions found" description="This user has no recent admin-visible sessions." />;
  }

  return (
    <section className="rounded-lg border border-slate-200 bg-white">
      <header className="border-b border-slate-200 px-4 py-3">
        <h2 className="text-base font-semibold text-slate-950">Sessions</h2>
        <p className="text-sm text-slate-600">Recent sessions with limited, admin-safe device data.</p>
      </header>

      <div className="divide-y divide-slate-100">
        {sessions.map((session) => {
          const Icon = isMobileSession(session) ? Smartphone : Monitor;

          return (
            <button
              type="button"
              key={session.session_id}
              onClick={() => setSelectedSessionId(session.session_id)}
              className="flex w-full items-start gap-3 px-4 py-3 text-left hover:bg-slate-50"
            >
              <Icon className="mt-0.5 h-4 w-4 shrink-0 text-slate-500" aria-hidden="true" />
              <span className="min-w-0 flex-1">
                <span className="block truncate text-sm font-medium text-slate-950">
                  {session.session_id}
                </span>
                <span className="block text-xs text-slate-500">{deviceSummary(session)}</span>
                <span className="block text-xs text-slate-500">
                  Last seen {formatDateTime(session.last_seen_at)} · Duration{" "}
                  {formatSessionDuration(session.started_at, session.last_seen_at)}
                </span>
              </span>
            </button>
          );
        })}
      </div>

      {selectedSessionId ? (
        <SessionJourneyDrawer sessionId={selectedSessionId} onClose={() => setSelectedSessionId(null)} />
      ) : null}
    </section>
  );
}
