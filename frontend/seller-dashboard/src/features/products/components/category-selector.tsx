import { useCategories } from "../hooks/use-categories";

type CategorySelectorProps = {
  value: string;
  error?: string;
  onChange: (categoryId: string) => void;
};

export function CategorySelector({ value, error, onChange }: CategorySelectorProps) {
  const categoriesQuery = useCategories();

  return (
    <label className="space-y-1">
      <span className="text-sm font-medium text-slate-700">Category</span>
      <select
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
      >
        <option value="">Select category</option>
        {categoriesQuery.data?.map((category) => (
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
      {error ? <p className="text-xs text-rose-600">{error}</p> : null}
    </label>
  );
}
