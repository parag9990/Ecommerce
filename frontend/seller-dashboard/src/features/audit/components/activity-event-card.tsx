import { Clock, FileClock } from "lucide-react";

import { cn } from "../../../lib/cn";
import type { SellerAuditLog } from "../types";
import { formatAuditTime, formatRelativeAuditTime } from "../utils/audit-formatters";
import { getActionLabel, getResourceLabel } from "../utils/audit-labels";
import { ActorSummary } from "./actor-summary";
import { DiffSummary } from "./diff-summary";
import { ResourceLink } from "./resource-link";

type ActivityEventCardProps = {
  log: SellerAuditLog;
};

const resourceBadgeStyles: Record<string, string> = {
  product: "border-blue-200 bg-blue-50 text-blue-700",
  variant: "border-cyan-200 bg-cyan-50 text-cyan-700",
  order: "border-emerald-200 bg-emerald-50 text-emerald-700",
  coupon: "border-amber-200 bg-amber-50 text-amber-700",
  campaign: "border-fuchsia-200 bg-fuchsia-50 text-fuchsia-700",
  team: "border-violet-200 bg-violet-50 text-violet-700",
  settings: "border-slate-200 bg-slate-50 text-slate-700",
};

export function ActivityEventCard({ log }: ActivityEventCardProps) {
  return (
    <article className="relative ml-9 rounded-md border border-slate-200 bg-white p-4 shadow-sm">
      <div className="absolute -left-[1.88rem] top-4 flex h-7 w-7 items-center justify-center rounded-full border border-slate-200 bg-white text-blue-700 shadow-sm">
        <FileClock className="h-3.5 w-3.5" aria-hidden="true" />
      </div>

      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span
              className={cn(
                "inline-flex h-6 items-center rounded border px-2 text-xs font-medium",
                resourceBadgeStyles[log.resource_type] ??
                  "border-slate-200 bg-white text-slate-700",
              )}
            >
              {getResourceLabel(log.resource_type)}
            </span>
            <h2 className="text-sm font-semibold text-slate-950">
              {getActionLabel(log.action)}
            </h2>
          </div>

          <div className="mt-2 flex flex-wrap items-center gap-2 text-sm text-slate-600">
            <ActorSummary log={log} />
            <span>on</span>
            <ResourceLink log={log} />
          </div>
        </div>

        <time
          className="flex shrink-0 items-center gap-1 text-xs text-slate-500"
          dateTime={log.created_at}
          title={formatAuditTime(log.created_at)}
        >
          <Clock className="h-3.5 w-3.5" aria-hidden="true" />
          {formatRelativeAuditTime(log.created_at)}
        </time>
      </div>

      <DiffSummary before={log.before} after={log.after} />
    </article>
  );
}
