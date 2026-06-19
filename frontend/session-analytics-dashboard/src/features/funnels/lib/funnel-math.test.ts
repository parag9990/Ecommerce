import { describe, expect, it } from "vitest";

import { normalizeFunnelReport } from "./funnel-math";

describe("normalizeFunnelReport", () => {
  it("calculates conversion and drop-off for the default ecommerce funnel", () => {
    const report = normalizeFunnelReport({
      steps: [
        { key: "product_view", count: 100 },
        { key: "add_to_cart", count: 40 },
        { key: "checkout_started", count: 20 },
        { key: "paid", count: 10 }
      ]
    });

    expect(report.totalStarted).toBe(100);
    expect(report.totalCompleted).toBe(10);
    expect(report.overallConversionRate).toBe(10);
    expect(report.biggestDropoffStepKey).toBe("add_to_cart");
    expect(report.steps[1]).toMatchObject({
      conversionFromPrevious: 40,
      conversionFromStart: 40,
      dropoffFromPrevious: 60,
      dropoffRateFromPrevious: 60
    });
  });

  it("normalizes flexible backend fields and missing steps", () => {
    const report = normalizeFunnelReport({
      minSegmentSize: 5,
      partial: true,
      steps: [
        { step: "Product Viewed", sessions: 12.4 },
        { step: "checkout_step", uniqueSessions: 6 },
        { step: "payment_result", count: 3 }
      ]
    });

    expect(report.partial).toBe(true);
    expect(report.steps.map((step) => [step.key, step.count])).toEqual([
      ["product_view", 12],
      ["add_to_cart", 0],
      ["checkout_started", 6],
      ["paid", 3]
    ]);
    expect(report.hasSmallCounts).toBe(true);
    expect(report.steps[1].dropoffFromPrevious).toBe(12);
  });

  it("avoids divide-by-zero when the first step is empty", () => {
    const report = normalizeFunnelReport({
      steps: [
        { key: "add_to_cart", count: 4 },
        { key: "paid", count: 2 }
      ]
    });

    expect(report.totalStarted).toBe(0);
    expect(report.overallConversionRate).toBe(0);
    expect(report.steps[1].conversionFromPrevious).toBe(0);
  });
});
