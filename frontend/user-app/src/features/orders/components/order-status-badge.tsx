import { formatOrderStatus } from '../order-rules';

const statusClasses: Record<string, string> = {
  cancelled: 'bg-red-50 text-red-700',
  created: 'bg-slate-100 text-slate-700',
  delivered: 'bg-emerald-100 text-emerald-800',
  packed: 'bg-blue-50 text-blue-700',
  paid: 'bg-emerald-50 text-emerald-700',
  pending_payment: 'bg-amber-50 text-amber-700',
  refunded: 'bg-purple-50 text-purple-700',
  shipped: 'bg-indigo-50 text-indigo-700',
};

export function OrderStatusBadge({ status }: { status?: string | undefined }) {
  const normalizedStatus = status ?? 'unknown';

  return (
    <span
      className={[
        'inline-flex rounded-full px-2 py-1 text-xs font-semibold uppercase',
        statusClasses[normalizedStatus] ?? 'bg-slate-100 text-slate-700',
      ].join(' ')}
    >
      {formatOrderStatus(status)}
    </span>
  );
}
