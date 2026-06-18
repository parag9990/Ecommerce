import { cn } from "../../../lib/cn";
import type { ProductStatus } from "../types";

type ProductStatusBadgeProps = {
  status: ProductStatus;
  className?: string;
};

const statusStyles: Record<ProductStatus, string> = {
  draft: "border-slate-200 bg-slate-100 text-slate-700",
  submitted: "border-blue-200 bg-blue-50 text-blue-700",
  approved: "border-emerald-200 bg-emerald-50 text-emerald-700",
  rejected: "border-rose-200 bg-rose-50 text-rose-700",
  published: "border-emerald-300 bg-emerald-100 text-emerald-800",
  unpublished: "border-amber-200 bg-amber-50 text-amber-700",
};

export function ProductStatusBadge({ status, className }: ProductStatusBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex h-6 items-center rounded border px-2 text-xs font-medium capitalize leading-none",
        statusStyles[status],
        className,
      )}
    >
      {status}
    </span>
  );
}
