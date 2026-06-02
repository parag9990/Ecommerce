import { CardsSkeleton, ChartSkeleton } from "../../../components/state/loading-skeleton";

export function AnalyticsLoadingState() {
  return (
    <div className="space-y-4" aria-label="Loading analytics">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div className="space-y-2">
          <div className="h-3 w-24 animate-pulse rounded bg-slate-200" />
          <div className="h-7 w-56 animate-pulse rounded bg-slate-200" />
        </div>
        <div className="h-9 w-80 animate-pulse rounded-md bg-slate-200" />
      </div>

      <CardsSkeleton count={4} />

      <div className="grid gap-3 xl:grid-cols-[2fr_1fr]">
        <ChartSkeleton />
        <ChartSkeleton />
      </div>
    </div>
  );
}
