import type { Campaign } from "../types";
import { formatMoney, formatWindow } from "../utils/offer-formatters";
import { CampaignStatusBadge } from "./campaign-status-badge";

type CampaignCardProps = {
  campaign: Campaign;
};

export function CampaignCard({ campaign }: CampaignCardProps) {
  return (
    <div className="space-y-1 rounded border border-blue-100 bg-blue-50 px-2 py-1">
      <div className="flex min-w-0 items-start justify-between gap-2">
        <div className="min-w-0 truncate text-xs font-medium text-blue-950">
          {campaign.name || "Untitled campaign"}
        </div>
        <CampaignStatusBadge status={campaign.status} />
      </div>
      <div className="text-[11px] leading-4 text-blue-700">
        {formatWindow(campaign.starts_at, campaign.ends_at)}
      </div>
      {campaign.budget ? (
        <div className="text-[11px] leading-4 text-blue-700">
          Budget {formatMoney(campaign.budget)}
        </div>
      ) : null}
    </div>
  );
}
