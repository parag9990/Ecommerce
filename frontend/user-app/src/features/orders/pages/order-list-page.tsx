import { Link } from 'react-router-dom';
import { useState } from 'react';

import { Alert } from '../../../components/ui/alert';
import { Button } from '../../../components/ui/button';
import { EmptyState } from '../../../components/ui/empty-state';
import { routePaths } from '../../../routes/route-paths';
import { OrderCard } from '../components/order-card';
import { useOrdersQuery } from '../hooks/use-orders-query';

const pageSize = 10;

function OrderListSkeleton() {
  return (
    <div className="space-y-4">
      {Array.from({ length: 3 }, (_, index) => (
        <div
          className="h-36 animate-pulse rounded-md border border-slate-200 bg-white"
          key={index}
        />
      ))}
    </div>
  );
}

export function OrderListPage() {
  const [page, setPage] = useState(1);
  const ordersQuery = useOrdersQuery(page, pageSize);
  const orders = ordersQuery.data?.orders ?? [];
  const total = ordersQuery.data?.total ?? orders.length;
  const error =
    ordersQuery.error instanceof Error
      ? ordersQuery.error.message
      : ordersQuery.isError
        ? 'Orders could not be loaded.'
        : undefined;

  const hasPreviousPage = page > 1;
  const hasNextPage = page * pageSize < total;
  const isEmpty = !ordersQuery.isLoading && orders.length === 0 && !error;

  return (
    <section className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-slate-950">
          Orders
        </h1>
        <p className="mt-1 text-sm text-slate-600">
          Track order history, payment status, and fulfilment progress.
        </p>
      </div>

      {error ? <Alert variant="error">{error}</Alert> : null}

      {ordersQuery.isLoading ? <OrderListSkeleton /> : null}

      {isEmpty ? (
        <EmptyState
          action={
            <Link
              className="inline-flex h-10 items-center justify-center rounded-md bg-blue-600 px-4 text-sm font-semibold text-white transition hover:bg-blue-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
              to={routePaths.home}
            >
              Browse products
            </Link>
          }
          description="Placed orders will appear here with their latest status."
          title="No orders yet"
        />
      ) : null}

      {!ordersQuery.isLoading && orders.length > 0 ? (
        <div className="space-y-4">
          {orders.map((order) => (
            <OrderCard
              key={order.order_id ?? `${order.created_at}-${order.status}`}
              order={order}
            />
          ))}

          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className="text-sm text-slate-500">
              Page {page} of {Math.max(1, Math.ceil(total / pageSize))}
            </p>
            <div className="flex gap-2">
              <Button
                disabled={!hasPreviousPage}
                onClick={() => {
                  setPage((current) => Math.max(1, current - 1));
                }}
                type="button"
                variant="secondary"
              >
                Previous
              </Button>
              <Button
                disabled={!hasNextPage}
                onClick={() => {
                  setPage((current) => current + 1);
                }}
                type="button"
                variant="secondary"
              >
                Next
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </section>
  );
}
