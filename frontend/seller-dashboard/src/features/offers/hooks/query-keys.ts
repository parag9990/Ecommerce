import type { CampaignFilters, CouponFilters } from "../types";

export const offerQueryKeys = {
  all: ["seller-offers"] as const,
  coupons: () => [...offerQueryKeys.all, "coupons"] as const,
  couponList: (sellerId: string, filters: CouponFilters) =>
    [...offerQueryKeys.coupons(), sellerId, filters] as const,
  campaigns: () => [...offerQueryKeys.all, "campaigns"] as const,
  campaignList: (sellerId: string, filters: CampaignFilters) =>
    [...offerQueryKeys.campaigns(), sellerId, filters] as const,
};
