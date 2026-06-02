import { useQuery } from "@tanstack/react-query";

import { getLiveMetrics } from "../api/sessions-api";

export const LIVE_METRICS_REFETCH_MS = 15_000;

export function useLiveMetrics() {
  return useQuery({
    queryKey: ["admin", "sessions", "live"],
    queryFn: getLiveMetrics,
    refetchInterval: LIVE_METRICS_REFETCH_MS,
    staleTime: 10_000
  });
}
