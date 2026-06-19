import { describe, expect, it } from "vitest";

import { normalizeHeatmapPoints } from "./heatmap-normalize";
import {
  calculateAverageScrollDepth,
  getHeatmapStats,
  getScrollDepthBuckets
} from "./heatmap-stats";

describe("heatmap helpers", () => {
  it("clamps coordinates and normalizes intensity", () => {
    const result = normalizeHeatmapPoints([
      { weight: 5, x: -10, y: 140 },
      { weight: 10, x: 50, y: 50 },
      { weight: -4, x: 20, y: 20 }
    ]);

    expect(result).toEqual([
      { intensity: 0.5, weight: 5, x: 0, y: 100 },
      { intensity: 1, weight: 10, x: 50, y: 50 }
    ]);
  });

  it("calculates weighted heatmap and scroll statistics", () => {
    const points = [
      { weight: 10, x: 40, y: 30 },
      { weight: 30, x: 50, y: 80 }
    ];

    expect(getHeatmapStats(points)).toMatchObject({
      averageScrollDepth: 67.5,
      maxWeight: 30,
      pointCount: 2,
      totalEvents: 40
    });
    expect(calculateAverageScrollDepth(points)).toBe(67.5);
  });

  it("builds cumulative scroll depth buckets", () => {
    const buckets = getScrollDepthBuckets([
      { weight: 20, x: 0, y: 40 },
      { weight: 10, x: 0, y: 90 }
    ]);

    expect(buckets).toEqual([
      { depth: 25, events: 30, percentage: 100 },
      { depth: 50, events: 10, percentage: 33 },
      { depth: 75, events: 10, percentage: 33 },
      { depth: 90, events: 10, percentage: 33 },
      { depth: 100, events: 0, percentage: 0 }
    ]);
  });
});
