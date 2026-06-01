import { Input } from '../../../components/ui/input';
import type { Category, ProductFilters } from '../types';

type FilterPanelProps = {
  categories?: Category[] | undefined;
  hideCategoryFilter?: boolean;
  onChange: (nextFilters: ProductFilters) => void;
  onClear: () => void;
  value: ProductFilters;
};

const selectClasses =
  'w-full rounded-md border border-slate-300 bg-white px-3 py-2.5 text-sm text-slate-950 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-100';

export function FilterPanel({
  categories = [],
  hideCategoryFilter = false,
  onChange,
  onClear,
  value,
}: FilterPanelProps) {
  function update(next: Partial<ProductFilters>) {
    onChange({
      ...value,
      ...next,
      page: 1,
    });
  }

  return (
    <aside className="space-y-5 rounded-md border border-slate-200 bg-white p-4">
      <div className="flex items-center justify-between gap-2">
        <h2 className="font-semibold text-slate-950">Filters</h2>
        <button
          className="rounded text-sm font-medium text-blue-700 hover:text-blue-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          onClick={onClear}
          type="button"
        >
          Clear
        </button>
      </div>

      {!hideCategoryFilter && categories.length > 0 ? (
        <label className="block space-y-1 text-sm" htmlFor="product-category">
          <span className="font-medium text-slate-700">Category</span>
          <select
            className={selectClasses}
            id="product-category"
            onChange={(event) => {
              update({ categoryId: event.target.value || undefined });
            }}
            value={value.categoryId ?? ''}
          >
            <option value="">All categories</option>
            {categories.map((category) => (
              <option key={category.category_id} value={category.category_id}>
                {category.name}
              </option>
            ))}
          </select>
        </label>
      ) : null}

      <label className="block space-y-1 text-sm" htmlFor="product-brand">
        <span className="font-medium text-slate-700">Brand</span>
        <Input
          id="product-brand"
          onChange={(event) => {
            update({ brand: event.target.value || undefined });
          }}
          placeholder="Nike, Apple..."
          value={value.brand ?? ''}
        />
      </label>

      <label className="block space-y-1 text-sm" htmlFor="product-seller">
        <span className="font-medium text-slate-700">Seller</span>
        <Input
          id="product-seller"
          onChange={(event) => {
            update({ sellerId: event.target.value || undefined });
          }}
          placeholder="Seller ID"
          value={value.sellerId ?? ''}
        />
      </label>

      <div className="grid grid-cols-2 gap-3">
        <label className="block space-y-1 text-sm" htmlFor="product-min-price">
          <span className="font-medium text-slate-700">Min price</span>
          <Input
            id="product-min-price"
            inputMode="decimal"
            onChange={(event) => {
              update({ minPrice: event.target.value || undefined });
            }}
            placeholder="0"
            value={value.minPrice ?? ''}
          />
        </label>

        <label className="block space-y-1 text-sm" htmlFor="product-max-price">
          <span className="font-medium text-slate-700">Max price</span>
          <Input
            id="product-max-price"
            inputMode="decimal"
            onChange={(event) => {
              update({ maxPrice: event.target.value || undefined });
            }}
            placeholder="5000"
            value={value.maxPrice ?? ''}
          />
        </label>
      </div>

      <label className="block space-y-1 text-sm" htmlFor="product-rating">
        <span className="font-medium text-slate-700">Minimum rating</span>
        <select
          className={selectClasses}
          id="product-rating"
          onChange={(event) => {
            update({ minRating: event.target.value || undefined });
          }}
          value={value.minRating ?? ''}
        >
          <option value="">Any rating</option>
          <option value="4">4 and above</option>
          <option value="3">3 and above</option>
          <option value="2">2 and above</option>
        </select>
      </label>

      <label className="flex items-center gap-2 text-sm font-medium text-slate-700">
        <input
          checked={value.inStock ?? false}
          className="h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-200"
          onChange={(event) => {
            update({ inStock: event.target.checked });
          }}
          type="checkbox"
        />
        In stock only
      </label>
    </aside>
  );
}
