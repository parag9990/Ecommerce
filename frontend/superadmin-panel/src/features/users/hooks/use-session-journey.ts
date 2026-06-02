import { useQuery } from "@tanstack/react-query";

import { getSessionJourney } from "../api/users-api";

export function useSessionJourney(sessionId: string | null) {
  return useQuery({
    queryKey: ["admin-session-journey", sessionId],
    queryFn: () => getSessionJourney(sessionId ?? ""),
    enabled: Boolean(sessionId),
    staleTime: 15_000
  });
}
