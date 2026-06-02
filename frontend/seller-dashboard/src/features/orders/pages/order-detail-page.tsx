import type { QueryClient } from "@tanstack/react-query";
import { useQueryClient } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useLocation, useParams } from "react-router-dom";

import { PermissionDeniedState } from "../../../components/state/permission-denied-state";
import { UnavailableState } from "../../../components/state/unavailable-state";
import { useSellerStore } from "../../../stores/seller-store";
import { useSellerPermissions } from "../../team/hooks/use-seller-permissions";
import { CustomerSummaryCard } from "../components/customer-summary-card";
import { MoneyCell } from "../components/money-cell";
import { OrderItemsTable } from "../components/order-items-table";
import { OrderStatusBadge } from "../components/order-status-badge";
import { RefundSummaryPanel } from "../components/refund-summary-panel";
import { ShipmentSummaryCard } from "../components/shipment-summary-card";
import { ShipmentUpdateForm } from "../components/shipment-update-form";
import { StatusHistoryList } from "../components/status-history-list";
import { orderQueryKeys } from "../hooks/query-keys";
import type { Order, OrderListResponse } from "../types";
import { formatDateTime, formatItemCount } from "../utils/order-formatters";
import { canUpdateFulfillment } from "../utils/order-status-rules";

type LocationState = {
  order?: Order;
};

function findCachedOrder(queryClient: QueryClient, orderId: string) {
  const detailOrder = queryClient.getQueryData<Order>(orderQueryKeys.detail(orderId));

  if (detailOrder) {
    return detailOrder;
  }

  const cachedLists = queryClient.getQueriesData<OrderListResponse>({
    queryKey: orderQueryKeys.lists(),
  });

  for (const [, list] of cachedLists) {
    const cachedOrder = list?.orders.find((order) => order.order_id === orderId);

    if (cachedOrder) {
      return cachedOrder;
    }
  }

  return undefined;
}

function getItemCount(order: Order) {
  return order.items.reduce((total, item) => total + (item.quantity ?? 1), 0);
}

export function OrderDetailPage() {
  const { orderId = "" } = useParams();
  const location = useLocation();
  const queryClient = useQueryClient();
  const activeSeller = useSellerStore((state) => state.activeSeller);
  const permissions = useSellerPermissions();
  const canViewOrders = permissions.can("orders:view");
  const canUpdateOrders = permissions.can("orders:update_fulfillment");
  const state = location.state as LocationState | null;
  const stateOrder = state?.order?.order_id === orderId ? state.order : undefined;
  const [order, setOrder] = useState<Order | undefined>(() =>
    stateOrder ?? findCachedOrder(queryClient, orderId),
  );

  useEffect(() => {
    setOrder(stateOrder ?? findCachedOrder(queryClient, orderId));
  }, [orderId, queryClient, stateOrder]);

  useEffect(() => {
    if (order) {
      queryClient.setQueryData(orderQueryKeys.detail(order.order_id), order);
    }
  }, [order, queryClient]);

  if (!activeSeller) {
    return (
      <PermissionDeniedState
        title="Active seller unavailable"
        description="Order detail load karne ke liye active seller context required hai."
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

  if (!order) {
    return (
      <UnavailableState
        title="Order detail unavailable"
        description="Open the order from the seller orders list. Direct detail fetch is not part of the current API contract."
        action={{ label: "Back to orders", href: "/seller/orders" }}
      />
    );
  }

  const editable = canUpdateFulfillment(order.status);

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 items-start gap-3">
          <Link
            to="/seller/orders"
            aria-label="Back to orders"
            title="Back to orders"
            className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          >
            <ArrowLeft className="h-4 w-4" aria-hidden="true" />
          </Link>
          <div className="min-w-0">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Order manager
            </p>
            <div className="mt-1 flex flex-wrap items-center gap-2">
              <h1 className="break-all text-xl font-semibold text-slate-950">
                {order.order_id}
              </h1>
              <OrderStatusBadge status={order.status} />
            </div>
            <p className="mt-1 text-sm text-slate-500">
              Created {formatDateTime(order.created_at)}
            </p>
          </div>
        </div>
      </div>

      <div className="grid gap-3 sm:grid-cols-3">
        <div className="rounded-md border border-slate-200 bg-white p-3 shadow-sm">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Total
          </p>
          <MoneyCell money={order.total} className="mt-1 block text-lg text-slate-950" />
        </div>
        <div className="rounded-md border border-slate-200 bg-white p-3 shadow-sm">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Items
          </p>
          <p className="mt-1 text-lg font-semibold text-slate-950">
            {formatItemCount(getItemCount(order))}
          </p>
        </div>
        <div className="rounded-md border border-slate-200 bg-white p-3 shadow-sm">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Refunds
          </p>
          <p className="mt-1 text-lg font-semibold text-slate-950">
            {order.refunds?.length ?? 0}
          </p>
        </div>
      </div>

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_360px]">
        <div className="space-y-4">
          <section className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
              <h2 className="text-sm font-semibold text-slate-950">Order items</h2>
              <MoneyCell money={order.total} />
            </div>
            <OrderItemsTable items={order.items} />
          </section>

          <StatusHistoryList history={order.status_history ?? []} />
          <RefundSummaryPanel refunds={order.refunds ?? []} />
        </div>

        <aside className="space-y-4">
          <CustomerSummaryCard
            customer={order.customer}
            shippingAddress={order.shipping_address}
          />
          <ShipmentSummaryCard shipments={order.shipments ?? []} />
          {editable && canUpdateOrders ? (
            <ShipmentUpdateForm order={order} onUpdated={setOrder} />
          ) : editable ? (
            <PermissionDeniedState
              title="Fulfillment permission required"
              description="Aapke current seller role ke paas fulfillment update karne ka permission nahi hai."
              action={null}
            />
          ) : (
            <div className="rounded-md border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
              Fulfillment updates are unavailable for this status.
            </div>
          )}
        </aside>
      </div>
    </section>
  );
}
