import { CreditCard } from "lucide-react";
import { Link } from "react-router-dom";

import { AmountText } from "../../../components/ui/amount-text";
import { DataState } from "../../../components/ui/data-state";
import { StatusBadge } from "../../../components/ui/status-badge";
import { TablePagination } from "../../../components/ui/table-pagination";
import { formatDateTime } from "../../../lib/format";
import type { AdminPayment } from "../types";

export function PaymentTable({
  payments,
  isLoading,
  isFetching,
  error,
  page,
  limit,
  totalCount,
  onPageChange,
  onRetry,
  onSelect
}: {
  payments: AdminPayment[];
  isLoading: boolean;
  isFetching: boolean;
  error: unknown;
  page: number;
  limit: number;
  totalCount?: number;
  onPageChange: (page: number) => void;
  onRetry: () => void;
  onSelect: (payment: AdminPayment) => void;
}) {
  if (isLoading) {
    return <DataState title="Loading payments" description="Fetching payment status records." />;
  }

  if (error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load payments"
        description={error instanceof Error ? error.message : "The payment list request failed."}
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

  if (payments.length === 0) {
    return <DataState title="No payments found" description="Try a different status, provider, or order id." />;
  }

  return (
    <div className="min-h-0 flex-1 overflow-hidden rounded-lg border border-slate-200 bg-white">
      <div className="min-h-0 overflow-auto">
        <table className="min-w-full border-separate border-spacing-0 text-left text-sm">
          <thead className="sticky top-0 bg-slate-100 text-xs uppercase text-slate-600">
            <tr>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Payment</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Order</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Provider</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Status</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Amount</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Created</th>
              <th className="border-b border-slate-200 px-4 py-3 text-right font-semibold">Action</th>
            </tr>
          </thead>
          <tbody>
            {payments.map((payment) => (
              <tr key={payment.payment_id} className="hover:bg-slate-50">
                <td className="border-b border-slate-100 px-4 py-3">
                  <div className="flex min-w-0 items-center gap-3">
                    <span className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                      <CreditCard size={17} aria-hidden="true" />
                    </span>
                    <button
                      type="button"
                      onClick={() => onSelect(payment)}
                      className="min-w-0 text-left font-medium text-slate-950 hover:text-blue-700"
                    >
                      <span className="block max-w-[220px] truncate">{payment.payment_id}</span>
                      <span className="block max-w-[220px] truncate text-xs font-normal text-slate-500">
                        {payment.provider_payment_id ?? "No provider payment id"}
                      </span>
                    </button>
                  </div>
                </td>
                <td className="border-b border-slate-100 px-4 py-3">
                  <Link
                    to={`/admin/orders/${encodeURIComponent(payment.order_id)}`}
                    className="font-medium text-blue-700 hover:text-blue-900"
                  >
                    {payment.order_id}
                  </Link>
                </td>
                <td className="border-b border-slate-100 px-4 py-3 uppercase text-slate-700">
                  {payment.provider}
                </td>
                <td className="border-b border-slate-100 px-4 py-3">
                  <StatusBadge status={payment.status} />
                </td>
                <td className="border-b border-slate-100 px-4 py-3 font-medium text-slate-950">
                  <AmountText money={payment.amount} />
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                  {formatDateTime(payment.created_at)}
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-right">
                  <button
                    type="button"
                    onClick={() => onSelect(payment)}
                    className="font-medium text-blue-700 hover:text-blue-900"
                  >
                    Inspect
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <TablePagination
        page={page}
        limit={limit}
        itemCount={payments.length}
        totalCount={totalCount}
        isFetching={isFetching}
        onPageChange={onPageChange}
      />
    </div>
  );
}
