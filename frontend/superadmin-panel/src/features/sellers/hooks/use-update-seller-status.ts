import { useMutation, useQueryClient } from "@tanstack/react-query";

import { updateSellerStatus } from "../api/sellers-api";

export function useUpdateSellerStatus() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: updateSellerStatus,
    onSuccess: (_data, input) => {
      queryClient.invalidateQueries({ queryKey: ["admin-sellers"] });
      queryClient.invalidateQueries({ queryKey: ["admin-seller", input.sellerId] });
      queryClient.invalidateQueries({ queryKey: ["seller-catalog", input.sellerId] });
      queryClient.invalidateQueries({ queryKey: ["seller-kyc-documents", input.sellerId] });
    }
  });
}
