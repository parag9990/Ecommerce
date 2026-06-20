import { useQuery } from "@tanstack/react-query";

import { getAdminSeller } from "../api/sellers-api";

export function useAdminSeller(sellerId: string) {
  return useQuery({
    queryKey: ["admin-seller", sellerId],
    queryFn: () => getAdminSeller(sellerId),
    enabled: Boolean(sellerId),
    staleTime: 30_000
  });
}
