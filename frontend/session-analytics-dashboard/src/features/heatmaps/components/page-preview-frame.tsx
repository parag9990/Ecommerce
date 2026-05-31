import type {
  HeatmapDeviceType,
  HeatmapMode,
  HeatmapPoint
} from "../../../api/session-api";
import { HeatmapCanvas } from "./heatmap-canvas";

type PagePreviewFrameProps = {
  deviceType: HeatmapDeviceType;
  mode: HeatmapMode;
  path: string;
  points: HeatmapPoint[];
};

const previewSizes: Record<
  HeatmapDeviceType,
  { height: number; label: string; width: number }
> = {
  desktop: { height: 720, label: "Desktop", width: 960 },
  mobile: { height: 844, label: "Mobile", width: 390 },
  tablet: { height: 840, label: "Tablet", width: 720 }
};

export function PagePreviewFrame({
  deviceType,
  mode,
  path,
  points
}: PagePreviewFrameProps) {
  const size = previewSizes[deviceType];

  return (
    <section className="min-w-0 rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="mb-4 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div className="min-w-0">
          <h2 className="text-sm font-semibold text-zinc-950">Page preview</h2>
          <p className="mt-1 truncate text-sm text-zinc-500">{path}</p>
        </div>
        <div className="flex items-center gap-2">
          <span className="rounded-md border border-zinc-200 bg-zinc-50 px-2.5 py-1 text-xs font-medium text-zinc-700">
            {size.label}
          </span>
          <span className="rounded-md border border-zinc-200 bg-zinc-50 px-2.5 py-1 text-xs font-medium capitalize text-zinc-700">
            {mode}
          </span>
        </div>
      </div>

      <div className="overflow-x-auto rounded-lg bg-zinc-100 p-3">
        <div
          className="relative mx-auto overflow-hidden rounded-md border border-zinc-300 bg-white shadow-sm"
          style={{ height: size.height, width: size.width }}
        >
          <PreviewChrome path={path} />
          <HeatmapCanvas
            height={size.height}
            mode={mode}
            points={points}
            width={size.width}
          />
        </div>
      </div>
    </section>
  );
}

function PreviewChrome({ path }: { path: string }) {
  const template = getTemplate(path);

  return (
    <div aria-hidden="true" className="h-full bg-white text-zinc-950">
      <header className="border-b border-zinc-200 px-5 py-4">
        <div className="flex items-center justify-between gap-4">
          <div className="h-3 w-28 rounded-full bg-zinc-900" />
          <div className="hidden items-center gap-3 md:flex">
            <div className="h-2 w-14 rounded-full bg-zinc-200" />
            <div className="h-2 w-16 rounded-full bg-zinc-200" />
            <div className="h-2 w-12 rounded-full bg-zinc-200" />
          </div>
        </div>
      </header>

      <main className="space-y-5 p-5">
        <section className="grid gap-4 md:grid-cols-[1.2fr_0.8fr]">
          <div className="h-40 rounded-md bg-zinc-100" />
          <div className="space-y-3">
            <div className="h-4 w-3/4 rounded-full bg-zinc-300" />
            <div className="h-3 w-full rounded-full bg-zinc-200" />
            <div className="h-3 w-5/6 rounded-full bg-zinc-200" />
            <div className="h-10 w-36 rounded-md bg-zinc-950" />
          </div>
        </section>

        <section className={template.gridClassName}>
          {Array.from({ length: template.cards }).map((_, index) => (
            <div
              className="rounded-md border border-zinc-200 bg-white p-3"
              key={index}
            >
              <div className="h-24 rounded bg-zinc-100" />
              <div className="mt-3 h-3 w-4/5 rounded-full bg-zinc-200" />
              <div className="mt-2 h-3 w-1/2 rounded-full bg-zinc-200" />
            </div>
          ))}
        </section>

        <section className="space-y-3 rounded-md bg-zinc-50 p-4">
          <div className="h-3 w-44 rounded-full bg-zinc-300" />
          <div className="h-3 w-full rounded-full bg-zinc-200" />
          <div className="h-3 w-11/12 rounded-full bg-zinc-200" />
          <div className="h-3 w-2/3 rounded-full bg-zinc-200" />
        </section>
      </main>
    </div>
  );
}

function getTemplate(path: string): { cards: number; gridClassName: string } {
  if (path.includes("checkout")) {
    return {
      cards: 2,
      gridClassName: "grid gap-4 md:grid-cols-2"
    };
  }

  if (path.includes("cart")) {
    return {
      cards: 3,
      gridClassName: "grid gap-4 md:grid-cols-3"
    };
  }

  return {
    cards: 4,
    gridClassName: "grid gap-4 md:grid-cols-4"
  };
}
