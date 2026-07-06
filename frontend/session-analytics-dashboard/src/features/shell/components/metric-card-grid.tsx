import {
  Activity,
  Clock3,
  Gauge,
  Radio,
  Users
} from "lucide-react";

import {
  formatDuration,
  formatNumber,
  formatRelativeTime
} from "../../../lib/format";
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
        helper="Open sessions in the live window"
        icon={Users}
        label="Active Sessions"
        value={formatNumber(metrics.activeSessionsNow)}
      />
      <MetricCard
        helper="Recent event throughput"
        icon={Radio}
        label="Events / Min"
        tone="positive"
        value={formatNumber(metrics.eventsPerMinute)}
      />
      <MetricCard
        helper="Live metrics lookback"
        icon={Clock3}
        label="Live Window"
        value={formatDuration(metrics.windowSeconds ?? 0)}
      />
      <MetricCard
        helper="Last service measurement"
        icon={Gauge}
        label="Measured"
        value={
          metrics.measuredAt ? formatRelativeTime(metrics.measuredAt) : "Unknown"
        }
      />
    </section>
  );
}

export function isLiveMetricsEmpty(metrics: LiveMetricsResponse): boolean {
  return (
    metrics.activeUsersNow === 0 &&
    metrics.activeSessionsNow === 0 &&
    metrics.eventsPerMinute === 0
  );
}
