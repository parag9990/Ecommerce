import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";

import { PermissionDenied } from "../../../components/ui/permission-denied";
import { PaymentDetailPanel } from "../components/payment-detail-panel";
import { PaymentFilterBar } from "../components/payment-filter-bar";
import { PaymentStatusSummary } from "../components/payment-status-summary";
import { PaymentTable } from "../components/payment-table";
import { ReconciliationAlertPanel } from "../components/reconciliation-alert-panel";
import { ReconciliationTable } from "../components/reconciliation-table";
import { useAdminPayments } from "../hooks/use-admin-payments";
import { useReconciliationAlerts } from "../hooks/use-reconciliation-alerts";
import { useRefunds } from "../hooks/use-refunds";
import { usePaymentPermissions } from "../permissions";
import type { AdminPayment, PaymentFilters, ReconciliationAlert, ReconciliationFilters } from "../types";

const PAYMENT_PAGE_SIZE = 25;
const RECONCILIATION_PAGE_SIZE = 5;

function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedValue(value), delayMs);

    return () => window.clearTimeout(timer);
  }, [delayMs, value]);

  return debouncedValue;
}

function initialPaymentFilters(): PaymentFilters {
  return {
    order_id: "",
    status: "all",
    provider: "all",
    page: 1,
    page_size: PAYMENT_PAGE_SIZE
  };
}

export function PaymentOperationsPage() {
  const permissions = usePaymentPermissions();
  const [filters, setFilters] = useState<PaymentFilters>(() => initialPaymentFilters());
  const [reconciliationFilters, setReconciliationFilters] = useState<ReconciliationFilters>({
    status: "mismatch",
    page: 1,
    page_size: RECONCILIATION_PAGE_SIZE
  });
  const [selectedPaymentId, setSelectedPaymentId] = useState<string | null>(null);
  const [selectedReconciliationId, setSelectedReconciliationId] = useState<string | null>(null);
  const debouncedOrderId = useDebouncedValue(filters.order_id?.trim() ?? "", 350);
  const queryFilters = useMemo(
    () => ({
      ...filters,
      order_id: debouncedOrderId
    }),
    [debouncedOrderId, filters]
  );
  const paymentsQuery = useAdminPayments(queryFilters);
  const refundsQuery = useRefunds({ status: "pending_review", page: 1, page_size: 25 });
  const reconciliationQuery = useReconciliationAlerts(reconciliationFilters);

  if (!permissions.canViewPayments) {
    return <PermissionDenied compact />;
  }

  const payments = paymentsQuery.data?.payments ?? [];
  const refunds = refundsQuery.data?.refunds ?? [];
  const alerts = reconciliationQuery.data?.alerts ?? [];

  return (
    <section className="flex min-h-[calc(100vh-6.5rem)] flex-col overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
      <header className="border-b border-slate-200 bg-white px-4 py-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h1 className="text-xl font-semibold text-slate-950">Payments</h1>
            <p className="mt-1 text-sm text-slate-600">
              Search payment status, inspect attempts, review refunds, and watch reconciliation mismatches.
            </p>
          </div>
          <Link
            to="/admin/payments/refunds"
            className="inline-flex h-10 items-center justify-center rounded-lg bg-slate-950 px-3 text-sm font-medium text-white hover:bg-slate-800"
          >
            Refund review
          </Link>
        </div>
      </header>

      <div className="border-b border-slate-200 bg-slate-50 p-4">
        <PaymentStatusSummary payments={payments} refunds={refunds} reconciliationAlerts={alerts} />
      </div>

      <PaymentFilterBar
        filters={filters}
        onChange={(patch) =>
          setFilters((current) => ({
            ...current,
            ...patch,
            page: 1
          }))
        }
        onReset={() => setFilters(initialPaymentFilters())}
      />

      <div className="grid min-h-0 flex-1 gap-4 p-4 xl:grid-cols-[minmax(0,1fr)_420px]">
        <main className="min-h-0">
          <PaymentTable
            payments={payments}
            isLoading={paymentsQuery.isLoading}
            isFetching={paymentsQuery.isFetching}
            error={paymentsQuery.error}
            page={filters.page}
            limit={PAYMENT_PAGE_SIZE}
            totalCount={paymentsQuery.data?.total}
            onPageChange={(page) => setFilters((current) => ({ ...current, page }))}
            onRetry={() => void paymentsQuery.refetch()}
            onSelect={(payment: AdminPayment) => setSelectedPaymentId(payment.payment_id)}
          />
        </main>

        {permissions.canViewReconciliation ? (
          <aside className="min-h-0 space-y-3">
            <div className="rounded-lg border border-slate-200 bg-white p-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <h2 className="text-base font-semibold text-slate-950">Reconciliation Alerts</h2>
                  <p className="mt-1 text-sm text-slate-600">
                    Provider/local mismatches that need finance inspection.
                  </p>
                </div>
                <label>
                  <span className="sr-only">Reconciliation status</span>
                  <select
                    value={reconciliationFilters.status ?? "all"}
                    onChange={(event) =>
                      setReconciliationFilters((current) => ({
                        ...current,
                        status: event.target.value as ReconciliationFilters["status"],
                        page: 1
                      }))
                    }
                    className="h-9 rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
                  >
                    <option value="all">All</option>
                    <option value="mismatch">Mismatch</option>
                    <option value="missing_local">Missing local</option>
                    <option value="missing_provider">Missing provider</option>
                    <option value="matched">Matched</option>
                  </select>
                </label>
              </div>
            </div>
            <ReconciliationTable
              alerts={alerts}
              isLoading={reconciliationQuery.isLoading}
              isFetching={reconciliationQuery.isFetching}
              error={reconciliationQuery.error}
              page={reconciliationFilters.page}
              limit={RECONCILIATION_PAGE_SIZE}
              totalCount={reconciliationQuery.data?.total}
              compact
              onPageChange={(page) => setReconciliationFilters((current) => ({ ...current, page }))}
              onRetry={() => void reconciliationQuery.refetch()}
              onSelect={(alert: ReconciliationAlert) =>
                setSelectedReconciliationId(alert.reconciliation_id)
              }
            />
          </aside>
        ) : null}
      </div>

      <PaymentDetailPanel paymentId={selectedPaymentId} onClose={() => setSelectedPaymentId(null)} />
      <ReconciliationAlertPanel
        reconciliationId={selectedReconciliationId}
        onClose={() => setSelectedReconciliationId(null)}
      />
    </section>
  );
}
