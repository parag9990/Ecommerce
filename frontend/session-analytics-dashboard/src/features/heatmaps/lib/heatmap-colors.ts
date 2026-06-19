export type HeatTone = "cool" | "warm" | "hot" | "peak";

export type HeatColorStop = {
  label: string;
  tone: HeatTone;
  swatchClassName: string;
};

export const heatmapColorStops: HeatColorStop[] = [
  {
    label: "Low",
    swatchClassName: "bg-sky-400",
    tone: "cool"
  },
  {
    label: "Medium",
    swatchClassName: "bg-emerald-400",
    tone: "warm"
  },
  {
    label: "High",
    swatchClassName: "bg-amber-400",
    tone: "hot"
  },
  {
    label: "Peak",
    swatchClassName: "bg-red-500",
    tone: "peak"
  }
];

export function getHeatColor(intensity: number): string {
  const value = clampIntensity(intensity);

  if (value >= 0.75) {
    return `rgba(239, 68, 68, ${0.52 + value * 0.3})`;
  }

  if (value >= 0.45) {
    return `rgba(245, 158, 11, ${0.42 + value * 0.28})`;
  }

  if (value >= 0.2) {
    return `rgba(16, 185, 129, ${0.36 + value * 0.24})`;
  }

  return `rgba(14, 165, 233, ${0.28 + value * 0.2})`;
}

export function getHeatOuterColor(intensity: number): string {
  const value = clampIntensity(intensity);

  if (value >= 0.75) {
    return "rgba(239, 68, 68, 0)";
  }

  if (value >= 0.45) {
    return "rgba(245, 158, 11, 0)";
  }

  if (value >= 0.2) {
    return "rgba(16, 185, 129, 0)";
  }

  return "rgba(14, 165, 233, 0)";
}

export function getScrollBandColor(intensity: number): string {
  const value = clampIntensity(intensity);

  if (value >= 0.75) {
    return `rgba(239, 68, 68, ${0.16 + value * 0.28})`;
  }

  if (value >= 0.45) {
    return `rgba(245, 158, 11, ${0.14 + value * 0.24})`;
  }

  if (value >= 0.2) {
    return `rgba(16, 185, 129, ${0.12 + value * 0.2})`;
  }

  return `rgba(14, 165, 233, ${0.1 + value * 0.16})`;
}

function clampIntensity(intensity: number): number {
  if (!Number.isFinite(intensity)) {
    return 0;
  }

  return Math.min(1, Math.max(0, intensity));
}
