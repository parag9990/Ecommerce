import { ReceiptText } from "lucide-react";

import type { Refund } from "../types";
import { formatDateTime, formatOrderStatus } from "../utils/order-formatters";
import { MoneyCell } from "./money-cell";

type RefundSummaryPanelProps = {
  refunds: Refund[];
};

export function RefundSummaryPanel({ refunds }: RefundSummaryPanelProps) {
  return (
    <section className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex items-center gap-2">
        <ReceiptText className="h-4 w-4 text-slate-500" aria-hidden="true" />
        <h2 className="text-sm font-semibold text-slate-950">Refunds</h2>
      </div>

      {refunds.length === 0 ? (
        <p className="mt-3 text-sm text-slate-500">No refunds linked with this order.</p>
      ) : (
        <div className="mt-3 divide-y divide-slate-100">
          {refunds.map((refund, index) => (
            <div
              key={refund.refund_id ?? `${refund.payment_id ?? "refund"}-${index}`}
              className="grid gap-1 py-3 text-sm"
            >
              <div className="flex items-center justify-between gap-3">
                <MoneyCell money={refund.amount} />
                <span className="inline-flex h-6 items-center rounded border border-amber-200 bg-amber-50 px-2 text-xs font-medium capitalize leading-none text-amber-700">
                  {formatOrderStatus(refund.status)}
                </span>
              </div>
              {refund.payment_id ? (
                <p className="text-xs text-slate-500">{refund.payment_id}</p>
              ) : null}
              {refund.reason ? (
                <p className="text-slate-600">{refund.reason}</p>
              ) : null}
              {refund.created_at ? (
                <p className="text-xs text-slate-500">
                  {formatDateTime(refund.created_at)}
                </p>
              ) : null}
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
