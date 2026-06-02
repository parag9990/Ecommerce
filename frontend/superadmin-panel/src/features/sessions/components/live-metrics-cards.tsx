import { Activity, Gauge, MousePointerClick } from "lucide-react";
import type { LucideIcon } from "lucide-react";

import type { LiveMetrics } from "../types";

type MetricCard = {
  label: string;
  value: number | undefined;
  icon: LucideIcon;
};

function formatMetric(value: number | undefined, isLoading: boolean): string {
  if (isLoading) {
    return "Loading";
  }

  return new Intl.NumberFormat("en").format(value ?? 0);
}

export function LiveMetricsCards({
  metrics,
  isLoading
}: {
  metrics?: LiveMetrics;
  isLoading: boolean;
}) {
  const cards: MetricCard[] = [
    {
      label: "Active Users",
      value: metrics?.active_users,
      icon: Activity
    },
    {
      label: "Active Sessions",
      value: metrics?.active_sessions,
      icon: Gauge
    },
    {
      label: "Events / Min",
      value: metrics?.events_per_minute,
      icon: MousePointerClick
    }
  ];

  return (
    <div className="grid gap-3 md:grid-cols-3">
      {cards.map((card) => {
        const Icon = card.icon;

        return (
          <div key={card.label} className="rounded-lg border border-slate-200 bg-white p-4">
            <div className="flex items-center justify-between gap-3">
              <p className="text-sm font-medium text-slate-600">{card.label}</p>
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                <Icon className="h-4 w-4" aria-hidden="true" />
              </span>
            </div>
            <p className="mt-3 text-2xl font-semibold text-slate-950">
              {formatMetric(card.value, isLoading)}
            </p>
          </div>
        );
      })}
    </div>
  );
}
