import { formatDateTime } from "../../../lib/format";
import type { SessionEvent } from "../types";

export function UserSessionTimeline({ events }: { events: SessionEvent[] }) {
  if (events.length === 0) {
    return <div className="text-sm text-slate-600">No journey events found.</div>;
  }

  return (
    <ol className="space-y-3">
      {events.map((event, index) => (
        <li key={`${event.session_id}-${event.occurred_at}-${index}`} className="flex gap-3">
          <span className="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-blue-600" aria-hidden="true" />
          <div className="min-w-0">
            <div className="text-sm font-medium text-slate-950">{event.event_type}</div>
            <div className="text-xs text-slate-500">{formatDateTime(event.occurred_at)}</div>
            {event.path ? <div className="mt-1 truncate text-xs text-slate-600">{event.path}</div> : null}
          </div>
        </li>
      ))}
    </ol>
  );
}
