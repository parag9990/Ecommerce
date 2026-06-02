import type { AnalyticsDateRange } from "../types";

export const analyticsQueryKeys = {
  summary: (sellerId: string, range: AnalyticsDateRange) =>
    ["seller-analytics", "summary", sellerId, range.from, range.to] as const,
};
