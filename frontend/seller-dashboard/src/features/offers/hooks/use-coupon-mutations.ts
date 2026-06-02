import { useMutation, useQueryClient } from "@tanstack/react-query";

import { createSellerCoupon, updateSellerCoupon } from "../api/seller-offers-api";
import type { CouponInput } from "../types";
import { offerQueryKeys } from "./query-keys";

export function useCreateCoupon() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CouponInput) => createSellerCoupon(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: offerQueryKeys.coupons() });
    },
  });
}

export function useUpdateCoupon(couponId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CouponInput) => updateSellerCoupon(couponId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: offerQueryKeys.coupons() });
    },
  });
}
