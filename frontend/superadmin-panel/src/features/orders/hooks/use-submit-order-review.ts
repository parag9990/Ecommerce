import { useMutation, useQueryClient } from "@tanstack/react-query";

import { submitOrderReview } from "../api/orders-api";
import type { ManualReviewDecision } from "../types";

export function useSubmitOrderReview(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: ManualReviewDecision) => submitOrderReview(orderId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-orders"] });
      queryClient.invalidateQueries({ queryKey: ["admin-order", orderId] });
      queryClient.invalidateQueries({ queryKey: ["order-disputes", orderId] });
    }
  });
}
