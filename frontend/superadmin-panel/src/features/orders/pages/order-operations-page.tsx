import { useEffect, useMemo, useState } from "react";

import { PermissionDenied } from "../../../components/ui/permission-denied";
import { OrderFilterBar } from "../components/order-filter-bar";
import { OrderTable } from "../components/order-table";
import { useAdminOrders } from "../hooks/use-admin-orders";
import { useOrderPermissions } from "../permissions";
import type { AdminOrderFilters } from "../types";

const ORDER_PAGE_LIMIT = 25;

function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedValue(value), delayMs);

    return () => window.clearTimeout(timer);
  }, [delayMs, value]);

  return debouncedValue;
}

function initialFilters(): AdminOrderFilters {
  return {
    q: "",
    status: "all",
    review_status: "all",
    user_id: "",
    seller_id: "",
    from: "",
    to: "",
    page: 1,
    limit: ORDER_PAGE_LIMIT
  };
}

export function OrderOperationsPage() {
  const permissions = useOrderPermissions();
  const [filters, setFilters] = useState<AdminOrderFilters>(() => initialFilters());
  const debouncedQuery = useDebouncedValue(filters.q?.trim() ?? "", 350);
  const debouncedUserId = useDebouncedValue(filters.user_id?.trim() ?? "", 350);
  const debouncedSellerId = useDebouncedValue(filters.seller_id?.trim() ?? "", 350);
  const queryFilters = useMemo(
    () => ({
      ...filters,
      q: debouncedQuery,
      user_id: debouncedUserId,
      seller_id: debouncedSellerId
    }),
    [debouncedQuery, debouncedSellerId, debouncedUserId, filters]
  );
  const ordersQuery = useAdminOrders(queryFilters);

  if (!permissions.canViewOrders) {
    return <PermissionDenied compact />;
  }

  return (
    <section className="flex min-h-[calc(100vh-6.5rem)] flex-col overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
      <header className="border-b border-slate-200 bg-white px-4 py-4">
        <h1 className="text-xl font-semibold text-slate-950">Orders</h1>
        <p className="mt-1 text-sm text-slate-600">
          Search orders, inspect disputes, and route problematic orders through manual review.
        </p>
      </header>

      <OrderFilterBar
        filters={filters}
        onChange={(patch) =>
          setFilters((current) => ({
            ...current,
            ...patch,
            page: 1
          }))
        }
        onReset={() => setFilters(initialFilters())}
      />

      <div className="min-h-0 flex-1 p-4">
        <OrderTable
          orders={ordersQuery.data?.orders ?? []}
          isLoading={ordersQuery.isLoading}
          isFetching={ordersQuery.isFetching}
          error={ordersQuery.error}
          page={filters.page}
          limit={ORDER_PAGE_LIMIT}
          totalCount={ordersQuery.data?.total}
          onPageChange={(page) => setFilters((current) => ({ ...current, page }))}
          onRetry={() => void ordersQuery.refetch()}
        />
      </div>
    </section>
  );
}
