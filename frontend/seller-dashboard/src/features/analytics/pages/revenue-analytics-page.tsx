import { useState } from "react";

import { PermissionDeniedState } from "../../../components/state/permission-denied-state";
import { RefreshingNotice } from "../../../components/state/refreshing-notice";
import { getSafeErrorMessage } from "../../../lib/api-error";
import { useSellerStore } from "../../../stores/seller-store";
import { useSellerPermissions } from "../../team/hooks/use-seller-permissions";
import { AnalyticsEmptyState } from "../components/analytics-empty-state";
import { AnalyticsErrorState } from "../components/analytics-error-state";
import { AnalyticsLoadingState } from "../components/analytics-loading-state";
import { ConversionChart } from "../components/conversion-chart";
import { DateRangeFilter } from "../components/date-range-filter";
import { MetricsGrid } from "../components/metrics-grid";
import { OrdersChart } from "../components/orders-chart";
import { RevenueGmvChart } from "../components/revenue-gmv-chart";
import { TopProductsChart } from "../components/top-products-chart";
import { TopProductsTable } from "../components/top-products-table";
import { useSellerAnalytics } from "../hooks/use-seller-analytics";
import {
  createPresetRange,
  formatDateRangeLabel,
} from "../utils/analytics-date-range";
import { hasAnalyticsData } from "../utils/analytics-adapter";

export function RevenueAnalyticsPage() {
  const activeSeller = useSellerStore((state) => state.activeSeller);
  const permissions = useSellerPermissions();
  const canViewAnalytics = permissions.can("analytics:view");
  const [range, setRange] = useState(() => createPresetRange("30d"));
  const analyticsQuery = useSellerAnalytics(
    range,
    canViewAnalytics ? activeSeller?.seller_id : undefined,
  );

  if (!activeSeller) {
    return (
      <PermissionDeniedState
        title="Active seller unavailable"
        description="Analytics load karne ke liye active seller context required hai."
      />
    );
  }

  if (!canViewAnalytics) {
    return (
      <PermissionDeniedState
        title="Analytics access unavailable"
        description="Aapke current seller role ke paas revenue analytics dekhne ka permission nahi hai."
      />
    );
  }

  if (analyticsQuery.isPending && !analyticsQuery.data) {
    return <AnalyticsLoadingState />;
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Performance
          </p>
          <div className="mt-1 flex flex-wrap items-center gap-2">
            <h1 className="text-xl font-semibold text-slate-950">Revenue analytics</h1>
          </div>
          <p className="mt-1 text-sm text-slate-500">
            Revenue, GMV, orders, conversion, and top products for{" "}
            {formatDateRangeLabel(range)}.
          </p>
        </div>

        <DateRangeFilter value={range} onChange={setRange} />
      </div>

      <RefreshingNotice
        show={analyticsQuery.isFetching && !analyticsQuery.isPending && !analyticsQuery.isError}
      />

      {analyticsQuery.isError && analyticsQuery.data ? (
        <RefreshingNotice
          show
          failed
          message={getSafeErrorMessage(analyticsQuery.error)}
          onRetry={() => analyticsQuery.refetch()}
        />
      ) : null}

      {analyticsQuery.isError && !analyticsQuery.data ? (
        <AnalyticsErrorState
          error={analyticsQuery.error}
          onRetry={() => analyticsQuery.refetch()}
        />
      ) : null}

      {analyticsQuery.data ? (
        <div className="space-y-4">
          {!hasAnalyticsData(analyticsQuery.data) ? (
            <AnalyticsEmptyState
              kind="empty"
              title="Not enough data yet"
              description="Metrics generate hone ke liye paid orders and traffic chahiye."
            />
          ) : null}

          <MetricsGrid analytics={analyticsQuery.data} />

          <div className="grid gap-3 xl:grid-cols-[2fr_1fr]">
            <RevenueGmvChart series={analyticsQuery.data.series} />
            <ConversionChart conversionRate={analyticsQuery.data.conversionRate} />
          </div>

          <OrdersChart series={analyticsQuery.data.series} />

          <div className="grid gap-3 xl:grid-cols-[1fr_1.25fr]">
            <TopProductsChart products={analyticsQuery.data.topProducts} />
            <TopProductsTable products={analyticsQuery.data.topProducts} />
          </div>
        </div>
      ) : null}
    </section>
  );
}
