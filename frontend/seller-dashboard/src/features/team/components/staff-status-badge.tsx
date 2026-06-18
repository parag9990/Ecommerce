import { cn } from "../../../lib/cn";
import type { SellerStaffStatus } from "../types";
import { STAFF_STATUS_LABELS } from "../utils/team-formatters";

type StaffStatusBadgeProps = {
  status: SellerStaffStatus;
};

const statusStyles: Record<SellerStaffStatus, string> = {
  invited: "border-amber-200 bg-amber-50 text-amber-700",
  active: "border-emerald-200 bg-emerald-50 text-emerald-700",
  disabled: "border-slate-200 bg-slate-100 text-slate-600",
};

export function StaffStatusBadge({ status }: StaffStatusBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex h-6 items-center rounded border px-2 text-xs font-medium leading-none",
        statusStyles[status],
      )}
    >
      {STAFF_STATUS_LABELS[status]}
    </span>
  );
}
