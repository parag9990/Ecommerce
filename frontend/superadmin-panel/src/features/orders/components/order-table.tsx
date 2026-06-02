import { PackageSearch } from "lucide-react";
import { Link } from "react-router-dom";

import { DataState } from "../../../components/ui/data-state";
import { StatusBadge } from "../../../components/ui/status-badge";
import { TablePagination } from "../../../components/ui/table-pagination";
import { formatDateTime, formatMoney } from "../../../lib/format";
import type { AdminOrder } from "../types";

function sellerSummary(order: AdminOrder): string {
  const sellerIds = Array.from(
    new Set(order.items.map((item) => item.seller_id).filter((sellerId): sellerId is string => Boolean(sellerId)))
  );

  if (sellerIds.length === 0) {
    return "Unknown";
  }

  if (sellerIds.length <= 2) {
    return sellerIds.join(", ");
  }

  return `${sellerIds.slice(0, 2).join(", ")} +${sellerIds.length - 2}`;
}

export function OrderTable({
  orders,
  isLoading,
  isFetching,
  error,
  page,
  limit,
  totalCount,
  onPageChange,
  onRetry
}: {
  orders: AdminOrder[];
  isLoading: boolean;
  isFetching: boolean;
  error: unknown;
  page: number;
  limit: number;
  totalCount?: number;
  onPageChange: (page: number) => void;
  onRetry: () => void;
}) {
  if (isLoading) {
    return <DataState title="Loading orders" description="Fetching cross-platform order records." />;
  }

  if (error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load orders"
        description={error instanceof Error ? error.message : "The order list request failed."}
        action={
          <button
            type="button"
            onClick={onRetry}
            className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
          >
            Retry
          </button>
        }
      />
    );
  }

  if (orders.length === 0) {
    return <DataState title="No orders found" description="Try a different search, status, or id filter." />;
  }

  return (
    <div className="min-h-0 flex-1 overflow-hidden rounded-lg border border-slate-200 bg-white">
      <div className="min-h-0 overflow-auto">
        <table className="min-w-full border-separate border-spacing-0 text-left text-sm">
          <thead className="sticky top-0 bg-slate-100 text-xs uppercase text-slate-600">
            <tr>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Order</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Status</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Review</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Buyer</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Sellers</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Items</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Total</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Created</th>
              <th className="border-b border-slate-200 px-4 py-3 text-right font-semibold">Action</th>
            </tr>
          </thead>
          <tbody>
            {orders.map((order) => (
              <tr key={order.order_id} className="hover:bg-slate-50">
                <td className="border-b border-slate-100 px-4 py-3">
                  <div className="flex min-w-0 items-center gap-3">
                    <span className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                      <PackageSearch size={17} aria-hidden="true" />
                    </span>
                    <div className="min-w-0">
                      <Link
                        to={`/admin/orders/${encodeURIComponent(order.order_id)}`}
                        className="block truncate font-medium text-slate-950 hover:text-blue-700"
                      >
                        {order.order_id}
                      </Link>
                      <div className="truncate text-xs text-slate-500">
                        {order.payment?.payment_id ?? "No payment id"}
                      </div>
                    </div>
                  </div>
                </td>
                <td className="border-b border-slate-100 px-4 py-3">
                  <StatusBadge status={order.status} />
                </td>
                <td className="border-b border-slate-100 px-4 py-3">
                  <StatusBadge status={order.review_status ?? "none"} />
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                  <div className="max-w-[180px] truncate">{order.user_id}</div>
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                  <div className="max-w-[220px] truncate">{sellerSummary(order)}</div>
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                  {order.items.length}
                </td>
                <td className="border-b border-slate-100 px-4 py-3 font-medium text-slate-950">
                  {formatMoney(order.total)}
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                  {formatDateTime(order.created_at)}
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-right">
                  <Link
                    to={`/admin/orders/${encodeURIComponent(order.order_id)}`}
                    className="font-medium text-blue-700 hover:text-blue-900"
                  >
                    Investigate
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <TablePagination
        page={page}
        limit={limit}
        itemCount={orders.length}
        totalCount={totalCount}
        isFetching={isFetching}
        onPageChange={onPageChange}
      />
    </div>
  );
}
