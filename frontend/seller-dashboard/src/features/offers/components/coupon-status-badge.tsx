import { cn } from "../../../lib/cn";
import type { CouponStatus } from "../types";
import { formatStatusLabel } from "../utils/offer-formatters";

type CouponStatusBadgeProps = {
  status: CouponStatus;
};

const statusStyles: Record<CouponStatus, string> = {
  draft: "border-slate-200 bg-slate-50 text-slate-700",
  active: "border-emerald-200 bg-emerald-50 text-emerald-700",
  paused: "border-amber-200 bg-amber-50 text-amber-700",
  expired: "border-rose-200 bg-rose-50 text-rose-700",
};

export function CouponStatusBadge({ status }: CouponStatusBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex h-5 items-center rounded border px-1.5 text-[11px] font-medium leading-none",
        statusStyles[status],
      )}
    >
      {formatStatusLabel(status)}
    </span>
  );
}
