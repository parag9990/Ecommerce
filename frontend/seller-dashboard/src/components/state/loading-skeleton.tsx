import { cn } from "../../lib/cn";

type TableSkeletonProps = {
  rows?: number;
  columns?: number;
  titleWidth?: string;
  className?: string;
};

type CountSkeletonProps = {
  count?: number;
  className?: string;
};

type RowsSkeletonProps = {
  rows?: number;
  className?: string;
};

export function TableSkeleton({
  rows = 6,
  columns = 5,
  titleWidth = "w-40",
  className,
}: TableSkeletonProps) {
  return (
    <div
      className={cn(
        "overflow-hidden rounded-md border border-slate-200 bg-white shadow-sm",
        className,
      )}
      aria-label="Loading table"
    >
      <div className="border-b border-slate-200 p-4">
        <div className={cn("h-4 animate-pulse rounded bg-slate-200", titleWidth)} />
      </div>
      <div className="divide-y divide-slate-100">
        {Array.from({ length: rows }).map((_, rowIndex) => (
          <div
            key={rowIndex}
            className="grid gap-3 p-4"
            style={{
              gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))`,
            }}
          >
            {Array.from({ length: columns }).map((__, columnIndex) => (
              <div
                key={columnIndex}
                className={cn(
                  "h-3 animate-pulse rounded bg-slate-200",
                  columnIndex === 0 ? "w-3/4" : "w-full",
                )}
              />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

export function CardsSkeleton({ count = 4, className }: CountSkeletonProps) {
  return (
    <div className={cn("grid gap-3 sm:grid-cols-2 xl:grid-cols-4", className)}>
      {Array.from({ length: count }).map((_, index) => (
        <div
          key={index}
          className="rounded-md border border-slate-200 bg-white p-4 shadow-sm"
        >
          <div className="h-3 w-24 animate-pulse rounded bg-slate-200" />
          <div className="mt-3 h-7 w-32 animate-pulse rounded bg-slate-200" />
          <div className="mt-2 h-3 w-20 animate-pulse rounded bg-slate-100" />
        </div>
      ))}
    </div>
  );
}

export function ChartSkeleton({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "h-80 animate-pulse rounded-md border border-slate-200 bg-white shadow-sm",
        className,
      )}
      aria-label="Loading chart"
    />
  );
}

export function FormSkeleton({ rows = 6, className }: RowsSkeletonProps) {
  return (
    <div
      className={cn(
        "space-y-4 rounded-md border border-slate-200 bg-white p-4 shadow-sm",
        className,
      )}
      aria-label="Loading form"
    >
      <div className="grid gap-4 md:grid-cols-2">
        {Array.from({ length: rows }).map((_, index) => (
          <div key={index} className="space-y-2">
            <div className="h-3 w-24 animate-pulse rounded bg-slate-200" />
            <div className="h-10 animate-pulse rounded-md bg-slate-100" />
          </div>
        ))}
      </div>
    </div>
  );
}

export function DetailSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn("space-y-4", className)} aria-label="Loading detail">
      <CardsSkeleton count={3} className="md:grid-cols-3 xl:grid-cols-3" />
      <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_360px]">
        <ChartSkeleton className="h-96" />
        <div className="space-y-4">
          <ChartSkeleton className="h-44" />
          <ChartSkeleton className="h-44" />
        </div>
      </div>
    </div>
  );
}

export function TimelineSkeleton({ rows = 5, className }: RowsSkeletonProps) {
  return (
    <div className={cn("space-y-3", className)} aria-label="Loading activity">
      {Array.from({ length: rows }).map((_, index) => (
        <div
          key={index}
          className="h-28 animate-pulse rounded-md border border-slate-200 bg-white shadow-sm"
        />
      ))}
    </div>
  );
}
