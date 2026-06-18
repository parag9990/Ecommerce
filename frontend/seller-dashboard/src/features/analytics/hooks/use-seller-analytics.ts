import { keepPreviousData, useQuery } from "@tanstack/react-query";

import { getSellerAnalyticsSummary } from "../api/seller-analytics-api";
import type { AnalyticsDateRange } from "../types";
import { normalizeSellerAnalytics } from "../utils/analytics-adapter";
import { analyticsQueryKeys } from "./query-keys";

export function useSellerAnalytics(range: AnalyticsDateRange, sellerId?: string) {
  return useQuery({
    queryKey: analyticsQueryKeys.summary(sellerId ?? "", range),
    queryFn: async () => {
      const response = await getSellerAnalyticsSummary(range);
      return normalizeSellerAnalytics(response);
    },
    enabled: Boolean(sellerId),
    placeholderData: keepPreviousData,
    staleTime: 60_000,
  });
}
