import { ArrowLeft } from 'lucide-react';
import { Link, useParams } from 'react-router-dom';
import { useState } from 'react';

import { Alert } from '../../../components/ui/alert';
import { Button } from '../../../components/ui/button';
import { formatMoney } from '../../../lib/format-money';
import { routePaths } from '../../../routes/route-paths';
import { CancelOrderDialog } from '../components/cancel-order-dialog';
import { OrderItemList } from '../components/order-item-list';
import { OrderStatusBadge } from '../components/order-status-badge';
import { OrderTimeline } from '../components/order-timeline';
import { useOrderDetailQuery } from '../hooks/use-order-detail-query';
import { useOrderMutations } from '../hooks/use-order-mutations';
import { canCancelOrder } from '../order-rules';

function getErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

function formatDateTime(value?: string) {
  if (!value) {
    return 'Date unavailable';
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

function OrderDetailSkeleton() {
  return (
    <div className="space-y-4">
      <div className="h-28 animate-pulse rounded-md bg-slate-200" />
      <div className="h-64 animate-pulse rounded-md bg-slate-200" />
    </div>
  );
}

export function OrderDetailPage() {
  const { orderId } = useParams();
  const [isCancelOpen, setIsCancelOpen] = useState(false);
  const [actionError, setActionError] = useState<string>();
  const [success, setSuccess] = useState<string>();
  const orderQuery = useOrderDetailQuery(orderId);
  const { cancel } = useOrderMutations(orderId);

  async function submitCancel(reason: string) {
    if (!orderId) {
      return;
    }

    setActionError(undefined);
    setSuccess(undefined);

    try {
      await cancel.mutateAsync(reason);
      setIsCancelOpen(false);
      setSuccess('Order cancelled.');
    } catch (error) {
      setActionError(getErrorMessage(error, 'Order could not be cancelled.'));
    }
  }

  if (!orderId) {
    return (
      <div className="mx-auto max-w-3xl">
        <Alert title="Order could not be loaded" variant="error">
          Order id is missing.
        </Alert>
      </div>
    );
  }

  if (orderQuery.isLoading) {
    return <OrderDetailSkeleton />;
  }

  const loadError =
    orderQuery.error instanceof Error
      ? orderQuery.error.message
      : orderQuery.isError
        ? 'Order could not be loaded.'
        : undefined;
  const order = orderQuery.data;

  if (!order) {
    return (
      <div className="mx-auto max-w-3xl space-y-4">
        <Alert title="Order could not be loaded" variant="error">
          {loadError ?? 'Please try again.'}
        </Alert>
        <Link
          className="inline-flex items-center rounded text-sm font-semibold text-blue-700 underline-offset-4 hover:text-blue-800 hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          to={routePaths.orders}
        >
          <ArrowLeft aria-hidden="true" className="mr-2 h-4 w-4" />
          Back to orders
        </Link>
      </div>
    );
  }

  const items = order.items ?? [];
  const canCancel = canCancelOrder(order.status);

  return (
    <section className="space-y-5">
      <Link
        className="inline-flex items-center rounded text-sm font-semibold text-blue-700 underline-offset-4 hover:text-blue-800 hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
        to={routePaths.orders}
      >
        <ArrowLeft aria-hidden="true" className="mr-2 h-4 w-4" />
        Back to orders
      </Link>

      <div className="rounded-md border border-slate-200 bg-white p-5">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <p className="text-sm text-slate-500">
              Order #{order.order_id ?? orderId}
            </p>
            <h1 className="mt-1 text-2xl font-semibold tracking-tight text-slate-950">
              {formatMoney(order.total)}
            </h1>
            <p className="mt-1 text-sm text-slate-500">
              Placed {formatDateTime(order.created_at)}
            </p>
          </div>
          <div className="text-right">
            <OrderStatusBadge status={order.status} />
            {order.payment_status ? (
              <p className="mt-2 text-sm text-slate-500">
                Payment: {order.payment_status.replaceAll('_', ' ')}
              </p>
            ) : null}
            {order.fulfillment_status ? (
              <p className="mt-1 text-sm text-slate-500">
                Fulfillment:{' '}
                {order.fulfillment_status.replaceAll('_', ' ')}
              </p>
            ) : null}
          </div>
        </div>

        {canCancel ? (
          <div className="mt-5">
            <Button
              className="bg-red-600 hover:bg-red-700 focus-visible:outline-red-600"
              onClick={() => {
                setIsCancelOpen(true);
              }}
              type="button"
            >
              Cancel order
            </Button>
          </div>
        ) : null}
      </div>

      {actionError ? <Alert variant="error">{actionError}</Alert> : null}
      {success ? <Alert variant="success">{success}</Alert> : null}

      <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <section className="space-y-3">
          <h2 className="text-lg font-semibold text-slate-950">Items</h2>
          <OrderItemList items={items} />
        </section>

        <aside className="space-y-3">
          <h2 className="text-lg font-semibold text-slate-950">Status</h2>
          <OrderTimeline order={order} />
        </aside>
      </div>

      {order.shipping_address ? (
        <section className="rounded-md border border-slate-200 bg-white p-4">
          <h2 className="text-lg font-semibold text-slate-950">
            Delivery address
          </h2>
          <address className="mt-2 not-italic text-sm leading-6 text-slate-600">
            {order.shipping_address.name ? (
              <>
                {order.shipping_address.name}
                <br />
              </>
            ) : null}
            {order.shipping_address.line1}
            {order.shipping_address.line2 ? (
              <>
                <br />
                {order.shipping_address.line2}
              </>
            ) : null}
            <br />
            {order.shipping_address.city}, {order.shipping_address.state}{' '}
            {order.shipping_address.postal_code}
            <br />
            {order.shipping_address.country}
          </address>
        </section>
      ) : null}

      {isCancelOpen ? (
        <CancelOrderDialog
          isSubmitting={cancel.isPending}
          onClose={() => {
            setIsCancelOpen(false);
          }}
          onSubmit={submitCancel}
        />
      ) : null}
    </section>
  );
}
