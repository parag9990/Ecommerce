import { useQuery } from "@tanstack/react-query";

import {
  getRetentionReport,
  type RetentionReportRequest
} from "../../../api/session-api";

type UseRetentionReportOptions = {
  enabled?: boolean;
};

export function useRetentionReport(
  request: RetentionReportRequest,
  options: UseRetentionReportOptions = {}
) {
  return useQuery({
    enabled: options.enabled ?? true,
    queryFn: ({ signal }) => getRetentionReport(request, { signal }),
    queryKey: [
      "analytics",
      "retention-report",
      request.dateRange.from,
      request.dateRange.to,
      request.filters.deviceType,
      request.filters.channel,
      request.filters.source,
      request.filters.userType,
      request.interval,
      request.window
    ],
    refetchOnWindowFocus: false,
    staleTime: 60 * 1000
  });
}
