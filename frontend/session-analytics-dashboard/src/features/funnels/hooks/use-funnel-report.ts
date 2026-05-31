import { useQuery } from "@tanstack/react-query";

import {
  getFunnelReport,
  type FunnelReportRequest
} from "../../../api/session-api";
import { normalizeFunnelReport } from "../lib/funnel-math";

type UseFunnelReportOptions = {
  enabled?: boolean;
};

export function useFunnelReport(
  request: FunnelReportRequest,
  options: UseFunnelReportOptions = {}
) {
  return useQuery({
    enabled: options.enabled ?? true,
    queryFn: ({ signal }) =>
      getFunnelReport(request, { signal }).then(normalizeFunnelReport),
    queryKey: [
      "analytics",
      "funnel-report",
      request.dateRange.from,
      request.dateRange.to,
      request.filters.deviceType,
      request.filters.channel,
      request.filters.source,
      request.filters.userType,
      request.steps?.join(",") ?? ""
    ],
    refetchOnWindowFocus: false,
    staleTime: 60 * 1000
  });
}
