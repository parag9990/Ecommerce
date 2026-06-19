import type { CohortBucket } from "../../../api/session-api";
import {
  formatRetentionCount,
  formatRetentionRate
} from "../lib/retention-format";

type RetentionTooltipProps = {
  bucket?: CohortBucket;
  cohortLabel: string;
  cohortSize: number;
  offsetLabel: string;
};

export function RetentionTooltip({
  bucket,
  cohortLabel,
  cohortSize,
  offsetLabel
}: RetentionTooltipProps) {
  return (
    <div className="pointer-events-none absolute bottom-full left-1/2 z-20 mb-2 hidden w-56 -translate-x-1/2 rounded-md border border-zinc-200 bg-white p-3 text-left text-xs text-zinc-700 shadow-lg group-hover:block group-focus-within:block">
      <p className="font-semibold text-zinc-950">{cohortLabel}</p>
      <dl className="mt-2 grid grid-cols-2 gap-x-3 gap-y-1">
        <dt className="text-zinc-500">Period</dt>
        <dd className="text-right font-medium text-zinc-900">{offsetLabel}</dd>
        <dt className="text-zinc-500">Cohort size</dt>
        <dd className="text-right font-medium text-zinc-900">
          {formatRetentionCount(cohortSize)}
        </dd>
        <dt className="text-zinc-500">Retained</dt>
        <dd className="text-right font-medium text-zinc-900">
          {bucket?.suppressed ? "Masked" : formatRetentionCount(bucket?.users ?? 0)}
        </dd>
        <dt className="text-zinc-500">Rate</dt>
        <dd className="text-right font-medium text-zinc-900">
          {bucket?.suppressed ? "Small sample" : formatRetentionRate(bucket?.rate ?? 0)}
        </dd>
      </dl>
    </div>
  );
}
