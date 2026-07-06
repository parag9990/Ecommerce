import type { FunnelStepKey } from "../../../api/session-api";

export const defaultFunnelSteps: FunnelStepKey[] = [
  "product_view",
  "add_to_cart",
  "checkout_step",
  "payment_result"
];

export const funnelStepLabels: Record<FunnelStepKey, string> = {
  add_to_cart: "Added to cart",
  checkout_step: "Checkout started",
  payment_result: "Paid",
  product_view: "Product viewed"
};

export const smallCountThreshold = 5;

export function formatFunnelCount(
  value: number,
  options: { maskSmallCounts?: boolean; threshold?: number } = {}
): string {
  const safeValue = Math.max(0, Math.round(Number.isFinite(value) ? value : 0));
  const threshold = options.threshold ?? smallCountThreshold;

  if (options.maskSmallCounts !== false && safeValue > 0 && safeValue < threshold) {
    return `<${threshold}`;
  }

  return new Intl.NumberFormat("en-US").format(safeValue);
}

export function formatFunnelPercent(value: number | null): string {
  if (value === null || !Number.isFinite(value)) {
    return "-";
  }

  return `${value.toFixed(1)}%`;
}
