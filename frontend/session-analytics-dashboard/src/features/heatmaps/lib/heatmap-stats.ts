import type { HeatmapPoint, HeatmapResponse } from "../../../api/session-api";

export type HeatmapStats = {
  averageScrollDepth: number;
  hottestPoint?: HeatmapPoint;
  maxWeight: number;
  pointCount: number;
  totalEvents: number;
};

export const scrollDepthThresholds = [25, 50, 75, 90, 100] as const;

export type ScrollDepthBucket = {
  depth: number;
  events: number;
  percentage: number;
};

export function getHeatmapStats(
  points: HeatmapPoint[],
  response?: HeatmapResponse
): HeatmapStats {
  const totalWeight = points.reduce((sum, point) => sum + safeWeight(point), 0);
  const maxWeight = points.reduce(
    (max, point) => Math.max(max, safeWeight(point)),
    0
  );
  const hottestPoint = points.find((point) => safeWeight(point) === maxWeight);

  return {
    averageScrollDepth:
      response?.averageScrollDepth ?? calculateAverageScrollDepth(points),
    hottestPoint,
    maxWeight: response?.maxWeight ?? maxWeight,
    pointCount: points.length,
    totalEvents: response?.totalEvents ?? totalWeight
  };
}

export function getScrollDepthBuckets(
  points: HeatmapPoint[]
): ScrollDepthBucket[] {
  const total = points.reduce((sum, point) => sum + safeWeight(point), 0);

  return scrollDepthThresholds.map((depth) => {
    const events = points
      .filter((point) => point.y >= depth)
      .reduce((sum, point) => sum + safeWeight(point), 0);

    return {
      depth,
      events,
      percentage: total > 0 ? Math.round((events / total) * 100) : 0
    };
  });
}

export function calculateAverageScrollDepth(points: HeatmapPoint[]): number {
  const total = points.reduce((sum, point) => sum + safeWeight(point), 0);

  if (total === 0) {
    return 0;
  }

  const weightedDepth = points.reduce(
    (sum, point) => sum + clampDepth(point.y) * safeWeight(point),
    0
  );

  return weightedDepth / total;
}

function safeWeight(point: HeatmapPoint): number {
  if (!Number.isFinite(point.weight)) {
    return 0;
  }

  return Math.max(0, point.weight);
}

function clampDepth(depth: number): number {
  if (!Number.isFinite(depth)) {
    return 0;
  }

  return Math.min(100, Math.max(0, depth));
}
