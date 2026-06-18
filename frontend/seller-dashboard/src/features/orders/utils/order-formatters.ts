import type { Money, OrderStatus } from "../types";

export function formatOrderStatus(status: OrderStatus | string) {
  return status.replace(/_/g, " ");
}

export function formatMoney(money?: Money | null) {
  if (!money) {
    return "-";
  }

  const amount = Number(money.amount);
  const currency = money.currency || "INR";
  const majorAmount = Number.isFinite(amount) ? amount / 100 : 0;

  try {
    return new Intl.NumberFormat("en-IN", {
      style: "currency",
      currency,
      maximumFractionDigits: 2,
    }).format(majorAmount);
  } catch {
    return `${currency} ${majorAmount.toFixed(2)}`;
  }
}

export function formatDateTime(value?: string | null) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "-";
  }

  return new Intl.DateTimeFormat("en-IN", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function formatItemCount(count: number) {
  return `${count} ${count === 1 ? "item" : "items"}`;
}
