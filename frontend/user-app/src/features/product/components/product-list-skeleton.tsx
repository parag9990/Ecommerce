export function ProductListSkeleton() {
  return (
    <div
      aria-label="Loading products"
      className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5"
    >
      {Array.from({ length: 10 }).map((_, index) => (
        <div
          className="overflow-hidden rounded-md border border-slate-200 bg-white"
          key={`product-skeleton-${index}`}
        >
          <div className="aspect-square animate-pulse bg-slate-100" />
          <div className="space-y-2 p-3">
            <div className="h-3 w-1/3 animate-pulse rounded bg-slate-100" />
            <div className="h-4 w-full animate-pulse rounded bg-slate-100" />
            <div className="h-4 w-2/3 animate-pulse rounded bg-slate-100" />
          </div>
        </div>
      ))}
    </div>
  );
}
