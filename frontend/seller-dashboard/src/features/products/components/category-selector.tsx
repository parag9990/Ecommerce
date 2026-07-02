import { useCategories } from "../hooks/use-categories";

type CategorySelectorProps = {
  value: string;
  error?: string;
  onChange: (categoryId: string) => void;
};

export function CategorySelector({ value, error, onChange }: CategorySelectorProps) {
  const categoriesQuery = useCategories();
  const categories = categoriesQuery.data ?? [];
  const hasCategories = categories.length > 0;
  const hasSelectedMissingCategory = Boolean(
    value && !categories.some((category) => category.category_id === value),
  );

  return (
    <label className="space-y-1">
      <span className="text-sm font-medium text-slate-700">Category</span>
      <select
        value={value}
        onChange={(event) => onChange(event.target.value)}
        disabled={categoriesQuery.isLoading || !hasCategories}
        className="h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:cursor-not-allowed disabled:bg-slate-50 disabled:text-slate-500"
      >
        <option value="">{categoriesQuery.isLoading ? "Loading categories..." : "Select category"}</option>
        {hasSelectedMissingCategory ? <option value={value}>{value}</option> : null}
        {categories.map((category) => (
          <option key={category.category_id} value={category.category_id}>
            {category.name}
          </option>
        ))}
      </select>

      {categoriesQuery.isLoading ? (
        <p className="text-xs text-slate-500">Loading categories...</p>
      ) : null}
      {categoriesQuery.isError ? (
        <p className="text-xs text-rose-600">Categories could not be loaded.</p>
      ) : null}
      {!categoriesQuery.isLoading && !categoriesQuery.isError && !hasCategories ? (
        <p className="text-xs text-rose-600">No active categories are available.</p>
      ) : null}
      {error ? <p className="text-xs text-rose-600">{error}</p> : null}
    </label>
  );
}
