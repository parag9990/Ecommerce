import {
  BarChart3,
  Flame,
  Gauge,
  MousePointerClick,
  type LucideIcon
} from "lucide-react";

import type {
  HeatmapMode,
  HeatmapPoint,
  HeatmapResponse
} from "../../../api/session-api";
import { formatNumber, formatPercent } from "../../../lib/format";
import { getHeatmapStats } from "../lib/heatmap-stats";

type HeatmapSummaryStripProps = {
  mode: HeatmapMode;
  points: HeatmapPoint[];
  response?: HeatmapResponse;
};

export function HeatmapSummaryStrip({
  mode,
  points,
  response
}: HeatmapSummaryStripProps) {
  const stats = getHeatmapStats(points, response);
  const hotspotValue = stats.hottestPoint
    ? `${stats.hottestPoint.x.toFixed(0)}%, ${stats.hottestPoint.y.toFixed(0)}%`
    : "No data";

  return (
    <section
      aria-label="Heatmap summary"
      className="grid gap-4 md:grid-cols-2 xl:grid-cols-4"
    >
      <SummaryMetric
        icon={MousePointerClick}
        label="Total events"
        tone="blue"
        value={formatNumber(stats.totalEvents)}
      />
      <SummaryMetric
        icon={BarChart3}
        label="Heat buckets"
        tone="green"
        value={formatNumber(stats.pointCount)}
      />
      <SummaryMetric
        icon={Flame}
        label="Peak bucket"
        tone="amber"
        value={formatNumber(stats.maxWeight)}
      />
      <SummaryMetric
        icon={mode === "scroll" ? Gauge : MousePointerClick}
        label={mode === "scroll" ? "Avg depth" : "Hotspot"}
        tone="red"
        value={
          mode === "scroll"
            ? formatPercent(stats.averageScrollDepth)
            : hotspotValue
        }
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
  blue: "bg-sky-50 text-sky-700",
  green: "bg-emerald-50 text-emerald-700",
  red: "bg-red-50 text-red-700"
};

function SummaryMetric({ icon: Icon, label, tone, value }: SummaryMetricProps) {
  return (
    <article className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
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
