import type { SellerTeamFilters } from "../types";

export const teamQueryKeys = {
  all: ["seller-team"] as const,
  lists: () => [...teamQueryKeys.all, "list"] as const,
  list: (sellerId: string, filters: SellerTeamFilters) =>
    [...teamQueryKeys.lists(), sellerId, filters] as const,
};
