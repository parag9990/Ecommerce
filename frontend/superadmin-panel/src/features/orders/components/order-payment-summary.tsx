import { CreditCard } from "lucide-react";

import { StatusBadge } from "../../../components/ui/status-badge";
import { formatMoney } from "../../../lib/format";
import type { OrderPaymentSummary as PaymentSummary } from "../types";

export function OrderPaymentSummary({ payment }: { payment?: PaymentSummary | null }) {
  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="flex items-center gap-2">
        <CreditCard className="h-4 w-4 text-slate-600" aria-hidden="true" />
        <h2 className="text-base font-semibold text-slate-950">Payment Summary</h2>
      </div>
      <p className="mt-1 text-sm text-slate-600">Read-only context for order investigation.</p>

      {!payment ? (
        <p className="mt-3 text-sm text-slate-600">No payment summary is available.</p>
      ) : (
        <dl className="mt-4 grid gap-3 text-sm">
          <div className="flex justify-between gap-3">
            <dt className="text-slate-600">Payment ID</dt>
            <dd className="break-all text-right font-medium text-slate-950">
              {payment.payment_id ?? "Not added"}
            </dd>
          </div>
          <div className="flex justify-between gap-3">
            <dt className="text-slate-600">Provider</dt>
            <dd className="text-right text-slate-700">{payment.provider ?? "Not added"}</dd>
          </div>
          <div className="flex items-center justify-between gap-3">
            <dt className="text-slate-600">Payment status</dt>
            <dd>{payment.status ? <StatusBadge status={payment.status} /> : "Not added"}</dd>
          </div>
          <div className="flex items-center justify-between gap-3">
            <dt className="text-slate-600">Refund status</dt>
            <dd>{payment.refund_status ? <StatusBadge status={payment.refund_status} /> : "Not added"}</dd>
          </div>
          <div className="flex justify-between gap-3">
            <dt className="text-slate-600">Amount</dt>
            <dd className="font-medium text-slate-950">{formatMoney(payment.amount)}</dd>
          </div>
        </dl>
      )}
    </section>
  );
}
