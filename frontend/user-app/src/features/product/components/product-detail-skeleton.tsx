export function ProductDetailSkeleton() {
  return (
    <div className="grid gap-8 lg:grid-cols-2">
      <div className="aspect-square animate-pulse rounded-md bg-slate-100" />
      <div className="space-y-4">
        <div className="h-4 w-24 animate-pulse rounded bg-slate-100" />
        <div className="h-9 w-2/3 animate-pulse rounded bg-slate-100" />
        <div className="h-5 w-32 animate-pulse rounded bg-slate-100" />
        <div className="h-24 w-full animate-pulse rounded bg-slate-100" />
        <div className="grid gap-2 sm:grid-cols-2">
          <div className="h-24 animate-pulse rounded-md bg-slate-100" />
          <div className="h-24 animate-pulse rounded-md bg-slate-100" />
        </div>
      </div>
    </div>
  );
}
