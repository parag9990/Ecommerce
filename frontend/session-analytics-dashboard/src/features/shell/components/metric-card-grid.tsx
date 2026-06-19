import {
  Activity,
  Clock3,
  MousePointerClick,
  ShoppingCart,
  TrendingUp,
  Users
} from "lucide-react";

import { formatDuration, formatNumber, formatPercent } from "../../../lib/format";
import type { LiveMetricsResponse } from "../types";
import { MetricCard } from "./metric-card";

type MetricCardGridProps = {
  metrics: LiveMetricsResponse;
};

export function MetricCardGrid({ metrics }: MetricCardGridProps) {
  return (
    <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      <MetricCard
        helper="Currently active sessions"
        icon={Activity}
        label="Active Users Now"
        tone="positive"
        value={formatNumber(metrics.activeUsersNow)}
      />
      <MetricCard
        helper="Sessions started today"
        icon={Users}
        label="Sessions Today"
        value={formatNumber(metrics.sessionsToday)}
      />
      <MetricCard
        helper="Checkout to paid conversion"
        icon={TrendingUp}
        label="Conversion Rate"
        tone="positive"
        value={formatPercent(metrics.conversionRate)}
      />
      <MetricCard
        helper="Mean engagement time"
        icon={Clock3}
        label="Avg. Session Duration"
        value={formatDuration(metrics.averageSessionDurationSeconds)}
      />
      <MetricCard
        helper="Single-page sessions"
        icon={MousePointerClick}
        label="Bounce Rate"
        tone={metrics.bounceRate > 60 ? "warning" : "neutral"}
        value={formatPercent(metrics.bounceRate)}
      />
      <MetricCard
        helper="Product views becoming carts"
        icon={ShoppingCart}
        label="View to Cart Rate"
        value={formatPercent(metrics.productViewToCartRate)}
      />
    </section>
  );
}

export function isLiveMetricsEmpty(metrics: LiveMetricsResponse): boolean {
  return (
    metrics.activeUsersNow === 0 &&
    metrics.sessionsToday === 0 &&
    metrics.conversionRate === 0 &&
    metrics.averageSessionDurationSeconds === 0 &&
    metrics.bounceRate === 0 &&
    metrics.productViewToCartRate === 0
  );
}
