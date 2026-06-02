import { useQuery } from "@tanstack/react-query";

import { listAdminSessions } from "../api/sessions-api";
import type { AdminSessionFilters } from "../types";

export function useAdminSessions(filters: AdminSessionFilters) {
  return useQuery({
    queryKey: ["admin", "sessions", "list", filters],
    queryFn: () => listAdminSessions(filters),
    placeholderData: (previous) => previous
  });
}
