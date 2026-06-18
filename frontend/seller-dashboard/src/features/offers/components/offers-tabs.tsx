import { CalendarDays, TicketPercent } from "lucide-react";

import { cn } from "../../../lib/cn";
import type { OffersTab } from "../types";

type OffersTabsProps = {
  activeTab: OffersTab;
  onChange: (tab: OffersTab) => void;
};

const tabs: Array<{ value: OffersTab; label: string; icon: typeof TicketPercent }> = [
  { value: "coupons", label: "Coupons", icon: TicketPercent },
  { value: "campaigns", label: "Campaigns", icon: CalendarDays },
];

export function OffersTabs({ activeTab, onChange }: OffersTabsProps) {
  return (
    <div className="inline-flex rounded-md border border-slate-200 bg-white p-1 shadow-sm">
      {tabs.map((tab) => {
        const Icon = tab.icon;

        return (
          <button
            key={tab.value}
            type="button"
            onClick={() => onChange(tab.value)}
            className={cn(
              "inline-flex h-8 items-center gap-2 rounded px-3 text-sm font-medium transition",
              "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600",
              activeTab === tab.value
                ? "bg-blue-600 text-white"
                : "text-slate-600 hover:bg-slate-50 hover:text-slate-950",
            )}
          >
            <Icon className="h-4 w-4" aria-hidden="true" />
            {tab.label}
          </button>
        );
      })}
    </div>
  );
}
