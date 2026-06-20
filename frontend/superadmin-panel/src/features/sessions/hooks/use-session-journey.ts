import { useQuery } from "@tanstack/react-query";

import { getAdminSessionJourney } from "../api/sessions-api";

export function useSessionJourney(sessionId?: string | null) {
  return useQuery({
    queryKey: ["admin", "sessions", "journey", sessionId],
    queryFn: () => getAdminSessionJourney(sessionId!),
    enabled: Boolean(sessionId)
  });
}
