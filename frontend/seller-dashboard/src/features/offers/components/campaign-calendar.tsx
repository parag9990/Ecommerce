import { CalendarDays, ChevronLeft, ChevronRight } from "lucide-react";

import { EmptyState } from "../../../components/state/empty-state";
import { cn } from "../../../lib/cn";
import type { Campaign } from "../types";
import {
  addMonths,
  dateFallsInsideRange,
  formatMonthLabel,
  getCalendarDays,
} from "../utils/offer-date-rules";
import { CampaignCard } from "./campaign-card";

type CampaignCalendarProps = {
  campaigns: Campaign[];
  month: Date;
  onMonthChange: (month: Date) => void;
};

const weekDays = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

function isToday(day: Date) {
  const today = new Date();

  return (
    day.getFullYear() === today.getFullYear() &&
    day.getMonth() === today.getMonth() &&
    day.getDate() === today.getDate()
  );
}

export function CampaignCalendar({
  campaigns,
  month,
  onMonthChange,
}: CampaignCalendarProps) {
  const days = getCalendarDays(month);

  return (
    <div className="overflow-hidden rounded-md border border-slate-200 bg-white shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-200 p-3">
        <div className="flex min-w-0 items-center gap-2">
          <CalendarDays className="h-4 w-4 text-slate-500" aria-hidden="true" />
          <h2 className="text-sm font-semibold text-slate-950">
            {formatMonthLabel(month)}
          </h2>
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            aria-label="Previous month"
            title="Previous month"
            onClick={() => onMonthChange(addMonths(month, -1))}
            className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          >
            <ChevronLeft className="h-4 w-4" aria-hidden="true" />
          </button>
          <button
            type="button"
            aria-label="Next month"
            title="Next month"
            onClick={() => onMonthChange(addMonths(month, 1))}
            className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          >
            <ChevronRight className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
      </div>

      <div className="grid grid-cols-7 border-b border-slate-200 bg-slate-50 text-center text-[11px] font-semibold uppercase tracking-wide text-slate-500">
        {weekDays.map((day) => (
          <div key={day} className="px-2 py-2">
            {day}
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-7">
        {days.map((day, index) => {
          const dayCampaigns = day
            ? campaigns.filter((campaign) =>
                dateFallsInsideRange(day, campaign.starts_at, campaign.ends_at),
              )
            : [];

          return (
            <div
              key={day?.toISOString() ?? `blank-${index}`}
              className={cn(
                "min-h-28 border-b border-slate-100 p-2 md:border-r",
                day ? "bg-white" : "hidden bg-slate-50 md:block",
              )}
            >
              {day ? (
                <>
                  <div
                    className={cn(
                      "mb-2 inline-flex h-6 min-w-6 items-center justify-center rounded text-xs font-semibold",
                      isToday(day) ? "bg-blue-600 px-1.5 text-white" : "text-slate-500",
                    )}
                  >
                    {day.getDate()}
                  </div>
                  <div className="space-y-1">
                    {dayCampaigns.slice(0, 3).map((campaign) => (
                      <CampaignCard key={campaign.campaign_id} campaign={campaign} />
                    ))}
                    {dayCampaigns.length > 3 ? (
                      <div className="rounded border border-slate-200 bg-slate-50 px-2 py-1 text-[11px] text-slate-500">
                        +{dayCampaigns.length - 3} more
                      </div>
                    ) : null}
                  </div>
                </>
              ) : null}
            </div>
          );
        })}
      </div>

      {campaigns.length === 0 ? (
        <div className="border-t border-slate-200 p-4">
          <EmptyState
            title="No campaigns scheduled"
            description="No campaigns scheduled for this month."
          />
        </div>
      ) : null}
    </div>
  );
}
