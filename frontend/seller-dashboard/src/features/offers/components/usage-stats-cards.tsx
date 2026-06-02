import { CalendarDays, CheckCircle2, Percent, Tag, TicketPercent } from "lucide-react";

import { cn } from "../../../lib/cn";
import type { Campaign, Coupon } from "../types";
import { formatMoney } from "../utils/offer-formatters";
import { buildUsageStats } from "../utils/usage-stats";

type UsageStatsCardsProps = {
  coupons?: Coupon[];
  campaigns?: Campaign[];
};

export function UsageStatsCards({ coupons, campaigns }: UsageStatsCardsProps) {
  const stats = buildUsageStats(coupons ?? [], campaigns ?? []);
  const cards = [
    {
      label: "Total coupons",
      value: coupons ? stats.totalCoupons : "Not available",
      icon: TicketPercent,
      unavailable: !coupons,
    },
    {
      label: "Active coupons",
      value: coupons ? stats.activeCoupons : "Not available",
      icon: CheckCircle2,
      unavailable: !coupons,
    },
    {
      label: "Expired coupons",
      value: coupons ? stats.expiredCoupons : "Not available",
      icon: Tag,
      unavailable: !coupons,
    },
    {
      label: "Scheduled campaigns",
      value: campaigns ? stats.scheduledCampaigns : "Not available",
      icon: CalendarDays,
      unavailable: !campaigns,
    },
    {
      label: "Known redemptions",
      value:
        coupons && stats.totalUsageLimit > 0
          ? `${stats.knownUsedCount}/${stats.totalUsageLimit}`
          : "Not available",
      icon: Percent,
      unavailable: !coupons || stats.totalUsageLimit === 0,
    },
  ];

  return (
    <section className="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
      {cards.map((card) => {
        const Icon = card.icon;

        return (
          <div
            key={card.label}
            className="rounded-md border border-slate-200 bg-white p-3 shadow-sm"
          >
            <div className="flex items-center justify-between gap-3">
              <span className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                {card.label}
              </span>
              <Icon className="h-4 w-4 shrink-0 text-slate-400" aria-hidden="true" />
            </div>
            <div
              className={cn(
                "mt-2 text-xl font-semibold",
                card.unavailable ? "text-slate-400" : "text-slate-950",
              )}
            >
              {card.value}
            </div>
          </div>
        );
      })}

      {coupons && stats.totalDiscountAmount !== undefined && stats.totalDiscountCurrency ? (
        <div className="rounded-md border border-slate-200 bg-white p-3 shadow-sm md:col-span-2 xl:col-span-5">
          <div className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Known discount issued
          </div>
          <div className="mt-2 text-xl font-semibold text-slate-950">
            {formatMoney({
              amount: stats.totalDiscountAmount,
              currency: stats.totalDiscountCurrency,
            })}
          </div>
        </div>
      ) : null}
    </section>
  );
}
