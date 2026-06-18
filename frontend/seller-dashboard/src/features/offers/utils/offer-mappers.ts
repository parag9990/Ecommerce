import type {
  Campaign,
  CampaignInput,
  CampaignStatus,
  Coupon,
  CouponInput,
  CouponStatus,
  DiscountType,
  Money,
} from "../types";
import { CAMPAIGN_STATUSES, COUPON_STATUSES, DISCOUNT_TYPES } from "../types";
import { fromDateTimeLocal, toDateTimeLocal } from "./offer-date-rules";
import type {
  CampaignFormInput,
  CampaignFormValues,
  CouponFormInput,
  CouponFormValues,
} from "./offer-validation";

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};
}

function normalizeString(value: unknown) {
  return typeof value === "string" ? value : "";
}

function normalizeOptionalString(value: unknown) {
  return typeof value === "string" && value.length > 0 ? value : undefined;
}

function normalizeNumber(value: unknown) {
  const numberValue = Number(value);
  return Number.isFinite(numberValue) ? numberValue : 0;
}

function normalizeOptionalNumber(value: unknown) {
  const numberValue = Number(value);
  return Number.isFinite(numberValue) ? numberValue : undefined;
}

function isDiscountType(value: unknown): value is DiscountType {
  return DISCOUNT_TYPES.some((type) => type === value);
}

function normalizeDiscountType(value: unknown): DiscountType {
  return isDiscountType(value) ? value : "fixed";
}

function isCouponStatus(value: unknown): value is CouponStatus {
  return COUPON_STATUSES.some((status) => status === value);
}

function normalizeCouponStatus(value: unknown): CouponStatus {
  return isCouponStatus(value) ? value : "draft";
}

function isCampaignStatus(value: unknown): value is CampaignStatus {
  return CAMPAIGN_STATUSES.some((status) => status === value);
}

function normalizeCampaignStatus(value: unknown): CampaignStatus {
  return isCampaignStatus(value) ? value : "draft";
}

export function normalizeMoney(value: unknown): Money {
  const candidate = asRecord(value);
  const currency = normalizeOptionalString(candidate.currency) ?? "INR";

  return {
    amount: normalizeNumber(candidate.amount),
    currency,
  };
}

function normalizeOptionalMoney(value: unknown) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return undefined;
  }

  return normalizeMoney(value);
}

export function normalizeCoupon(value: unknown): Coupon {
  if (!value || typeof value !== "object") {
    throw new Error("Coupon response was empty or invalid.");
  }

  const candidate = value as Record<string, unknown>;

  return {
    coupon_id:
      normalizeOptionalString(candidate.coupon_id) ??
      normalizeOptionalString(candidate.id) ??
      "",
    seller_id: normalizeOptionalString(candidate.seller_id),
    code: normalizeString(candidate.code),
    discount_type: normalizeDiscountType(candidate.discount_type),
    discount_value: normalizeNumber(candidate.discount_value),
    min_cart_amount: normalizeOptionalMoney(candidate.min_cart_amount),
    starts_at: normalizeOptionalString(candidate.starts_at),
    ends_at: normalizeOptionalString(candidate.ends_at),
    usage_limit: normalizeOptionalNumber(candidate.usage_limit),
    status: normalizeCouponStatus(candidate.status),
    used_count:
      normalizeOptionalNumber(candidate.used_count) ??
      normalizeOptionalNumber(candidate.redeemed_count) ??
      normalizeOptionalNumber(candidate.redemption_count),
    total_discount: normalizeOptionalMoney(candidate.total_discount),
    created_at: normalizeOptionalString(candidate.created_at),
    updated_at: normalizeOptionalString(candidate.updated_at),
  };
}

export function normalizeCampaign(value: unknown): Campaign {
  if (!value || typeof value !== "object") {
    throw new Error("Campaign response was empty or invalid.");
  }

  const candidate = value as Record<string, unknown>;

  return {
    campaign_id:
      normalizeOptionalString(candidate.campaign_id) ??
      normalizeOptionalString(candidate.id) ??
      "",
    seller_id: normalizeOptionalString(candidate.seller_id),
    name: normalizeString(candidate.name),
    starts_at: normalizeString(candidate.starts_at),
    ends_at: normalizeString(candidate.ends_at),
    budget: normalizeOptionalMoney(candidate.budget),
    status: normalizeCampaignStatus(candidate.status),
    used_budget: normalizeOptionalMoney(candidate.used_budget),
    created_at: normalizeOptionalString(candidate.created_at),
    updated_at: normalizeOptionalString(candidate.updated_at),
  };
}

function optionalMoneyInput(value?: { amount?: number; currency?: string }) {
  if (!value || value.amount === undefined || value.amount <= 0) {
    return undefined;
  }

  return {
    amount: Number(value.amount),
    currency: value.currency?.trim().toUpperCase() || "INR",
  };
}

export function toCouponInput(values: CouponFormValues): CouponInput {
  return {
    code: values.code.trim().toUpperCase(),
    discount_type: values.discount_type,
    discount_value: Number(values.discount_value),
    min_cart_amount: optionalMoneyInput(values.min_cart_amount),
    starts_at: fromDateTimeLocal(values.starts_at),
    ends_at: fromDateTimeLocal(values.ends_at),
    usage_limit: values.usage_limit ? Number(values.usage_limit) : undefined,
  };
}

export function toCouponFormValues(coupon?: Coupon): CouponFormInput {
  return {
    code: coupon?.code ?? "",
    discount_type: coupon?.discount_type ?? "percentage",
    discount_value: coupon?.discount_value ?? 10,
    min_cart_amount: {
      amount: coupon?.min_cart_amount?.amount,
      currency: coupon?.min_cart_amount?.currency ?? "INR",
    },
    starts_at: toDateTimeLocal(coupon?.starts_at) || undefined,
    ends_at: toDateTimeLocal(coupon?.ends_at) || undefined,
    usage_limit: coupon?.usage_limit,
  };
}

export function toCampaignInput(values: CampaignFormValues): CampaignInput {
  return {
    name: values.name.trim(),
    starts_at: fromDateTimeLocal(values.starts_at) ?? values.starts_at,
    ends_at: fromDateTimeLocal(values.ends_at) ?? values.ends_at,
    budget: optionalMoneyInput(values.budget),
  };
}
