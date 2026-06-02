import { http } from "../../../lib/http";
import type {
  Campaign,
  CampaignFilters,
  CampaignInput,
  CampaignListResponse,
  Coupon,
  CouponFilters,
  CouponInput,
  CouponListResponse,
} from "../types";
import { normalizeCampaign, normalizeCoupon } from "../utils/offer-mappers";

type RawCouponListResponse = {
  coupons?: unknown[];
  total?: number;
};

type RawCampaignListResponse = {
  campaigns?: unknown[];
  total?: number;
};

function toQuery(params: Record<string, string | number | undefined>) {
  const search = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      search.set(key, String(value));
    }
  });

  return search.toString();
}

export async function listSellerCoupons(filters: CouponFilters): Promise<CouponListResponse> {
  const query = toQuery({
    status: filters.status === "all" ? undefined : filters.status,
    q: filters.q?.trim(),
    starts_from: filters.starts_from,
    ends_before: filters.ends_before,
    page: filters.page,
    page_size: filters.page_size,
  });
  const response = await http<RawCouponListResponse>(
    `/api/v1/seller/coupons${query ? `?${query}` : ""}`,
  );
  const coupons = Array.isArray(response.coupons)
    ? response.coupons.map(normalizeCoupon)
    : [];

  return {
    coupons,
    total: Number(response.total ?? coupons.length),
  };
}

export async function createSellerCoupon(input: CouponInput): Promise<Coupon> {
  const coupon = await http<unknown>("/api/v1/seller/coupons", {
    method: "POST",
    body: JSON.stringify(input),
  });

  return normalizeCoupon(coupon);
}

export async function updateSellerCoupon(
  couponId: string,
  input: CouponInput,
): Promise<Coupon> {
  const coupon = await http<unknown>(`/api/v1/seller/coupons/${couponId}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });

  return normalizeCoupon(coupon);
}

export async function listSellerCampaigns(
  filters: CampaignFilters,
): Promise<CampaignListResponse> {
  const query = toQuery({
    status: filters.status === "all" ? undefined : filters.status,
    month: filters.month,
    page: filters.page,
    page_size: filters.page_size,
  });
  const response = await http<RawCampaignListResponse>(
    `/api/v1/seller/campaigns${query ? `?${query}` : ""}`,
  );
  const campaigns = Array.isArray(response.campaigns)
    ? response.campaigns.map(normalizeCampaign)
    : [];

  return {
    campaigns,
    total: Number(response.total ?? campaigns.length),
  };
}

export async function createSellerCampaign(input: CampaignInput): Promise<Campaign> {
  const campaign = await http<unknown>("/api/v1/seller/campaigns", {
    method: "POST",
    body: JSON.stringify(input),
  });

  return normalizeCampaign(campaign);
}
