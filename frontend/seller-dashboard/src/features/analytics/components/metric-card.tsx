import type { LucideIcon } from "lucide-react";

import { cn } from "../../../lib/cn";

type MetricCardProps = {
  label: string;
  value: string;
  helper: string;
  icon: LucideIcon;
  unavailable?: boolean;
};

export function MetricCard({
  label,
  value,
  helper,
  icon: Icon,
  unavailable,
}: MetricCardProps) {
  return (
    <section className="rounded-md border border-slate-200 bg-white p-3 shadow-sm">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            {label}
          </p>
          <p
            className={cn(
              "mt-2 truncate text-2xl font-semibold",
              unavailable ? "text-slate-400" : "text-slate-950",
            )}
            title={value}
          >
            {value}
          </p>
        </div>
        <span className="rounded-md bg-slate-100 p-2 text-slate-500">
          <Icon className="h-4 w-4" aria-hidden="true" />
        </span>
      </div>
      <p className="mt-3 text-xs leading-5 text-slate-500">{helper}</p>
    </section>
  );
}
