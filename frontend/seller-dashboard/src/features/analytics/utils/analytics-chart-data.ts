import type { AnalyticsSeriesPoint, TopProduct } from "../types";
import { moneyToMajorUnit } from "./analytics-formatters";

export type RevenueGmvChartRow = {
  label: string;
  revenue?: number;
  gmv?: number;
};

export type OrdersChartRow = {
  label: string;
  orders: number;
};

export type TopProductChartMetric = "revenue" | "orders" | "units_sold" | "conversion_rate";

export type TopProductChartRow = {
  productId: string;
  name: string;
  value: number;
};

export function toRevenueGmvRows(series: AnalyticsSeriesPoint[]): RevenueGmvChartRow[] {
  return series
    .map((point) => ({
      label: point.date,
      revenue: point.revenue ? moneyToMajorUnit(point.revenue) : undefined,
      gmv: point.gmv ? moneyToMajorUnit(point.gmv) : undefined,
    }))
    .filter((row) => row.revenue !== undefined || row.gmv !== undefined);
}

export function toOrdersRows(series: AnalyticsSeriesPoint[]): OrdersChartRow[] {
  return series
    .filter((point) => point.orders !== undefined)
    .map((point) => ({
      label: point.date,
      orders: point.orders ?? 0,
    }));
}

export function getTopProductChartMetric(products: TopProduct[]): TopProductChartMetric | null {
  if (products.some((product) => product.revenue)) {
    return "revenue";
  }

  if (products.some((product) => product.orders !== undefined)) {
    return "orders";
  }

  if (products.some((product) => product.units_sold !== undefined)) {
    return "units_sold";
  }

  if (products.some((product) => product.conversion_rate !== undefined)) {
    return "conversion_rate";
  }

  return null;
}

export function toTopProductRows(
  products: TopProduct[],
  metric: TopProductChartMetric,
): TopProductChartRow[] {
  return products
    .map((product) => {
      const value =
        metric === "revenue"
          ? moneyToMajorUnit(product.revenue ?? { amount: 0, currency: "INR" })
          : product[metric] ?? 0;

      return {
        productId: product.product_id,
        name: product.name,
        value,
      };
    })
    .filter((row) => row.value > 0)
    .sort((left, right) => right.value - left.value)
    .slice(0, 8);
}
