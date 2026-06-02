import type { Coupon, Money } from "../types";
import { formatShortDate } from "./offer-date-rules";

export function formatMoney(money: Money) {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: money.currency || "INR",
    maximumFractionDigits: 0,
  }).format(money.amount);
}

export function formatDiscount(coupon: Pick<Coupon, "discount_type" | "discount_value" | "min_cart_amount">) {
  if (coupon.discount_type === "percentage") {
    return `${coupon.discount_value}% off`;
  }

  return `${formatMoney({
    amount: coupon.discount_value,
    currency: coupon.min_cart_amount?.currency ?? "INR",
  })} off`;
}

export function formatWindow(startsAt?: string, endsAt?: string) {
  if (!startsAt && !endsAt) {
    return "Always available";
  }

  if (startsAt && endsAt) {
    return `${formatShortDate(startsAt)} - ${formatShortDate(endsAt)}`;
  }

  if (startsAt) {
    return `From ${formatShortDate(startsAt)}`;
  }

  return `Until ${formatShortDate(endsAt)}`;
}

export function formatStatusLabel(value: string) {
  return value
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}
