import { AlertTriangle } from "lucide-react";

import { DataState } from "../../../components/ui/data-state";
import { MaskedIdentity } from "./masked-identity";
import type { AdminSession, SessionRisk } from "../types";

export type SuspiciousSessionItem = {
  session: AdminSession;
  risk: SessionRisk;
};

function reasonSummary(risk: SessionRisk): string {
  return risk.reasons.length > 0 ? risk.reasons.join(", ") : "Backend risk signal";
}

export function SuspiciousActivityFeed({
  items,
  onOpen
}: {
  items: SuspiciousSessionItem[];
  onOpen: (sessionId: string) => void;
}) {
  if (items.length === 0) {
    return (
      <DataState
        title="No suspicious activity"
        description="No high-risk sessions were found in the current result set."
      />
    );
  }

  return (
    <section className="rounded-lg border border-slate-200 bg-white">
      <header className="border-b border-slate-200 px-4 py-3">
        <h2 className="text-base font-semibold text-slate-950">Suspicious Activity</h2>
      </header>
      <div className="divide-y divide-slate-100">
        {items.map(({ session, risk }) => (
          <button
            type="button"
            key={session.session_id}
            onClick={() => onOpen(session.session_id)}
            className="flex w-full items-start gap-3 px-4 py-3 text-left hover:bg-red-50"
          >
            <span className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-red-50 text-red-700">
              <AlertTriangle className="h-4 w-4" aria-hidden="true" />
            </span>
            <span className="min-w-0 flex-1">
              <span className="flex flex-wrap items-center justify-between gap-2">
                <MaskedIdentity value={session.session_id} label="Session id" />
                <span className="rounded-md border border-red-200 bg-red-50 px-2 py-0.5 text-xs font-medium text-red-700">
                  {risk.score}/100
                </span>
              </span>
              <span className="mt-2 block text-sm text-slate-700">{reasonSummary(risk)}</span>
            </span>
          </button>
        ))}
      </div>
    </section>
  );
}
