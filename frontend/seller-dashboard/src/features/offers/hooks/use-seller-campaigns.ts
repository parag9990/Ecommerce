import { keepPreviousData, useQuery } from "@tanstack/react-query";

import { listSellerCampaigns } from "../api/seller-offers-api";
import type { CampaignFilters } from "../types";
import { offerQueryKeys } from "./query-keys";

export function useSellerCampaigns(filters: CampaignFilters, sellerId?: string) {
  return useQuery({
    queryKey: offerQueryKeys.campaignList(sellerId ?? "", filters),
    queryFn: () => listSellerCampaigns(filters),
    enabled: Boolean(sellerId),
    placeholderData: keepPreviousData,
  });
}
