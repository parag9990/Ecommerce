import type { Campaign, Coupon } from "../types";
import { isExpired } from "./offer-date-rules";

export type UsageStats = {
  totalCoupons: number;
  activeCoupons: number;
  expiredCoupons: number;
  scheduledCampaigns: number;
  knownUsedCount: number;
  totalUsageLimit: number;
  totalDiscountAmount?: number;
  totalDiscountCurrency?: string;
};

export function buildUsageStats(coupons: Coupon[], campaigns: Campaign[]): UsageStats {
  const activeCoupons = coupons.filter((coupon) => coupon.status === "active").length;
  const expiredCoupons = coupons.filter(
    (coupon) => coupon.status === "expired" || isExpired(coupon.ends_at),
  ).length;
  const scheduledCampaigns = campaigns.filter(
    (campaign) => campaign.status !== "completed" && !isExpired(campaign.ends_at),
  ).length;
  const knownUsedCount = coupons.reduce((total, coupon) => total + (coupon.used_count ?? 0), 0);
  const totalUsageLimit = coupons.reduce((total, coupon) => total + (coupon.usage_limit ?? 0), 0);
  const discounts = coupons
    .map((coupon) => coupon.total_discount)
    .filter((money): money is NonNullable<typeof money> => Boolean(money));

  const discountCurrency = discounts[0]?.currency;
  const canAggregateDiscount =
    Boolean(discountCurrency) &&
    discounts.every((money) => money.currency === discountCurrency);

  return {
    totalCoupons: coupons.length,
    activeCoupons,
    expiredCoupons,
    scheduledCampaigns,
    knownUsedCount,
    totalUsageLimit,
    totalDiscountAmount: canAggregateDiscount
      ? discounts.reduce((total, money) => total + money.amount, 0)
      : undefined,
    totalDiscountCurrency: canAggregateDiscount ? discountCurrency : undefined,
  };
}
