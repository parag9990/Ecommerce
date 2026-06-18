import type { Category, ProductStatusFilter } from "../types";
import { PRODUCT_STATUSES } from "../types";

type ProductFiltersBarProps = {
  status: ProductStatusFilter;
  categoryId: string;
  pageSize: number;
  categories: Category[];
  onStatusChange: (status: ProductStatusFilter) => void;
  onCategoryChange: (categoryId: string) => void;
  onPageSizeChange: (pageSize: number) => void;
};

const statuses: ProductStatusFilter[] = ["all", ...PRODUCT_STATUSES];
const pageSizes = [10, 20, 50];

export function ProductFiltersBar({
  status,
  categoryId,
  pageSize,
  categories,
  onStatusChange,
  onCategoryChange,
  onPageSizeChange,
}: ProductFiltersBarProps) {
  return (
    <div className="flex flex-wrap items-end gap-3 rounded-md border border-slate-200 bg-white p-3 shadow-sm">
      <div className="min-w-0 flex-1 space-y-2">
        <span className="text-xs font-semibold uppercase tracking-wide text-slate-500">
          Status
        </span>
        <div className="flex flex-wrap gap-2" role="group" aria-label="Product status">
          {statuses.map((item) => (
            <button
              key={item}
              type="button"
              onClick={() => onStatusChange(item)}
              className={
                item === status
                  ? "h-8 rounded-md bg-blue-600 px-3 text-xs font-medium capitalize text-white transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
                  : "h-8 rounded-md border border-slate-200 bg-white px-3 text-xs font-medium capitalize text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
              }
            >
              {item}
            </button>
          ))}
        </div>
      </div>

      <label className="min-w-48 space-y-1">
        <span className="text-xs font-semibold uppercase tracking-wide text-slate-500">
          Category
        </span>
        <select
          value={categoryId}
          onChange={(event) => onCategoryChange(event.target.value)}
          className="h-9 w-full rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
        >
          <option value="">All categories</option>
          {categories.map((category) => (
            <option key={category.category_id} value={category.category_id}>
              {category.name}
            </option>
          ))}
        </select>
      </label>

      <label className="w-28 space-y-1">
        <span className="text-xs font-semibold uppercase tracking-wide text-slate-500">
          Page size
        </span>
        <select
          value={pageSize}
          onChange={(event) => onPageSizeChange(Number(event.target.value))}
          className="h-9 w-full rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
        >
          {pageSizes.map((size) => (
            <option key={size} value={size}>
              {size}
            </option>
          ))}
        </select>
      </label>
    </div>
  );
}
