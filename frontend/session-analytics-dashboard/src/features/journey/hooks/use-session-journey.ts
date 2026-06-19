import { useQuery } from "@tanstack/react-query";

import { getSessionJourney } from "../../../api/session-api";

export function useSessionJourney(sessionId?: string) {
  const normalizedSessionId = sessionId?.trim();

  return useQuery({
    enabled: Boolean(normalizedSessionId),
    queryFn: ({ signal }) => getSessionJourney(normalizedSessionId ?? "", { signal }),
    queryKey: ["analytics", "session-journey", normalizedSessionId ?? ""],
    retry: 1,
    staleTime: 20 * 1000
  });
}
