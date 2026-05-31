import type {
  RetentionCohort,
  RetentionInterval
} from "../../../api/session-api";
import { getRetentionCellClass } from "../lib/retention-colors";
import {
  formatOffsetLabel,
  formatRetentionCount,
  formatRetentionRate
} from "../lib/retention-format";
import { findBucket } from "../lib/retention-math";
import { RetentionTooltip } from "./retention-tooltip";

type CohortRetentionMatrixProps = {
  cohorts: RetentionCohort[];
  interval: RetentionInterval;
  window: number;
};

export function CohortRetentionMatrix({
  cohorts,
  interval,
  window
}: CohortRetentionMatrixProps) {
  const offsets = Array.from({ length: window }, (_, index) => index);

  return (
    <section className="rounded-lg border border-zinc-200 bg-white shadow-panel">
      <div className="flex flex-col gap-3 border-b border-zinc-200 p-4 md:flex-row md:items-start md:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase text-emerald-700">
            Cohort matrix
          </p>
          <h2 className="mt-1 text-base font-semibold text-zinc-950">
            Cohort retention
          </h2>
        </div>
      </div>

      <div className="overflow-x-auto p-4">
        <table className="min-w-[820px] border-separate border-spacing-1">
          <thead>
            <tr>
              <th className="sticky left-0 z-10 w-48 bg-white px-2 py-2 text-left text-xs font-semibold uppercase text-zinc-500">
                Cohort
              </th>
              <th className="w-28 px-2 py-2 text-right text-xs font-semibold uppercase text-zinc-500">
                Users
              </th>
              {offsets.map((offset) => (
                <th
                  className="w-24 px-2 py-2 text-center text-xs font-semibold uppercase text-zinc-500"
                  key={offset}
                >
                  {formatOffsetLabel(interval, offset)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {cohorts.map((cohort) => (
              <tr key={cohort.cohortKey}>
                <th className="sticky left-0 z-10 bg-white px-2 py-2 text-left text-sm font-medium text-zinc-800">
                  <span className="block max-w-44 truncate">
                    {cohort.cohortLabel}
                  </span>
                </th>
                <td className="px-2 py-2 text-right text-sm text-zinc-600">
                  {formatRetentionCount(cohort.cohortSize)}
                </td>
                {offsets.map((offset) => {
                  const bucket = findBucket(cohort, offset);
                  const label = formatOffsetLabel(interval, offset);
                  const cellText = bucket?.suppressed
                    ? "Masked"
                    : bucket
                      ? formatRetentionRate(bucket.rate)
                      : "-";

                  return (
                    <td className="p-0.5" key={offset}>
                      <div
                        aria-label={`${cohort.cohortLabel} ${label} ${cellText}`}
                        className={[
                          "group relative flex h-12 min-w-24 items-center justify-center rounded-md border px-2 text-sm font-semibold outline-none",
                          getRetentionCellClass(
                            bucket?.rate ?? 0,
                            bucket?.suppressed ?? cohort.suppressed
                          )
                        ].join(" ")}
                        tabIndex={0}
                        title={`${cohort.cohortLabel}, ${label}: ${cellText}`}
                      >
                        {cellText}
                        <RetentionTooltip
                          bucket={bucket}
                          cohortLabel={cohort.cohortLabel}
                          cohortSize={cohort.cohortSize}
                          offsetLabel={label}
                        />
                      </div>
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
