import { funnelStepColors } from "../../../lib/chart-theme";
import { formatFunnelCount, formatFunnelPercent } from "../lib/funnel-format";
import type { FunnelStep } from "../lib/funnel-math";

type FunnelStepCardProps = {
  index: number;
  minSegmentSize: number;
  step: FunnelStep;
};

export function FunnelStepCard({
  index,
  minSegmentSize,
  step
}: FunnelStepCardProps) {
  const color = funnelStepColors[index % funnelStepColors.length];

  return (
    <article className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase text-zinc-500">
            Step {index + 1}
          </p>
          <h3 className="mt-1 text-sm font-semibold text-zinc-950">
            {step.label}
          </h3>
        </div>
        <span
          aria-hidden="true"
          className="h-3 w-3 rounded-full"
          style={{ backgroundColor: color }}
        />
      </div>

      <p className="mt-4 text-2xl font-semibold text-zinc-950">
        {formatFunnelCount(step.count, { threshold: minSegmentSize })}
      </p>

      <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
        <div>
          <dt className="text-xs font-medium uppercase text-zinc-500">
            From previous
          </dt>
          <dd className="mt-1 font-semibold text-zinc-900">
            {formatFunnelPercent(step.conversionFromPrevious)}
          </dd>
        </div>
        <div>
          <dt className="text-xs font-medium uppercase text-zinc-500">
            Drop-off
          </dt>
          <dd className="mt-1 font-semibold text-red-700">
            {formatFunnelPercent(step.dropoffRateFromPrevious)}
          </dd>
        </div>
      </dl>
    </article>
  );
}
