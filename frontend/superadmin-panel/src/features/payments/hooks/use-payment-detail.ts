import { useQuery } from "@tanstack/react-query";

import { getAdminPayment } from "../api/payments-api";

export function usePaymentDetail(paymentId?: string | null) {
  return useQuery({
    queryKey: ["admin-payment-detail", paymentId],
    queryFn: () => getAdminPayment(paymentId ?? ""),
    enabled: Boolean(paymentId),
    staleTime: 30_000
  });
}
