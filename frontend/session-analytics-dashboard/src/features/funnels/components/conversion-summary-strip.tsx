import {
  AlertTriangle,
  CheckCircle2,
  MousePointerClick,
  TrendingDown,
  type LucideIcon
} from "lucide-react";

import {
  funnelStepLabels,
  formatFunnelCount,
  formatFunnelPercent
} from "../lib/funnel-format";
import type { FunnelReport } from "../lib/funnel-math";

type ConversionSummaryStripProps = {
  report: FunnelReport;
};

export function ConversionSummaryStrip({ report }: ConversionSummaryStripProps) {
  const biggestDropLabel = report.biggestDropoffStepKey
    ? funnelStepLabels[report.biggestDropoffStepKey]
    : "No drop-off";

  return (
    <section
      aria-label="Funnel summary"
      className="grid gap-4 md:grid-cols-2 xl:grid-cols-4"
    >
      <SummaryMetric
        icon={MousePointerClick}
        label="Started funnel"
        tone="blue"
        value={formatFunnelCount(report.totalStarted)}
      />
      <SummaryMetric
        icon={CheckCircle2}
        label="Paid sessions"
        tone="green"
        value={formatFunnelCount(report.totalCompleted)}
      />
      <SummaryMetric
        icon={TrendingDown}
        label="Paid conversion"
        tone="amber"
        value={formatFunnelPercent(report.overallConversionRate)}
      />
      <SummaryMetric
        icon={AlertTriangle}
        label="Biggest drop"
        tone="red"
        value={biggestDropLabel}
      />
    </section>
  );
}

type SummaryMetricProps = {
  icon: LucideIcon;
  label: string;
  tone: "amber" | "blue" | "green" | "red";
  value: string;
};

const toneClasses: Record<SummaryMetricProps["tone"], string> = {
  amber: "bg-amber-50 text-amber-700",
  blue: "bg-blue-50 text-blue-700",
  green: "bg-emerald-50 text-emerald-700",
  red: "bg-red-50 text-red-700"
};

function SummaryMetric({ icon: Icon, label, tone, value }: SummaryMetricProps) {
  return (
    <article className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-xs font-medium uppercase text-zinc-500">{label}</p>
          <p className="mt-2 break-words text-2xl font-semibold text-zinc-950">
            {value}
          </p>
        </div>
        <span className={`rounded-md p-2 ${toneClasses[tone]}`}>
          <Icon className="h-4 w-4" aria-hidden="true" />
        </span>
      </div>
    </article>
  );
}
