export function TablePagination({
  page,
  limit,
  itemCount,
  totalCount,
  isFetching = false,
  onPageChange
}: {
  page: number;
  limit: number;
  itemCount: number;
  totalCount?: number;
  isFetching?: boolean;
  onPageChange: (page: number) => void;
}) {
  const hasKnownTotal = typeof totalCount === "number";
  const hasNextPage = hasKnownTotal ? page * limit < totalCount : itemCount >= limit;

  return (
    <div className="flex flex-col gap-2 border-t border-slate-200 bg-white px-4 py-3 text-sm text-slate-600 sm:flex-row sm:items-center sm:justify-between">
      <p>
        Page {page}
        {hasKnownTotal ? ` of ${Math.max(Math.ceil(totalCount / limit), 1)}` : ""}
        {isFetching ? " - refreshing" : ""}
      </p>
      <div className="flex items-center gap-2">
        <button
          type="button"
          disabled={page <= 1 || isFetching}
          onClick={() => onPageChange(page - 1)}
          className="h-9 rounded-md border border-slate-300 px-3 font-medium text-slate-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Previous
        </button>
        <button
          type="button"
          disabled={!hasNextPage || isFetching}
          onClick={() => onPageChange(page + 1)}
          className="h-9 rounded-md border border-slate-300 px-3 font-medium text-slate-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Next
        </button>
      </div>
    </div>
  );
}
