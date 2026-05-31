import type {
  HeatmapDeviceType,
  HeatmapMode,
  HeatmapPoint
} from "../../../api/session-api";

export type HeatmapFilters = {
  deviceType: HeatmapDeviceType;
  mode: HeatmapMode;
  path: string;
};

export type CanvasHeatmapPoint = HeatmapPoint & {
  intensity: number;
};

export const defaultHeatmapFilters: HeatmapFilters = {
  deviceType: "desktop",
  mode: "click",
  path: "/products/prod_123"
};

export const heatmapPagePathOptions = [
  { label: "Home", value: "/" },
  { label: "Product detail", value: "/products/prod_123" },
  { label: "Search results", value: "/search" },
  { label: "Cart", value: "/cart" },
  { label: "Checkout", value: "/checkout" }
] as const;

export function normalizeHeatmapPoints(
  points: HeatmapPoint[]
): CanvasHeatmapPoint[] {
  const safePoints = points
    .map((point) => ({
      weight: normalizeWeight(point.weight),
      x: clampPercent(point.x),
      y: clampPercent(point.y)
    }))
    .filter((point) => point.weight > 0);
  const maxWeight = Math.max(...safePoints.map((point) => point.weight), 1);

  return safePoints.map((point) => ({
    ...point,
    intensity: Math.max(0.08, point.weight / maxWeight)
  }));
}

export function clampPercent(value: number): number {
  if (!Number.isFinite(value)) {
    return 0;
  }

  return Math.min(100, Math.max(0, value));
}

function normalizeWeight(weight: number): number {
  if (!Number.isFinite(weight)) {
    return 0;
  }

  return Math.max(0, weight);
}
