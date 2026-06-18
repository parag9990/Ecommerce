import { useMutation, useQueryClient } from "@tanstack/react-query";

import { createSellerCampaign } from "../api/seller-offers-api";
import type { CampaignInput } from "../types";
import { offerQueryKeys } from "./query-keys";

export function useCreateCampaign() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CampaignInput) => createSellerCampaign(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: offerQueryKeys.campaigns() });
    },
  });
}
