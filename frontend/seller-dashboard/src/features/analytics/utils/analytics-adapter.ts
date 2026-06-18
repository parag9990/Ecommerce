import type {
  AnalyticsSeriesPoint,
  Money,
  NormalizedSellerAnalytics,
  SellerAnalyticsResponse,
  TopProduct,
} from "../types";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function normalizeString(value: unknown) {
  return typeof value === "string" && value.trim().length > 0 ? value.trim() : undefined;
}

function normalizeNumber(value: unknown) {
  const numberValue = typeof value === "number" ? value : Number(value);
  return Number.isFinite(numberValue) ? numberValue : undefined;
}

function normalizeMoney(value: unknown): Money | undefined {
  if (!isRecord(value)) {
    return undefined;
  }

  const amount = normalizeNumber(value.amount);
  const currency = normalizeString(value.currency);

  if (amount === undefined || !currency) {
    return undefined;
  }

  return { amount, currency };
}

function normalizeTopProduct(value: unknown): TopProduct | null {
  if (!isRecord(value)) {
    return null;
  }

  const productId = normalizeString(value.product_id) ?? normalizeString(value.id);
  const name =
    normalizeString(value.name) ??
    normalizeString(value.product_name) ??
    normalizeString(value.title);

  if (!productId || !name) {
    return null;
  }

  return {
    product_id: productId,
    name,
    sku: normalizeString(value.sku),
    revenue: normalizeMoney(value.revenue),
    gmv: normalizeMoney(value.gmv),
    orders: normalizeNumber(value.orders) ?? normalizeNumber(value.order_count),
    units_sold:
      normalizeNumber(value.units_sold) ??
      normalizeNumber(value.units) ??
      normalizeNumber(value.quantity_sold),
    conversion_rate:
      normalizeNumber(value.conversion_rate) ?? normalizeNumber(value.conversion),
  };
}

function normalizeSeriesPoint(value: unknown): AnalyticsSeriesPoint | null {
  if (!isRecord(value)) {
    return null;
  }

  const date = normalizeString(value.date) ?? normalizeString(value.day) ?? normalizeString(value.bucket);

  if (!date) {
    return null;
  }

  return {
    date,
    revenue: normalizeMoney(value.revenue),
    gmv: normalizeMoney(value.gmv),
    orders: normalizeNumber(value.orders) ?? normalizeNumber(value.order_count),
    conversion_rate:
      normalizeNumber(value.conversion_rate) ?? normalizeNumber(value.conversion),
  };
}

export function normalizeSellerAnalytics(
  response: SellerAnalyticsResponse,
): NormalizedSellerAnalytics {
  return {
    revenue: normalizeMoney(response.revenue) ?? null,
    gmv: normalizeMoney(response.gmv) ?? null,
    orders: normalizeNumber(response.orders) ?? null,
    conversionRate: normalizeNumber(response.conversion_rate) ?? null,
    topProducts: (response.top_products ?? [])
      .map(normalizeTopProduct)
      .filter((product): product is TopProduct => product !== null),
    series: (response.series ?? [])
      .map(normalizeSeriesPoint)
      .filter((point): point is AnalyticsSeriesPoint => point !== null),
  };
}

export function hasAnalyticsData(analytics: NormalizedSellerAnalytics) {
  return Boolean(
    analytics.revenue ||
      analytics.gmv ||
      analytics.orders !== null ||
      analytics.conversionRate !== null ||
      analytics.topProducts.length > 0 ||
      analytics.series.length > 0,
  );
}
