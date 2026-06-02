export const DISCOUNT_TYPES = ["fixed", "percentage"] as const;
export type DiscountType = (typeof DISCOUNT_TYPES)[number];

export const COUPON_STATUSES = ["draft", "active", "paused", "expired"] as const;
export type CouponStatus = (typeof COUPON_STATUSES)[number];
export type CouponStatusFilter = CouponStatus | "all";

export const CAMPAIGN_STATUSES = ["draft", "active", "paused", "completed"] as const;
export type CampaignStatus = (typeof CAMPAIGN_STATUSES)[number];
export type CampaignStatusFilter = CampaignStatus | "all";

export type Money = {
  amount: number;
  currency: string;
};

export type CouponInput = {
  code: string;
  discount_type: DiscountType;
  discount_value: number;
  min_cart_amount?: Money;
  starts_at?: string;
  ends_at?: string;
  usage_limit?: number;
};

export type Coupon = CouponInput & {
  coupon_id: string;
  seller_id?: string;
  status: CouponStatus;
  used_count?: number;
  total_discount?: Money;
  created_at?: string;
  updated_at?: string;
};

export type CouponListResponse = {
  coupons: Coupon[];
  total: number;
};

export type CampaignInput = {
  name: string;
  starts_at: string;
  ends_at: string;
  budget?: Money;
};

export type Campaign = CampaignInput & {
  campaign_id: string;
  seller_id?: string;
  status: CampaignStatus;
  used_budget?: Money;
  created_at?: string;
  updated_at?: string;
};

export type CampaignListResponse = {
  campaigns: Campaign[];
  total: number;
};

export type CouponFilters = {
  status?: CouponStatusFilter;
  q?: string;
  starts_from?: string;
  ends_before?: string;
  page: number;
  page_size: number;
};

export type CampaignFilters = {
  status?: CampaignStatusFilter;
  month?: string;
  page: number;
  page_size: number;
};

export type OffersTab = "coupons" | "campaigns";
