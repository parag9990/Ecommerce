import { heatmapColorStops } from "../lib/heatmap-colors";

type HeatmapLegendProps = {
  compact?: boolean;
};

export function HeatmapLegend({ compact = false }: HeatmapLegendProps) {
  return (
    <section
      aria-label="Heatmap intensity legend"
      className={[
        "rounded-lg border border-zinc-200 bg-white p-4 shadow-panel",
        compact ? "" : "min-w-0"
      ].join(" ")}
    >
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-semibold text-zinc-950">Intensity</h2>
        <span className="text-xs font-medium text-zinc-500">Aggregate</span>
      </div>
      <div className="mt-4 grid grid-cols-4 gap-2">
        {heatmapColorStops.map((stop) => (
          <div key={stop.tone}>
            <div className={`h-2 rounded-full ${stop.swatchClassName}`} />
            <p className="mt-1 text-xs font-medium text-zinc-500">{stop.label}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
