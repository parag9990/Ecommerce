import { useQuery } from "@tanstack/react-query";

import { listSellerKycDocuments } from "../api/sellers-api";

export function useSellerKycDocuments(sellerId: string, enabled = true) {
  return useQuery({
    queryKey: ["seller-kyc-documents", sellerId],
    queryFn: () => listSellerKycDocuments(sellerId),
    enabled: Boolean(sellerId) && enabled,
    staleTime: 30_000
  });
}
