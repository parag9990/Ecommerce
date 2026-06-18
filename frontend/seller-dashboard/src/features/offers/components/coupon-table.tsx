import { ChevronLeft, ChevronRight, Pencil } from "lucide-react";
import { Link } from "react-router-dom";

import { EmptyState } from "../../../components/state/empty-state";
import type { Coupon, CouponFilters } from "../types";
import { formatDiscount, formatMoney, formatWindow } from "../utils/offer-formatters";
import { CouponStatusBadge } from "./coupon-status-badge";
import { UsageProgressBar } from "./usage-progress-bar";

type CouponTableProps = {
  coupons: Coupon[];
  total: number;
  filters: CouponFilters;
  onFiltersChange: (filters: CouponFilters) => void;
};

export function CouponTable({
  coupons,
  total,
  filters,
  onFiltersChange,
}: CouponTableProps) {
  const hasPreviousPage = filters.page > 1;
  const hasNextPage = filters.page * filters.page_size < total;

  if (coupons.length === 0) {
    return (
      <EmptyState
        title="No coupons found"
        description="Create a coupon or adjust the current filters."
      />
    );
  }

  return (
    <div className="overflow-hidden rounded-md border border-slate-200 bg-white shadow-sm">
      <div className="overflow-x-auto">
        <table className="w-full min-w-[980px] border-collapse text-left text-sm">
          <thead className="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-4 py-3 font-semibold">Code</th>
              <th className="px-4 py-3 font-semibold">Discount</th>
              <th className="px-4 py-3 font-semibold">Minimum cart</th>
              <th className="px-4 py-3 font-semibold">Validity</th>
              <th className="px-4 py-3 font-semibold">Usage</th>
              <th className="px-4 py-3 font-semibold">Status</th>
              <th className="px-4 py-3 text-right font-semibold">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {coupons.map((coupon) => (
              <tr key={coupon.coupon_id || coupon.code} className="align-middle">
                <td className="px-4 py-3">
                  <div className="font-medium text-slate-950">{coupon.code}</div>
                  {coupon.coupon_id ? (
                    <div className="mt-0.5 text-xs text-slate-500">{coupon.coupon_id}</div>
                  ) : null}
                </td>
                <td className="px-4 py-3 font-medium text-slate-800">
                  {formatDiscount(coupon)}
                </td>
                <td className="px-4 py-3 text-slate-600">
                  {coupon.min_cart_amount ? formatMoney(coupon.min_cart_amount) : "No minimum"}
                </td>
                <td className="px-4 py-3 text-slate-600">
                  {formatWindow(coupon.starts_at, coupon.ends_at)}
                </td>
                <td className="px-4 py-3">
                  <UsageProgressBar used={coupon.used_count} limit={coupon.usage_limit} />
                </td>
                <td className="px-4 py-3">
                  <CouponStatusBadge status={coupon.status} />
                </td>
                <td className="px-4 py-3 text-right">
                  <Link
                    to={`/seller/offers/coupons/${coupon.coupon_id}/edit`}
                    state={{ coupon }}
                    className="inline-flex h-8 items-center gap-1.5 rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
                  >
                    <Pencil className="h-3.5 w-3.5" aria-hidden="true" />
                    Edit
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 px-4 py-3 text-sm">
        <span className="text-slate-500">
          Page {filters.page} - {total} total
        </span>
        <div className="flex items-center gap-2">
          <button
            type="button"
            aria-label="Previous page"
            title="Previous page"
            disabled={!hasPreviousPage}
            onClick={() => onFiltersChange({ ...filters, page: filters.page - 1 })}
            className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <ChevronLeft className="h-4 w-4" aria-hidden="true" />
          </button>
          <button
            type="button"
            aria-label="Next page"
            title="Next page"
            disabled={!hasNextPage}
            onClick={() => onFiltersChange({ ...filters, page: filters.page + 1 })}
            className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <ChevronRight className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
      </div>
    </div>
  );
}
