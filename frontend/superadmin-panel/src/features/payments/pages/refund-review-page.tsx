import { ArrowLeft } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";

import { PermissionDenied } from "../../../components/ui/permission-denied";
import { RefundReviewDrawer } from "../components/refund-review-drawer";
import { RefundTable } from "../components/refund-table";
import { useRefunds } from "../hooks/use-refunds";
import { usePaymentPermissions } from "../permissions";
import type { Refund, RefundFilters } from "../types";

const REFUND_PAGE_SIZE = 25;

function initialRefundFilters(): RefundFilters {
  return {
    status: "pending_review",
    page: 1,
    page_size: REFUND_PAGE_SIZE
  };
}

export function RefundReviewPage() {
  const permissions = usePaymentPermissions();
  const [filters, setFilters] = useState<RefundFilters>(() => initialRefundFilters());
  const [selectedRefund, setSelectedRefund] = useState<Refund | null>(null);
  const refundsQuery = useRefunds(filters);

  if (!permissions.canViewRefunds) {
    return <PermissionDenied compact />;
  }

  return (
    <section className="flex min-h-[calc(100vh-6.5rem)] flex-col overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
      <header className="border-b border-slate-200 bg-white px-4 py-4">
        <Link
          to="/admin/payments"
          className="mb-3 inline-flex items-center gap-2 text-sm font-medium text-slate-600 hover:text-slate-950"
        >
          <ArrowLeft size={16} aria-hidden="true" />
          Back to payments
        </Link>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h1 className="text-xl font-semibold text-slate-950">Refund Review</h1>
            <p className="mt-1 text-sm text-slate-600">
              Finance queue for approving or rejecting refund requests with mandatory audit reasons.
            </p>
          </div>
          <label>
            <span className="sr-only">Refund status</span>
            <select
              value={filters.status ?? "all"}
              onChange={(event) =>
                setFilters((current) => ({
                  ...current,
                  status: event.target.value as RefundFilters["status"],
                  page: 1
                }))
              }
              className="h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
            >
              <option value="all">All refunds</option>
              <option value="requested">Requested</option>
              <option value="pending_review">Pending review</option>
              <option value="approved">Approved</option>
              <option value="rejected">Rejected</option>
              <option value="processing">Processing</option>
              <option value="succeeded">Succeeded</option>
              <option value="failed">Failed</option>
            </select>
          </label>
        </div>
      </header>

      <div className="min-h-0 flex-1 p-4">
        <RefundTable
          refunds={refundsQuery.data?.refunds ?? []}
          isLoading={refundsQuery.isLoading}
          isFetching={refundsQuery.isFetching}
          error={refundsQuery.error}
          page={filters.page}
          limit={REFUND_PAGE_SIZE}
          totalCount={refundsQuery.data?.total}
          canReview={permissions.canReviewRefunds}
          onPageChange={(page) => setFilters((current) => ({ ...current, page }))}
          onRetry={() => void refundsQuery.refetch()}
          onReview={setSelectedRefund}
        />
      </div>

      <RefundReviewDrawer
        refund={selectedRefund}
        canReview={permissions.canReviewRefunds}
        onClose={() => setSelectedRefund(null)}
      />
    </section>
  );
}
