import { useQuery } from "@tanstack/react-query";

import { getLiveMetrics, type LiveMetricsRequest } from "../../../api/session-api";

type UseLiveMetricsOptions = {
  enabled?: boolean;
};

export function useLiveMetrics(
  request: LiveMetricsRequest,
  options: UseLiveMetricsOptions = {}
) {
  return useQuery({
    enabled: options.enabled ?? true,
    queryFn: ({ signal }) => getLiveMetrics(request, { signal }),
    queryKey: [
      "analytics",
      "live-metrics",
      request.dateRange.from,
      request.dateRange.to,
      request.filters.deviceType,
      request.filters.channel,
      request.filters.source,
      request.filters.userType
    ]
  });
}
