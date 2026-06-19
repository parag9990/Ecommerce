import { AlertTriangle, ShieldCheck, TrendingDown } from "lucide-react";

import {
  funnelStepLabels,
  formatFunnelCount,
  formatFunnelPercent
} from "../lib/funnel-format";
import type { FunnelReport } from "../lib/funnel-math";

type FunnelInsightPanelProps = {
  report: FunnelReport;
};

export function FunnelInsightPanel({ report }: FunnelInsightPanelProps) {
  const biggestDrop = report.steps.find(
    (step) => step.key === report.biggestDropoffStepKey
  );

  return (
    <aside className="space-y-4">
      <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
        <div className="flex items-center gap-2">
          <TrendingDown className="h-4 w-4 text-amber-600" aria-hidden="true" />
          <h2 className="text-sm font-semibold text-zinc-950">Readout</h2>
        </div>
        <div className="mt-3 space-y-3 text-sm text-zinc-600">
          <p>
            Paid conversion is{" "}
            <span className="font-semibold text-zinc-950">
              {formatFunnelPercent(report.overallConversionRate)}
            </span>{" "}
            across {formatFunnelCount(report.totalStarted)} started sessions.
          </p>
          {biggestDrop ? (
            <p>
              The largest loss is at{" "}
              <span className="font-semibold text-red-700">
                {funnelStepLabels[biggestDrop.key]}
              </span>{" "}
              with {formatFunnelPercent(biggestDrop.dropoffRateFromPrevious)} drop-off.
            </p>
          ) : (
            <p>No material drop-off is visible in this aggregate view.</p>
          )}
        </div>
      </section>

      <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
        <div className="flex items-center gap-2">
          <ShieldCheck className="h-4 w-4 text-emerald-700" aria-hidden="true" />
          <h2 className="text-sm font-semibold text-zinc-950">Privacy</h2>
        </div>
        <p className="mt-3 text-sm text-zinc-600">
          This view only renders aggregate step counts and masks small buckets
          below {report.minSegmentSize} sessions.
        </p>
      </section>

      {report.partial || report.suppressed ? (
        <section className="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
          <div className="flex gap-2">
            <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
            <p>
              {report.suppressed
                ? "Some low-count segments were suppressed."
                : "This response is marked as partial."}
            </p>
          </div>
        </section>
      ) : null}
    </aside>
  );
}
