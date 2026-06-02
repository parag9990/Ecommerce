import { cn } from "../../../lib/cn";
import type { CampaignStatus } from "../types";
import { formatStatusLabel } from "../utils/offer-formatters";

type CampaignStatusBadgeProps = {
  status: CampaignStatus;
};

const statusStyles: Record<CampaignStatus, string> = {
  draft: "border-slate-200 bg-slate-50 text-slate-700",
  active: "border-emerald-200 bg-emerald-50 text-emerald-700",
  paused: "border-amber-200 bg-amber-50 text-amber-700",
  completed: "border-blue-200 bg-blue-50 text-blue-700",
};

export function CampaignStatusBadge({ status }: CampaignStatusBadgeProps) {
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
