import { Percent, TrendingUp, UserPlus, Users } from "lucide-react";

import type { RetentionSummary } from "../../../api/session-api";
import { getRetentionSummaryTone } from "../lib/retention-colors";
import {
  formatRetentionCount,
  formatRetentionRate
} from "../lib/retention-format";

type CohortSummaryCardsProps = {
  summary: RetentionSummary;
};

const metrics = [
  {
    icon: UserPlus,
    key: "newUsers",
    label: "New users",
    value: (summary: RetentionSummary) =>
      formatRetentionCount(summary.newUsers)
  },
  {
    icon: Users,
    key: "returningUsers",
    label: "Returning users",
    value: (summary: RetentionSummary) =>
      formatRetentionCount(summary.returningUsers)
  },
  {
    icon: Percent,
    key: "returningRate",
    label: "Returning rate",
    value: (summary: RetentionSummary) =>
      formatRetentionRate(summary.returningRate)
  },
  {
    icon: TrendingUp,
    key: "averageRetention",
    label: "Avg retention",
    value: (summary: RetentionSummary) =>
      formatRetentionRate(summary.averageRetention)
  }
] as const;

export function CohortSummaryCards({ summary }: CohortSummaryCardsProps) {
  return (
    <section
      aria-label="Retention summary"
      className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4"
    >
      {metrics.map((metric, index) => {
        const Icon = metric.icon;

        return (
          <article
            className={[
              "rounded-lg border p-4 shadow-panel",
              getRetentionSummaryTone(index)
            ].join(" ")}
            key={metric.key}
          >
            <div className="flex items-center justify-between gap-3">
              <p className="text-sm font-medium opacity-80">{metric.label}</p>
              <Icon className="h-4 w-4 shrink-0 opacity-70" aria-hidden="true" />
            </div>
            <p className="mt-3 text-2xl font-semibold">
              {metric.value(summary)}
            </p>
          </article>
        );
      })}
    </section>
  );
}
