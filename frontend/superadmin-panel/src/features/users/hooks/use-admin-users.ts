import { useQuery } from "@tanstack/react-query";

import { listAdminUsers } from "../api/users-api";
import type { AdminUserFilters } from "../types";

export function useAdminUsers(filters: AdminUserFilters) {
  return useQuery({
    queryKey: ["admin-users", filters],
    queryFn: () => listAdminUsers(filters),
    staleTime: 30_000
  });
}
