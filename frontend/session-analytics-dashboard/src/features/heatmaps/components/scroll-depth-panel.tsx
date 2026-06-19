import { ScrollText } from "lucide-react";

import type { HeatmapPoint } from "../../../api/session-api";
import { formatNumber } from "../../../lib/format";
import { getScrollDepthBuckets } from "../lib/heatmap-stats";

type ScrollDepthPanelProps = {
  points: HeatmapPoint[];
};

export function ScrollDepthPanel({ points }: ScrollDepthPanelProps) {
  const buckets = getScrollDepthBuckets(points);

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase text-zinc-500">
            Scroll reach
          </p>
          <h2 className="mt-1 text-sm font-semibold text-zinc-950">
            Depth thresholds
          </h2>
        </div>
        <span className="rounded-md bg-sky-50 p-2 text-sky-700">
          <ScrollText className="h-4 w-4" aria-hidden="true" />
        </span>
      </div>

      <div className="mt-4 space-y-3">
        {buckets.map((bucket) => (
          <div className="grid gap-1.5" key={bucket.depth}>
            <div className="flex items-center justify-between gap-3 text-xs">
              <span className="font-medium text-zinc-700">
                {bucket.depth}% depth
              </span>
              <span className="text-zinc-500">
                {bucket.percentage}% | {formatNumber(bucket.events)}
              </span>
            </div>
            <div className="h-2 overflow-hidden rounded-full bg-zinc-100">
              <div
                aria-hidden="true"
                className="h-full rounded-full bg-emerald-500"
                style={{ width: `${bucket.percentage}%` }}
              />
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
