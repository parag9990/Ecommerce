import { formatOrderStatus } from '../order-rules';
import type { Order } from '../types';

type OrderTimelineProps = {
  order: Order;
};

function formatDateTime(value?: string) {
  if (!value) {
    return 'Time unavailable';
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleString('en-IN', {
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    month: 'short',
    year: 'numeric',
  });
}

export function OrderTimeline({ order }: OrderTimelineProps) {
  const events =
    order.status_history && order.status_history.length > 0
      ? order.status_history
      : [
          {
            created_at: order.created_at,
            status: order.status,
          },
        ];

  return (
    <ol className="space-y-3 rounded-md border border-slate-200 bg-white p-4">
      {events.map((event, index) => (
        <li className="flex gap-3" key={`${event.status ?? 'status'}-${index}`}>
          <span className="mt-1 h-2.5 w-2.5 shrink-0 rounded-full bg-blue-600" />
          <span>
            <span className="block text-sm font-semibold capitalize text-slate-950">
              {formatOrderStatus(event.status)}
            </span>
            <span className="mt-1 block text-sm text-slate-500">
              {formatDateTime(event.created_at)}
            </span>
            {event.note ? (
              <span className="mt-1 block text-sm text-slate-600">
                {event.note}
              </span>
            ) : null}
          </span>
        </li>
      ))}
    </ol>
  );
}
