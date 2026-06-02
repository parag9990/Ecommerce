import { ChevronLeft, ChevronRight, Eye } from "lucide-react";
import { Link } from "react-router-dom";

import { EmptyState } from "../../../components/state/empty-state";
import type { Order, SellerOrderFilters } from "../types";
import { formatDateTime, formatItemCount } from "../utils/order-formatters";
import { MoneyCell } from "./money-cell";
import { OrderStatusBadge } from "./order-status-badge";

type OrderTableProps = {
  orders: Order[];
  total: number;
  filters: SellerOrderFilters;
  onFiltersChange: (filters: SellerOrderFilters) => void;
};

function getItemCount(order: Order) {
  return order.items.reduce((total, item) => total + (item.quantity ?? 1), 0);
}

function getShipmentLabel(order: Order) {
  const latestShipment = order.shipments?.[0];

  if (!latestShipment) {
    return "Not added";
  }

  if (latestShipment.tracking_number) {
    return latestShipment.tracking_number;
  }

  return latestShipment.carrier ?? "Added";
}

export function OrderTable({
  orders,
  total,
  filters,
  onFiltersChange,
}: OrderTableProps) {
  const hasPreviousPage = filters.page > 1;
  const hasNextPage = filters.page * filters.page_size < total;

  if (orders.length === 0) {
    return (
      <EmptyState
        title="No orders found"
        description="Try a different status, date range, or order ID."
      />
    );
  }

  return (
    <div className="overflow-hidden rounded-md border border-slate-200 bg-white shadow-sm">
      <div className="overflow-x-auto">
        <table className="w-full min-w-[960px] border-collapse text-left text-sm">
          <thead className="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-4 py-3 font-semibold">Order</th>
              <th className="px-4 py-3 font-semibold">Created</th>
              <th className="px-4 py-3 font-semibold">Items</th>
              <th className="px-4 py-3 font-semibold">Total</th>
              <th className="px-4 py-3 font-semibold">Status</th>
              <th className="px-4 py-3 font-semibold">Shipment</th>
              <th className="px-4 py-3 font-semibold">Refunds</th>
              <th className="px-4 py-3 text-right font-semibold">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {orders.map((order) => (
              <tr key={order.order_id} className="align-middle">
                <td className="px-4 py-3">
                  <div className="font-medium text-slate-950">{order.order_id}</div>
                  {order.user_id ? (
                    <div className="mt-0.5 text-xs text-slate-500">{order.user_id}</div>
                  ) : null}
                </td>
                <td className="px-4 py-3 text-slate-600">
                  {formatDateTime(order.created_at)}
                </td>
                <td className="px-4 py-3 text-slate-600">
                  {formatItemCount(getItemCount(order))}
                </td>
                <td className="px-4 py-3">
                  <MoneyCell money={order.total} />
                </td>
                <td className="px-4 py-3">
                  <OrderStatusBadge status={order.status} />
                </td>
                <td className="max-w-44 truncate px-4 py-3 text-slate-600">
                  {getShipmentLabel(order)}
                </td>
                <td className="px-4 py-3 text-slate-600">
                  {order.refunds?.length ?? 0}
                </td>
                <td className="px-4 py-3 text-right">
                  <Link
                    to={`/seller/orders/${order.order_id}`}
                    state={{ order }}
                    className="inline-flex h-8 items-center gap-1.5 rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
                  >
                    <Eye className="h-3.5 w-3.5" aria-hidden="true" />
                    View
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 px-4 py-3 text-sm">
        <span className="text-slate-500">
          Page {filters.page} - {total} total
        </span>
        <div className="flex items-center gap-2">
          <button
            type="button"
            aria-label="Previous page"
            title="Previous page"
            disabled={!hasPreviousPage}
            onClick={() => onFiltersChange({ ...filters, page: filters.page - 1 })}
            className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <ChevronLeft className="h-4 w-4" aria-hidden="true" />
          </button>
          <button
            type="button"
            aria-label="Next page"
            title="Next page"
            disabled={!hasNextPage}
            onClick={() => onFiltersChange({ ...filters, page: filters.page + 1 })}
            className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <ChevronRight className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
      </div>
    </div>
  );
}
