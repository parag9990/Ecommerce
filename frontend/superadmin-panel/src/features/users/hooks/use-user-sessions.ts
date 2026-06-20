import { useQuery } from "@tanstack/react-query";

import { listUserSessions } from "../api/users-api";

export function useUserSessions(userId: string) {
  return useQuery({
    queryKey: ["admin-user-sessions", userId],
    queryFn: () => listUserSessions(userId),
    enabled: Boolean(userId),
    staleTime: 15_000
  });
}
