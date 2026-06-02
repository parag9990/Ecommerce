import { ReceiptText } from "lucide-react";

import { StatusBadge } from "../../../components/ui/status-badge";
import { formatDateTime, formatMoney } from "../../../lib/format";
import type { AdminOrder } from "../types";

export function OrderSummaryPanel({ order }: { order: AdminOrder }) {
  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-slate-950">Order Summary</h2>
          <p className="mt-1 text-sm text-slate-600">Fulfillment and review state are tracked separately.</p>
        </div>
        <span className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
          <ReceiptText className="h-5 w-5" aria-hidden="true" />
        </span>
      </div>

      <dl className="mt-4 grid gap-3 text-sm">
        <div className="flex items-center justify-between gap-3">
          <dt className="text-slate-600">Order status</dt>
          <dd>
            <StatusBadge status={order.status} />
          </dd>
        </div>
        <div className="flex items-center justify-between gap-3">
          <dt className="text-slate-600">Review status</dt>
          <dd>
            <StatusBadge status={order.review_status ?? "none"} />
          </dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-slate-600">Buyer user id</dt>
          <dd className="min-w-0 break-all text-right font-medium text-slate-950">{order.user_id}</dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-slate-600">Items</dt>
          <dd className="font-medium text-slate-950">{order.items.length}</dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-slate-600">Total</dt>
          <dd className="font-medium text-slate-950">{formatMoney(order.total)}</dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-slate-600">Created</dt>
          <dd className="text-right text-slate-700">{formatDateTime(order.created_at)}</dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-slate-600">Updated</dt>
          <dd className="text-right text-slate-700">{formatDateTime(order.updated_at)}</dd>
        </div>
      </dl>
    </section>
  );
}
