import { http } from "../../../lib/http";
import type { AnalyticsDateRange, SellerAnalyticsResponse } from "../types";

function toQuery(range: AnalyticsDateRange) {
  const search = new URLSearchParams();
  search.set("from", range.from);
  search.set("to", range.to);
  return search.toString();
}

export async function getSellerAnalyticsSummary(
  range: AnalyticsDateRange,
): Promise<SellerAnalyticsResponse> {
  return http<SellerAnalyticsResponse>(
    `/api/v1/seller/dashboard/summary?${toQuery(range)}`,
  );
}
