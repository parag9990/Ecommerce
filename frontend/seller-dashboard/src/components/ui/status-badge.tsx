import type { SellerStatus } from "../../api/seller-session-api";
import { cn } from "../../lib/cn";

type StatusBadgeProps = {
  status: SellerStatus;
  className?: string;
};

const statusStyles: Record<SellerStatus, string> = {
  active: "border-emerald-200 bg-emerald-50 text-emerald-700",
  pending: "border-amber-200 bg-amber-50 text-amber-700",
  suspended: "border-rose-200 bg-rose-50 text-rose-700",
};

export function StatusBadge({ status, className }: StatusBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex h-5 items-center rounded border px-1.5 text-[11px] font-medium capitalize leading-none",
        statusStyles[status],
        className,
      )}
    >
      {status}
    </span>
  );
}
