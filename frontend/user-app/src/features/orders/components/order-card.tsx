import { Link } from 'react-router-dom';

import { formatMoney } from '../../../lib/format-money';
import { routePaths } from '../../../routes/route-paths';
import type { Order } from '../types';
import { OrderStatusBadge } from './order-status-badge';

type OrderCardProps = {
  order: Order;
};

function formatDate(value?: string) {
  if (!value) {
    return 'Date unavailable';
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleDateString('en-IN', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
}

export function OrderCard({ order }: OrderCardProps) {
  const itemCount = order.items?.length ?? 0;
  const orderId = order.order_id ?? '';

  return (
    <article className="rounded-md border border-slate-200 bg-white p-4">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="text-sm text-slate-500">Order #{orderId || 'pending'}</p>
          <h2 className="mt-1 font-semibold text-slate-950">
            {itemCount} item{itemCount === 1 ? '' : 's'}
          </h2>
          <p className="mt-1 text-sm text-slate-500">
            Placed on {formatDate(order.created_at)}
          </p>
        </div>

        <div className="text-right">
          <OrderStatusBadge status={order.status} />
          <p className="mt-2 font-semibold text-slate-950">
            {formatMoney(order.total)}
          </p>
        </div>
      </div>

      {orderId ? (
        <Link
          className="mt-4 inline-flex rounded text-sm font-semibold text-blue-700 underline-offset-4 hover:text-blue-800 hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          to={routePaths.orderDetail(orderId)}
        >
          View details
        </Link>
      ) : null}
    </article>
  );
}
