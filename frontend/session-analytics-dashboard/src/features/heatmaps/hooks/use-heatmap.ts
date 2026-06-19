import { useQuery } from "@tanstack/react-query";

import { getHeatmap, type HeatmapRequest } from "../../../api/session-api";

type UseHeatmapOptions = {
  enabled?: boolean;
};

export function useHeatmap(
  request: HeatmapRequest,
  options: UseHeatmapOptions = {}
) {
  return useQuery({
    enabled:
      (options.enabled ?? true) &&
      Boolean(request.path && request.deviceType && request.from && request.to),
    queryFn: ({ signal }) => getHeatmap(request, { signal }),
    queryKey: [
      "analytics",
      "heatmap",
      request.path,
      request.deviceType,
      request.mode ?? "click",
      request.from,
      request.to
    ],
    refetchOnWindowFocus: false,
    staleTime: 60 * 1000
  });
}
