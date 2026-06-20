import { X } from "lucide-react";

import { DataState } from "../../../components/ui/data-state";
import { RiskBadge } from "../../../components/ui/risk-badge";
import { formatDateTime, formatSessionDuration } from "../../../lib/format";
import { useSessionJourney } from "../hooks/use-session-journey";
import { calculateSessionRisk } from "../risk-rules";
import { DeviceSummaryCard } from "./device-summary-card";
import { MaskedIdentity } from "./masked-identity";
import { SessionEventTimeline } from "./session-event-timeline";

export function SessionDetailDrawer({
  sessionId,
  onClose
}: {
  sessionId: string | null;
  onClose: () => void;
}) {
  const journeyQuery = useSessionJourney(sessionId);

  if (!sessionId) {
    return null;
  }

  const session = journeyQuery.data?.session;
  const events = journeyQuery.data?.events ?? [];
  const risk = session ? calculateSessionRisk(session, events) : null;

  return (
    <div className="fixed inset-0 z-40">
      <button
        type="button"
        className="absolute inset-0 bg-slate-950/30"
        aria-label="Close session journey"
        onClick={onClose}
      />
      <aside className="absolute inset-y-0 right-0 flex w-full max-w-2xl flex-col overflow-hidden border-l border-slate-200 bg-slate-50 shadow-xl">
        <header className="border-b border-slate-200 bg-white px-4 py-4">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h2 className="text-lg font-semibold text-slate-950">Session Journey</h2>
              <div className="mt-1">
                <MaskedIdentity value={sessionId} label="Session id" />
              </div>
            </div>
            <button
              type="button"
              onClick={onClose}
              title="Close session journey"
              className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-100"
            >
              <X className="h-4 w-4" aria-hidden="true" />
              <span className="sr-only">Close session journey</span>
            </button>
          </div>
        </header>

        <div className="min-h-0 flex-1 overflow-auto p-4">
          {journeyQuery.isLoading ? (
            <DataState title="Loading journey" description="Fetching the ordered session events." />
          ) : null}

          {journeyQuery.error ? (
            <DataState
              tone="danger"
              title="Unable to load journey"
              description={
                journeyQuery.error instanceof Error
                  ? journeyQuery.error.message
                  : "The session journey request failed."
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
          ) : null}

          {session && risk ? (
            <div className="space-y-4">
              <section className="rounded-lg border border-slate-200 bg-white p-4">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <h3 className="text-base font-semibold text-slate-950">Risk Summary</h3>
                    <p className="mt-1 text-sm text-slate-600">
                      Last seen {formatDateTime(session.last_seen_at)} | Duration{" "}
                      {formatSessionDuration(session.started_at, session.last_seen_at)}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <RiskBadge level={risk.level} />
                    <span className="text-sm font-semibold text-slate-700">{risk.score}/100</span>
                  </div>
                </div>
                <div className="mt-3 flex flex-wrap gap-2">
                  {(risk.reasons.length > 0 ? risk.reasons : ["No elevated UI risk signals"]).map((reason) => (
                    <span
                      key={reason}
                      className="rounded-md border border-slate-200 bg-slate-50 px-2 py-1 text-xs font-medium text-slate-700"
                    >
                      {reason}
                    </span>
                  ))}
                </div>
              </section>

              <DeviceSummaryCard session={session} />

              <section className="rounded-lg border border-slate-200 bg-white p-4">
                <h3 className="text-base font-semibold text-slate-950">Event Timeline</h3>
                <div className="mt-4">
                  <SessionEventTimeline events={events} />
                </div>
              </section>
            </div>
          ) : null}
        </div>
      </aside>
    </div>
  );
}
