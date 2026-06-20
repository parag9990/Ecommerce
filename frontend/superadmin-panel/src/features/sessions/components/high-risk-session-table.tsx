import { DataState } from "../../../components/ui/data-state";
import { RiskBadge } from "../../../components/ui/risk-badge";
import { formatDateTime } from "../../../lib/format";
import { MaskedIdentity } from "./masked-identity";
import type { SuspiciousSessionItem } from "./suspicious-activity-feed";

export function HighRiskSessionTable({
  items,
  onOpen
}: {
  items: SuspiciousSessionItem[];
  onOpen: (sessionId: string) => void;
}) {
  if (items.length === 0) {
    return <DataState title="No high-risk sessions" description="Current filters did not return high-risk rows." />;
  }

  return (
    <section className="overflow-hidden rounded-lg border border-slate-200 bg-white">
      <header className="border-b border-slate-200 px-4 py-3">
        <h2 className="text-base font-semibold text-slate-950">High-Risk Sessions</h2>
      </header>
      <div className="overflow-auto">
        <table className="min-w-full border-separate border-spacing-0 text-left text-sm">
          <thead className="bg-slate-100 text-xs uppercase text-slate-600">
            <tr>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Session</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Risk</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Last Seen</th>
              <th className="border-b border-slate-200 px-4 py-3 text-right font-semibold">Action</th>
            </tr>
          </thead>
          <tbody>
            {items.map(({ session, risk }) => (
              <tr key={session.session_id} className="hover:bg-red-50">
                <td className="border-b border-slate-100 px-4 py-3">
                  <MaskedIdentity value={session.session_id} label="Session id" />
                </td>
                <td className="border-b border-slate-100 px-4 py-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <RiskBadge level={risk.level} />
                    <span className="text-xs font-medium text-slate-600">{risk.score}/100</span>
                  </div>
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                  {formatDateTime(session.last_seen_at)}
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-right">
                  <button
                    type="button"
                    onClick={() => onOpen(session.session_id)}
                    className="font-medium text-blue-700 hover:text-blue-900"
                  >
                    Inspect
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
