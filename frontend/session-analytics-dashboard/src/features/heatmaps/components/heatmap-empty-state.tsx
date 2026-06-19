import { MousePointerClick } from "lucide-react";

export function HeatmapEmptyState() {
  return (
    <section className="rounded-lg border border-dashed border-zinc-300 bg-white px-4 py-10 text-center shadow-panel">
      <div className="mx-auto flex h-10 w-10 items-center justify-center rounded-md bg-zinc-100 text-zinc-600">
        <MousePointerClick className="h-5 w-5" aria-hidden="true" />
      </div>
      <h2 className="mt-4 text-sm font-semibold text-zinc-950">
        No heatmap data for this selection
      </h2>
      <p className="mx-auto mt-2 max-w-md text-sm text-zinc-500">
        Widen the date range or choose another page, device, or heatmap mode.
      </p>
    </section>
  );
}
