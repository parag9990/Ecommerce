import { RotateCcw } from "lucide-react";
import { Link } from "react-router-dom";

import { AmountText } from "../../../components/ui/amount-text";
import { DataState } from "../../../components/ui/data-state";
import { StatusBadge } from "../../../components/ui/status-badge";
import { TablePagination } from "../../../components/ui/table-pagination";
import { formatDateTime } from "../../../lib/format";
import type { Refund } from "../types";

function canReviewStatus(refund: Refund): boolean {
  return ["requested", "pending_review"].includes(refund.status);
}

export function RefundTable({
  refunds,
  isLoading,
  isFetching,
  error,
  page,
  limit,
  totalCount,
  hidePagination = false,
  canReview,
  onPageChange,
  onRetry,
  onReview
}: {
  refunds: Refund[];
  isLoading: boolean;
  isFetching: boolean;
  error: unknown;
  page: number;
  limit: number;
  totalCount?: number;
  hidePagination?: boolean;
  canReview: boolean;
  onPageChange: (page: number) => void;
  onRetry: () => void;
  onReview: (refund: Refund) => void;
}) {
  if (isLoading) {
    return <DataState title="Loading refunds" description="Fetching refund review queue." />;
  }

  if (error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load refunds"
        description={error instanceof Error ? error.message : "The refund list request failed."}
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

  if (refunds.length === 0) {
    return <DataState title="No refunds found" description="There are no refund records for this view." />;
  }

  return (
    <div className="overflow-hidden rounded-lg border border-slate-200 bg-white">
      <div className="overflow-auto">
        <table className="min-w-full border-separate border-spacing-0 text-left text-sm">
          <thead className="sticky top-0 bg-slate-100 text-xs uppercase text-slate-600">
            <tr>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Refund</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Payment</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Order</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Status</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Amount</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Requested By</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Created</th>
              <th className="border-b border-slate-200 px-4 py-3 text-right font-semibold">Action</th>
            </tr>
          </thead>
          <tbody>
            {refunds.map((refund) => (
              <tr key={refund.refund_id} className="hover:bg-slate-50">
                <td className="border-b border-slate-100 px-4 py-3">
                  <div className="flex min-w-0 items-center gap-3">
                    <span className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                      <RotateCcw size={17} aria-hidden="true" />
                    </span>
                    <div className="min-w-0">
                      <div className="max-w-[220px] truncate font-medium text-slate-950">
                        {refund.refund_id}
                      </div>
                      <div className="max-w-[220px] truncate text-xs text-slate-500">{refund.reason}</div>
                    </div>
                  </div>
                </td>
                <td className="border-b border-slate-100 px-4 py-3">
                  <button
                    type="button"
                    onClick={() => onReview(refund)}
                    className="break-all text-left font-medium text-blue-700 hover:text-blue-900"
                  >
                    {refund.payment_id}
                  </button>
                </td>
                <td className="border-b border-slate-100 px-4 py-3">
                  {refund.order_id ? (
                    <Link
                      to={`/admin/orders/${encodeURIComponent(refund.order_id)}`}
                      className="font-medium text-blue-700 hover:text-blue-900"
                    >
                      {refund.order_id}
                    </Link>
                  ) : (
                    "Not linked"
                  )}
                </td>
                <td className="border-b border-slate-100 px-4 py-3">
                  <StatusBadge status={refund.status} />
                </td>
                <td className="border-b border-slate-100 px-4 py-3 font-medium text-slate-950">
                  <AmountText money={refund.amount} />
                </td>
                <td className="border-b border-slate-100 px-4 py-3 capitalize text-slate-700">
                  {refund.requested_by}
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                  {formatDateTime(refund.created_at)}
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-right">
                  {canReview && canReviewStatus(refund) ? (
                    <button
                      type="button"
                      onClick={() => onReview(refund)}
                      className="font-medium text-blue-700 hover:text-blue-900"
                    >
                      Review
                    </button>
                  ) : (
                    <span className="text-slate-500">View only</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {!hidePagination ? (
        <TablePagination
          page={page}
          limit={limit}
          itemCount={refunds.length}
          totalCount={totalCount}
          isFetching={isFetching}
          onPageChange={onPageChange}
        />
      ) : null}
    </div>
  );
}
