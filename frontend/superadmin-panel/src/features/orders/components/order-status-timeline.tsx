import { Timeline } from "../../../components/ui/timeline";
import { formatDateTime } from "../../../lib/format";
import type { OrderStatusEvent } from "../types";

export function OrderStatusTimeline({ events }: { events: OrderStatusEvent[] }) {
  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4">
      <h2 className="text-base font-semibold text-slate-950">Status Timeline</h2>
      <p className="mt-1 text-sm text-slate-600">Raw order lifecycle events and admin review signals.</p>
      <div className="mt-4">
        <Timeline
          emptyLabel="No status history was returned for this order."
          items={events.map((event, index) => ({
            id: `${event.status}-${event.created_at ?? "unknown"}-${index}`,
            title: String(event.status).replace(/_/g, " "),
            description: (
              <>
                <span className="capitalize">{event.actor_type}</span>
                {event.actor_id ? <span> {event.actor_id}</span> : null}
                {event.note ? <span> - {event.note}</span> : null}
              </>
            ),
            timestamp: formatDateTime(event.created_at)
          }))}
        />
      </div>
    </section>
  );
}
