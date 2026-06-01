import type { Category, ProductFilters } from '../../product/types';

type FacetSummaryProps = {
  categories?: Category[] | undefined;
  filters: ProductFilters;
  onChange: (nextFilters: ProductFilters) => void;
};

type FacetChip = {
  label: string;
  onRemove: () => void;
};

function getCategoryName(categories: Category[], categoryId: string) {
  return (
    categories.find((category) => category.category_id === categoryId)?.name ??
    categoryId
  );
}

export function FacetSummary({
  categories = [],
  filters,
  onChange,
}: FacetSummaryProps) {
  const chips: FacetChip[] = [];

  if (filters.categoryId) {
    chips.push({
      label: `Category: ${getCategoryName(categories, filters.categoryId)}`,
      onRemove: () => {
        onChange({ ...filters, categoryId: undefined, page: 1 });
      },
    });
  }

  if (filters.brand) {
    chips.push({
      label: `Brand: ${filters.brand}`,
      onRemove: () => {
        onChange({ ...filters, brand: undefined, page: 1 });
      },
    });
  }

  if (filters.sellerId) {
    chips.push({
      label: `Seller: ${filters.sellerId}`,
      onRemove: () => {
        onChange({ ...filters, sellerId: undefined, page: 1 });
      },
    });
  }

  if (filters.minPrice || filters.maxPrice) {
    chips.push({
      label: `Price: ${filters.minPrice ?? '0'}-${filters.maxPrice ?? 'any'}`,
      onRemove: () => {
        onChange({
          ...filters,
          maxPrice: undefined,
          minPrice: undefined,
          page: 1,
        });
      },
    });
  }

  if (filters.minRating) {
    chips.push({
      label: `${filters.minRating}+ rating`,
      onRemove: () => {
        onChange({ ...filters, minRating: undefined, page: 1 });
      },
    });
  }

  if (filters.inStock) {
    chips.push({
      label: 'In stock',
      onRemove: () => {
        onChange({ ...filters, inStock: false, page: 1 });
      },
    });
  }

  if (chips.length === 0) {
    return null;
  }

  return (
    <div className="flex flex-wrap gap-2" aria-label="Active filters">
      {chips.map((chip) => (
        <button
          className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs font-medium text-slate-700 hover:border-slate-400 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          key={chip.label}
          onClick={chip.onRemove}
          type="button"
        >
          {chip.label} x
        </button>
      ))}
    </div>
  );
}
