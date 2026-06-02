import { keepPreviousData, useQuery } from "@tanstack/react-query";

import { listSellerCoupons } from "../api/seller-offers-api";
import type { CouponFilters } from "../types";
import { offerQueryKeys } from "./query-keys";

export function useSellerCoupons(filters: CouponFilters, sellerId?: string) {
  return useQuery({
    queryKey: offerQueryKeys.couponList(sellerId ?? "", filters),
    queryFn: () => listSellerCoupons(filters),
    enabled: Boolean(sellerId),
    placeholderData: keepPreviousData,
  });
}
