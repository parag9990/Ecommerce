import type {
  FunnelStepKey,
  RawFunnelReportResponse,
  RawFunnelStep
} from "../../../api/session-api";
import {
  defaultFunnelSteps,
  funnelStepLabels,
  smallCountThreshold
} from "./funnel-format";

export type FunnelStep = {
  key: FunnelStepKey;
  label: string;
  count: number;
  conversionFromPrevious: number | null;
  conversionFromStart: number;
  dropoffFromPrevious: number | null;
  dropoffRateFromPrevious: number | null;
};

export type FunnelReport = {
  steps: FunnelStep[];
  totalStarted: number;
  totalCompleted: number;
  overallConversionRate: number;
  biggestDropoffStepKey: FunnelStepKey | null;
  generatedAt?: string;
  hasSmallCounts: boolean;
  minSegmentSize: number;
  partial: boolean;
  suppressed: boolean;
};

export function normalizeFunnelReport(
  response: RawFunnelReportResponse
): FunnelReport {
  const minSegmentSize = Math.max(
    smallCountThreshold,
    Math.round(response.minSegmentSize ?? smallCountThreshold)
  );
  const countByStep = new Map<FunnelStepKey, number>();

  for (const item of response.steps ?? []) {
    const key = readStepKey(item);
    if (!key) {
      continue;
    }

    const count = sanitizeCount(readStepCount(item));
    const current = countByStep.get(key) ?? 0;
    countByStep.set(key, Math.max(current, count));
  }

  const firstCount = countByStep.get(defaultFunnelSteps[0]) ?? 0;

  const steps = defaultFunnelSteps.map((key, index) => {
    const count = countByStep.get(key) ?? 0;
    const previousKey = defaultFunnelSteps[index - 1];
    const previousCount = previousKey ? countByStep.get(previousKey) ?? 0 : null;
    const dropoff =
      previousCount === null ? null : Math.max(0, previousCount - count);

    return {
      count,
      conversionFromPrevious:
        previousCount === null ? null : safePercent(count, previousCount),
      conversionFromStart: safePercent(count, firstCount),
      dropoffFromPrevious: dropoff,
      dropoffRateFromPrevious:
        previousCount === null || dropoff === null
          ? null
          : safePercent(dropoff, previousCount),
      key,
      label: funnelStepLabels[key]
    };
  });

  const biggestDrop = steps.reduce<FunnelStep | null>((winner, step) => {
    if ((step.dropoffFromPrevious ?? 0) <= 0) {
      return winner;
    }

    if (!winner || (step.dropoffFromPrevious ?? 0) > (winner.dropoffFromPrevious ?? 0)) {
      return step;
    }

    return winner;
  }, null);
  const totalCompleted = steps[steps.length - 1]?.count ?? 0;

  return {
    biggestDropoffStepKey: biggestDrop?.key ?? null,
    generatedAt: response.generatedAt,
    hasSmallCounts: steps.some(
      (step) => step.count > 0 && step.count < minSegmentSize
    ),
    minSegmentSize,
    overallConversionRate: safePercent(totalCompleted, firstCount),
    partial: response.partial ?? false,
    steps,
    suppressed: response.suppressed ?? false,
    totalCompleted,
    totalStarted: firstCount
  };
}

function readStepKey(item: RawFunnelStep): FunnelStepKey | null {
  const rawKey = normalizeToken(item.key ?? item.step ?? item.label);

  if (rawKey === "product_view" || rawKey === "product_viewed") {
    return "product_view";
  }

  if (rawKey === "add_to_cart" || rawKey === "added_to_cart") {
    return "add_to_cart";
  }

  if (
    rawKey === "checkout_started" ||
    rawKey === "checkout_start" ||
    rawKey === "checkout_step"
  ) {
    return "checkout_step";
  }

  if (
    rawKey === "paid" ||
    rawKey === "payment_result" ||
    rawKey === "payment_success" ||
    rawKey === "payment_succeeded"
  ) {
    return "payment_result";
  }

  return null;
}

function readStepCount(item: RawFunnelStep): number {
  return (
    item.count ??
    item.sessions ??
    item.uniqueSessions ??
    item.users ??
    item.uniqueUsers ??
    0
  );
}

function sanitizeCount(value: number): number {
  if (!Number.isFinite(value)) {
    return 0;
  }

  return Math.max(0, Math.round(value));
}

function safePercent(part: number, total: number): number {
  if (total <= 0) {
    return 0;
  }

  return (part / total) * 100;
}

function normalizeToken(value: string | undefined): string {
  return value?.trim().toLowerCase().replace(/[\s-]+/g, "_") ?? "";
}
