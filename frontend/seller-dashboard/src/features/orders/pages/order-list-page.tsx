import { useState } from "react";

import { AsyncStateBoundary } from "../../../components/state/async-state-boundary";
import { EmptyState } from "../../../components/state/empty-state";
import { TableSkeleton } from "../../../components/state/loading-skeleton";
import { PermissionDeniedState } from "../../../components/state/permission-denied-state";
import { RefreshingNotice } from "../../../components/state/refreshing-notice";
import { useSellerStore } from "../../../stores/seller-store";
import { useSellerPermissions } from "../../team/hooks/use-seller-permissions";
import { OrderFiltersBar } from "../components/order-filters-bar";
import { OrderTable } from "../components/order-table";
import { useSellerOrders } from "../hooks/use-seller-orders";
import type { SellerOrderFilters } from "../types";

const defaultFilters: SellerOrderFilters = {
  status: "all",
  page: 1,
  page_size: 20,
};

export function OrderListPage() {
  const activeSeller = useSellerStore((state) => state.activeSeller);
  const permissions = useSellerPermissions();
  const canViewOrders = permissions.can("orders:view");
  const [filters, setFilters] = useState<SellerOrderFilters>(defaultFilters);
  const ordersQuery = useSellerOrders(
    filters,
    canViewOrders ? activeSeller?.seller_id : undefined,
  );

  const hasFilters =
    filters.status !== "all" ||
    Boolean(filters.q?.trim()) ||
    Boolean(filters.date_from) ||
    Boolean(filters.date_to);

  function clearFilters() {
    setFilters(defaultFilters);
  }

  if (!activeSeller) {
    return (
      <PermissionDeniedState
        title="Active seller unavailable"
        description="Orders load karne ke liye active seller context required hai."
      />
    );
  }

  if (!canViewOrders) {
    return (
      <PermissionDeniedState
        title="Orders access unavailable"
        description="Aapke current seller role ke paas seller orders dekhne ka permission nahi hai."
      />
    );
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Fulfillment
          </p>
          <h1 className="mt-1 text-xl font-semibold text-slate-950">Orders</h1>
        </div>
      </div>

      <OrderFiltersBar filters={filters} onChange={setFilters} />

      <RefreshingNotice show={ordersQuery.isFetching && !ordersQuery.isPending} />

      <AsyncStateBoundary
        isLoading={ordersQuery.isPending}
        isError={ordersQuery.isError}
        error={ordersQuery.error}
        data={ordersQuery.data}
        isEmpty={(data) => data.orders.length === 0}
        loadingFallback={<TableSkeleton columns={8} />}
        emptyFallback={
          <EmptyState
            title="No orders found"
            description={
              hasFilters
                ? "Selected filters ke liye seller orders nahi mile."
                : "Buyer orders aate hi yahan dikhenge."
            }
            action={
              hasFilters
                ? {
                    label: "Clear filters",
                    onClick: clearFilters,
                    variant: "secondary",
                  }
                : undefined
            }
          />
        }
        onRetry={() => ordersQuery.refetch()}
      >
        {(data) => (
          <OrderTable
            orders={data.orders}
            total={data.total}
            filters={filters}
            onFiltersChange={setFilters}
          />
        )}
      </AsyncStateBoundary>
    </section>
  );
}
