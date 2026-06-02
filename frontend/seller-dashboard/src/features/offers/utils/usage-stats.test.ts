import { describe, expect, it } from "vitest";

import type { Campaign, Coupon } from "../types";
import { buildUsageStats } from "./usage-stats";

describe("buildUsageStats", () => {
  it("derives reliable offer counts without a dedicated usage endpoint", () => {
    const coupons: Coupon[] = [
      {
        coupon_id: "coupon-1",
        code: "SAVE10",
        discount_type: "percentage",
        discount_value: 10,
        status: "active",
        usage_limit: 100,
        used_count: 12,
      },
      {
        coupon_id: "coupon-2",
        code: "OLD",
        discount_type: "fixed",
        discount_value: 100,
        status: "expired",
        usage_limit: 10,
        used_count: 10,
      },
    ];
    const campaigns: Campaign[] = [
      {
        campaign_id: "campaign-1",
        name: "June Sale",
        starts_at: "2026-06-01T00:00:00.000Z",
        ends_at: "2099-06-30T23:59:59.000Z",
        status: "active",
      },
    ];

    expect(buildUsageStats(coupons, campaigns)).toMatchObject({
      totalCoupons: 2,
      activeCoupons: 1,
      expiredCoupons: 1,
      scheduledCampaigns: 1,
      knownUsedCount: 22,
      totalUsageLimit: 110,
    });
  });
});
