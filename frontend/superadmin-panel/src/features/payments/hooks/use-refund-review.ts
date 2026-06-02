import { useMutation, useQueryClient } from "@tanstack/react-query";

import { reviewRefund } from "../api/payments-api";
import type { RefundReviewDecision } from "../types";

export function useRefundReview() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: { refundId: string } & RefundReviewDecision) =>
      reviewRefund(input.refundId, {
        decision: input.decision,
        reason: input.reason
      }),
    onSuccess: (refund) => {
      void queryClient.invalidateQueries({ queryKey: ["admin-refunds"] });
      void queryClient.invalidateQueries({ queryKey: ["admin-payments"] });
      void queryClient.invalidateQueries({ queryKey: ["admin-payment-detail", refund.payment_id] });
    }
  });
}
