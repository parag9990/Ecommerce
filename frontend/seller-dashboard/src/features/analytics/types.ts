export const ANALYTICS_PRESETS = ["7d", "30d", "90d"] as const;

export type AnalyticsPreset = (typeof ANALYTICS_PRESETS)[number];
export type AnalyticsRangeMode = AnalyticsPreset | "custom";

export type Money = {
  amount: number;
  currency: string;
};

export type AnalyticsDateRange = {
  from: string;
  to: string;
  preset: AnalyticsRangeMode;
};

export type AnalyticsSeriesPoint = {
  date: string;
  revenue?: Money;
  gmv?: Money;
  orders?: number;
  conversion_rate?: number;
};

export type TopProduct = {
  product_id: string;
  name: string;
  sku?: string;
  revenue?: Money;
  gmv?: Money;
  orders?: number;
  units_sold?: number;
  conversion_rate?: number;
};

export type SellerAnalyticsResponse = {
  revenue?: Money;
  gmv?: Money;
  orders?: number;
  conversion_rate?: number;
  top_products?: unknown[];
  series?: unknown[];
};

export type NormalizedSellerAnalytics = {
  revenue: Money | null;
  gmv: Money | null;
  orders: number | null;
  conversionRate: number | null;
  topProducts: TopProduct[];
  series: AnalyticsSeriesPoint[];
};
