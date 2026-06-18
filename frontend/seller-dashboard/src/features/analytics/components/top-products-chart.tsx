import type { TopProduct } from "../types";
import {
  getTopProductChartMetric,
  toTopProductRows,
  type TopProductChartMetric,
} from "../utils/analytics-chart-data";
import { formatMoney, formatNumber, formatPercentage } from "../utils/analytics-formatters";
import { AnalyticsEmptyState } from "./analytics-empty-state";

type TopProductsChartProps = {
  products: TopProduct[];
};

const metricLabels: Record<TopProductChartMetric, string> = {
  revenue: "Revenue",
  orders: "Orders",
  units_sold: "Units sold",
  conversion_rate: "Conversion",
};

function formatMetricValue(value: number, metric: TopProductChartMetric, currency: string) {
  if (metric === "revenue") {
    return formatMoney({ amount: Math.round(value * 100), currency });
  }

  if (metric === "conversion_rate") {
    return formatPercentage(value);
  }

  return formatNumber(value);
}

export function TopProductsChart({ products }: TopProductsChartProps) {
  const metric = getTopProductChartMetric(products);
  const rows = metric ? toTopProductRows(products, metric) : [];

  if (!metric || rows.length === 0) {
    return (
      <AnalyticsEmptyState
        title="Top products"
        description="Top product metrics are not available for the selected range."
        className="min-h-80"
      />
    );
  }

  const maxValue = Math.max(...rows.map((row) => row.value), 1);
  const currency = products.find((product) => product.revenue)?.revenue?.currency ?? "INR";

  return (
    <section className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
      <div>
        <h2 className="text-base font-semibold text-slate-950">Top products</h2>
        <p className="mt-1 text-xs text-slate-500">
          Ranked by {metricLabels[metric].toLowerCase()}.
        </p>
      </div>

      <div className="mt-4 space-y-3">
        {rows.map((row) => (
          <div key={row.productId} className="grid grid-cols-[minmax(7rem,12rem)_1fr_auto] items-center gap-3">
            <span className="truncate text-sm font-medium text-slate-700" title={row.name}>
              {row.name}
            </span>
            <div className="h-3 overflow-hidden rounded-full bg-slate-100">
              <div
                className="h-full rounded-full bg-cyan-600"
                style={{ width: `${Math.max(3, (row.value / maxValue) * 100)}%` }}
              />
            </div>
            <span className="text-right text-xs font-medium text-slate-500">
              {formatMetricValue(row.value, metric, currency)}
            </span>
          </div>
        ))}
      </div>
    </section>
  );
}
