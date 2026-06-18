import type { Money } from "../types";

const unavailableLabel = "Not available";

function formatCurrency(amount: number, currency: string) {
  const normalizedCurrency = currency || "INR";

  try {
    return new Intl.NumberFormat("en-IN", {
      style: "currency",
      currency: normalizedCurrency,
      maximumFractionDigits: 0,
    }).format(amount);
  } catch {
    return new Intl.NumberFormat("en-IN", {
      style: "currency",
      currency: "INR",
      maximumFractionDigits: 0,
    }).format(amount);
  }
}

export function moneyToMajorUnit(value: Money) {
  return value.amount / 100;
}

export function formatMoney(value: Money | null | undefined) {
  if (!value) {
    return unavailableLabel;
  }

  return formatCurrency(moneyToMajorUnit(value), value.currency);
}

export function formatCompactMoney(value: Money | null | undefined) {
  if (!value) {
    return unavailableLabel;
  }

  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: value.currency || "INR",
    notation: "compact",
    maximumFractionDigits: 1,
  }).format(moneyToMajorUnit(value));
}

export function formatNumber(value: number | null | undefined) {
  if (value === null || value === undefined) {
    return unavailableLabel;
  }

  return new Intl.NumberFormat("en-IN").format(value);
}

export function formatCompactNumber(value: number | null | undefined) {
  if (value === null || value === undefined) {
    return unavailableLabel;
  }

  return new Intl.NumberFormat("en-IN", {
    notation: "compact",
    maximumFractionDigits: 1,
  }).format(value);
}

export function formatPercentage(value: number | null | undefined) {
  if (value === null || value === undefined) {
    return unavailableLabel;
  }

  return `${value.toFixed(2)}%`;
}

export function formatShortDate(value: string) {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("en-IN", {
    day: "2-digit",
    month: "short",
  }).format(date);
}
