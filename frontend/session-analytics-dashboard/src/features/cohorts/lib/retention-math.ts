import type { CohortBucket, RetentionCohort } from "../../../api/session-api";

export function safePercent(value: number, total: number): number {
  if (!Number.isFinite(value) || !Number.isFinite(total) || total <= 0) {
    return 0;
  }

  return Number(((value / total) * 100).toFixed(2));
}

export function getAverageRetentionForOffset(
  cohorts: RetentionCohort[],
  offset: number
): number {
  const rates = cohorts
    .map((cohort) => findBucket(cohort, offset))
    .filter((bucket): bucket is CohortBucket => Boolean(bucket))
    .filter((bucket) => !bucket.suppressed)
    .map((bucket) => bucket.rate)
    .filter(Number.isFinite);

  if (rates.length === 0) {
    return 0;
  }

  return Number(
    (rates.reduce((sum, rate) => sum + rate, 0) / rates.length).toFixed(2)
  );
}

export function hasRetentionData(cohorts: RetentionCohort[]): boolean {
  return cohorts.some(
    (cohort) =>
      cohort.cohortSize > 0 ||
      cohort.buckets.some((bucket) => bucket.users > 0 || bucket.rate > 0)
  );
}

export function hasSuppressedBuckets(cohorts: RetentionCohort[]): boolean {
  return cohorts.some(
    (cohort) => cohort.suppressed || cohort.buckets.some((bucket) => bucket.suppressed)
  );
}

export function findBucket(
  cohort: RetentionCohort,
  offset: number
): CohortBucket | undefined {
  return cohort.buckets.find((bucket) => bucket.offset === offset);
}
