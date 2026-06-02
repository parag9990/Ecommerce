import { BarChart3, CircleDollarSign, PackageCheck, TrendingUp } from "lucide-react";

import type { NormalizedSellerAnalytics } from "../types";
import {
  formatMoney,
  formatNumber,
  formatPercentage,
} from "../utils/analytics-formatters";
import { MetricCard } from "./metric-card";

type MetricsGridProps = {
  analytics: NormalizedSellerAnalytics;
};

export function MetricsGrid({ analytics }: MetricsGridProps) {
  return (
    <section className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
      <MetricCard
        label="Revenue"
        value={formatMoney(analytics.revenue)}
        helper="Paid seller revenue in the selected range."
        icon={CircleDollarSign}
        unavailable={!analytics.revenue}
      />
      <MetricCard
        label="GMV"
        value={formatMoney(analytics.gmv)}
        helper="Gross merchandise value when the API provides it."
        icon={TrendingUp}
        unavailable={!analytics.gmv}
      />
      <MetricCard
        label="Orders"
        value={formatNumber(analytics.orders)}
        helper="Total seller orders in the selected range."
        icon={PackageCheck}
        unavailable={analytics.orders === null}
      />
      <MetricCard
        label="Conversion"
        value={formatPercentage(analytics.conversionRate)}
        helper="Session-to-order conversion when available."
        icon={BarChart3}
        unavailable={analytics.conversionRate === null}
      />
    </section>
  );
}
