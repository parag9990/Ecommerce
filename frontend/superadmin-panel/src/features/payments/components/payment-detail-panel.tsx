import { X } from "lucide-react";
import { Link } from "react-router-dom";

import { AmountText } from "../../../components/ui/amount-text";
import { DataState } from "../../../components/ui/data-state";
import { StatusBadge } from "../../../components/ui/status-badge";
import { formatDateTime } from "../../../lib/format";
import { usePaymentDetail } from "../hooks/use-payment-detail";
import { PaymentAttemptsTimeline } from "./payment-attempts-timeline";
import { RefundTable } from "./refund-table";

export function PaymentDetailPanel({
  paymentId,
  onClose
}: {
  paymentId: string | null;
  onClose: () => void;
}) {
  const detailQuery = usePaymentDetail(paymentId);

  if (!paymentId) {
    return null;
  }

  const payment = detailQuery.data?.payment;
  const attempts = detailQuery.data?.attempts ?? payment?.attempts ?? [];
  const refunds = detailQuery.data?.refunds ?? payment?.refunds ?? [];

  return (
    <div className="fixed inset-0 z-40">
      <button
        type="button"
        className="absolute inset-0 bg-slate-950/30"
        aria-label="Close payment detail"
        onClick={onClose}
      />
      <aside className="absolute inset-y-0 right-0 flex w-full max-w-2xl flex-col overflow-hidden border-l border-slate-200 bg-slate-50 shadow-xl">
        <header className="border-b border-slate-200 bg-white px-4 py-4">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h2 className="text-lg font-semibold text-slate-950">Payment Detail</h2>
              <p className="mt-1 break-all text-sm text-slate-600">{paymentId}</p>
            </div>
            <button
              type="button"
              onClick={onClose}
              title="Close payment detail"
              className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-100"
            >
              <X className="h-4 w-4" aria-hidden="true" />
              <span className="sr-only">Close payment detail</span>
            </button>
          </div>
        </header>

        <div className="min-h-0 flex-1 overflow-auto p-4">
          {detailQuery.isLoading ? (
            <DataState title="Loading payment detail" description="Fetching attempts and refund context." />
          ) : null}

          {detailQuery.error ? (
            <DataState
              tone="danger"
              title="Unable to load payment detail"
              description={
                detailQuery.error instanceof Error
                  ? detailQuery.error.message
                  : "The payment detail request failed."
              }
              action={
                <button
                  type="button"
                  onClick={() => void detailQuery.refetch()}
                  className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
                >
                  Retry
                </button>
              }
            />
          ) : null}

          {payment ? (
            <div className="space-y-4">
              <section className="rounded-lg border border-slate-200 bg-white p-4">
                <h3 className="text-base font-semibold text-slate-950">Status Overview</h3>
                <p className="mt-1 text-sm text-slate-600">
                  Financial metadata only. Raw card data, provider secrets, and client tokens are never shown.
                </p>
                <dl className="mt-4 grid gap-3 text-sm">
                  <div className="flex items-center justify-between gap-3">
                    <dt className="text-slate-600">Status</dt>
                    <dd>
                      <StatusBadge status={payment.status} />
                    </dd>
                  </div>
                  <div className="flex justify-between gap-3">
                    <dt className="text-slate-600">Amount</dt>
                    <dd className="font-medium text-slate-950">
                      <AmountText money={payment.amount} />
                    </dd>
                  </div>
                  <div className="flex justify-between gap-3">
                    <dt className="text-slate-600">Order</dt>
                    <dd className="break-all text-right font-medium">
                      <Link
                        to={`/admin/orders/${encodeURIComponent(payment.order_id)}`}
                        className="text-blue-700 hover:text-blue-900"
                      >
                        {payment.order_id}
                      </Link>
                    </dd>
                  </div>
                  <div className="flex justify-between gap-3">
                    <dt className="text-slate-600">Provider</dt>
                    <dd className="text-right uppercase text-slate-700">{payment.provider}</dd>
                  </div>
                  <div className="flex justify-between gap-3">
                    <dt className="text-slate-600">Provider payment ID</dt>
                    <dd className="break-all text-right text-slate-700">
                      {payment.provider_payment_id ?? "Not added"}
                    </dd>
                  </div>
                  <div className="flex justify-between gap-3">
                    <dt className="text-slate-600">Captured</dt>
                    <dd className="text-right text-slate-700">{formatDateTime(payment.captured_at)}</dd>
                  </div>
                  <div className="flex justify-between gap-3">
                    <dt className="text-slate-600">Created</dt>
                    <dd className="text-right text-slate-700">{formatDateTime(payment.created_at)}</dd>
                  </div>
                </dl>
              </section>

              <PaymentAttemptsTimeline attempts={attempts} />

              <div>
                <h3 className="text-base font-semibold text-slate-950">Linked Refunds</h3>
                <p className="mt-1 text-sm text-slate-600">Read-only refund context for this payment.</p>
                <div className="mt-3">
                  <RefundTable
                    refunds={refunds}
                    isLoading={false}
                    isFetching={false}
                    error={null}
                    page={1}
                    limit={refunds.length || 1}
                    hidePagination
                    canReview={false}
                    onPageChange={() => undefined}
                    onRetry={() => undefined}
                    onReview={() => undefined}
                  />
                </div>
              </div>
            </div>
          ) : null}
        </div>
      </aside>
    </div>
  );
}
