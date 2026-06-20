import { useQuery } from "@tanstack/react-query";

import { getAdminUser } from "../api/users-api";

export function useAdminUser(userId: string) {
  return useQuery({
    queryKey: ["admin-user", userId],
    queryFn: () => getAdminUser(userId),
    enabled: Boolean(userId),
    staleTime: 30_000
  });
}
