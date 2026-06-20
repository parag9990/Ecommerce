import { AlertTriangle, CreditCard, RotateCcw, WalletCards } from "lucide-react";

import { AmountText } from "../../../components/ui/amount-text";
import type { AdminPayment, ReconciliationAlert, Refund } from "../types";

function sumAmounts(items: Array<{ amount: { amount: number; currency: string } }>) {
  const firstCurrency = items[0]?.amount.currency ?? "INR";
  const amount = items.reduce((total, item) => total + item.amount.amount, 0);

  return { amount, currency: firstCurrency };
}

export function PaymentStatusSummary({
  payments,
  refunds,
  reconciliationAlerts
}: {
  payments: AdminPayment[];
  refunds: Refund[];
  reconciliationAlerts: ReconciliationAlert[];
}) {
  const capturedPayments = payments.filter((payment) => payment.status === "captured");
  const failedPayments = payments.filter((payment) => payment.status === "failed");
  const reviewRefunds = refunds.filter((refund) =>
    ["requested", "pending_review"].includes(refund.status)
  );
  const mismatchAlerts = reconciliationAlerts.filter((alert) => alert.status !== "matched");

  const metrics = [
    {
      label: "Captured volume",
      value: <AmountText money={sumAmounts(capturedPayments)} />,
      helper: `${capturedPayments.length} captured on this page`,
      icon: WalletCards
    },
    {
      label: "Failed payments",
      value: failedPayments.length,
      helper: "Requires provider or buyer retry follow-up",
      icon: CreditCard
    },
    {
      label: "Refund review",
      value: reviewRefunds.length,
      helper: "Requested or pending finance decision",
      icon: RotateCcw
    },
    {
      label: "Recon alerts",
      value: mismatchAlerts.length,
      helper: "Open provider/local mismatches",
      icon: AlertTriangle
    }
  ];

  return (
    <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
      {metrics.map((metric) => {
        const Icon = metric.icon;

        return (
          <section key={metric.label} className="rounded-lg border border-slate-200 bg-white p-4">
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="text-sm font-medium text-slate-600">{metric.label}</p>
                <div className="mt-2 text-2xl font-semibold text-slate-950">{metric.value}</div>
              </div>
              <span className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                <Icon className="h-5 w-5" aria-hidden="true" />
              </span>
            </div>
            <p className="mt-2 text-xs text-slate-500">{metric.helper}</p>
          </section>
        );
      })}
    </div>
  );
}
