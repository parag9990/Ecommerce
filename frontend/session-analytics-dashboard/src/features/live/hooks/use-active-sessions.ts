import { useQuery } from "@tanstack/react-query";

import {
  getActiveSessions,
  type ActiveSessionsRequest
} from "../../../api/session-api";

type UseActiveSessionsOptions = {
  autoRefresh: boolean;
  enabled?: boolean;
  request: ActiveSessionsRequest;
};

export function useActiveSessions({
  autoRefresh,
  enabled = true,
  request
}: UseActiveSessionsOptions) {
  return useQuery({
    enabled,
    queryFn: ({ signal }) => getActiveSessions(request, { signal }),
    queryKey: [
      "analytics",
      "active-sessions",
      request.from ?? "",
      request.to ?? "",
      request.q ?? "",
      request.deviceType ?? "all",
      request.country ?? "",
      request.entryPage ?? "",
      request.limit ?? 50
    ],
    refetchInterval: autoRefresh ? 15 * 1000 : false,
    staleTime: 10 * 1000
  });
}
