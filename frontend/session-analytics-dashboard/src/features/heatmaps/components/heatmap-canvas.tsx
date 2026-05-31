import { useEffect, useMemo, useRef } from "react";

import type { HeatmapMode, HeatmapPoint } from "../../../api/session-api";
import {
  getHeatColor,
  getHeatOuterColor,
  getScrollBandColor
} from "../lib/heatmap-colors";
import {
  type CanvasHeatmapPoint,
  normalizeHeatmapPoints
} from "../lib/heatmap-normalize";

type HeatmapCanvasProps = {
  height: number;
  mode: HeatmapMode;
  points: HeatmapPoint[];
  width: number;
};

export function HeatmapCanvas({
  height,
  mode,
  points,
  width
}: HeatmapCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const normalizedPoints = useMemo(
    () => normalizeHeatmapPoints(points),
    [points]
  );

  useEffect(() => {
    const canvas = canvasRef.current;
    const context = canvas?.getContext("2d");

    if (!canvas || !context || width <= 0 || height <= 0) {
      return;
    }

    const pixelRatio = window.devicePixelRatio || 1;
    canvas.width = Math.round(width * pixelRatio);
    canvas.height = Math.round(height * pixelRatio);
    canvas.style.width = `${width}px`;
    canvas.style.height = `${height}px`;
    context.setTransform(pixelRatio, 0, 0, pixelRatio, 0, 0);
    context.clearRect(0, 0, width, height);

    if (mode === "scroll") {
      drawScrollDepth(context, normalizedPoints, width, height);
      return;
    }

    drawClickHeatmap(context, normalizedPoints, width, height);
  }, [height, mode, normalizedPoints, width]);

  return (
    <canvas
      aria-label={
        mode === "scroll" ? "Scroll depth heatmap overlay" : "Click heatmap overlay"
      }
      className="pointer-events-none absolute inset-0 h-full w-full"
      ref={canvasRef}
    />
  );
}

function drawClickHeatmap(
  context: CanvasRenderingContext2D,
  points: CanvasHeatmapPoint[],
  width: number,
  height: number
) {
  context.globalCompositeOperation = "source-over";

  points.forEach((point) => {
    const centerX = (point.x / 100) * width;
    const centerY = (point.y / 100) * height;
    const baseRadius = Math.max(22, Math.min(width, height) * 0.035);
    const radius = baseRadius + point.intensity * 46;
    const gradient = context.createRadialGradient(
      centerX,
      centerY,
      0,
      centerX,
      centerY,
      radius
    );

    gradient.addColorStop(0, getHeatColor(point.intensity));
    gradient.addColorStop(1, getHeatOuterColor(point.intensity));

    context.fillStyle = gradient;
    context.beginPath();
    context.arc(centerX, centerY, radius, 0, Math.PI * 2);
    context.fill();
  });
}

function drawScrollDepth(
  context: CanvasRenderingContext2D,
  points: CanvasHeatmapPoint[],
  width: number,
  height: number
) {
  const totalWeight = points.reduce((sum, point) => sum + point.weight, 0);

  if (totalWeight <= 0) {
    return;
  }

  const step = 2;
  const bandHeight = Math.ceil((step / 100) * height) + 1;

  for (let depth = 0; depth <= 100; depth += step) {
    const reachedWeight = points
      .filter((point) => point.y >= depth)
      .reduce((sum, point) => sum + point.weight, 0);
    const intensity = reachedWeight / totalWeight;

    context.fillStyle = getScrollBandColor(intensity);
    context.fillRect(0, (depth / 100) * height, width, bandHeight);
  }

  [25, 50, 75, 90].forEach((depth) => {
    const y = Math.round((depth / 100) * height);
    context.fillStyle = "rgba(24, 24, 27, 0.34)";
    context.fillRect(0, y, width, 1);
  });
}
