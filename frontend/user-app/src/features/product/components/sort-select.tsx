import { useId } from 'react';

import type { SortOption } from '../types';

type SortSelectProps = {
  onChange: (value: SortOption | undefined) => void;
  value?: SortOption | undefined;
};

const sortOptions: Array<{ label: string; value: SortOption }> = [
  { label: 'Popular', value: 'popularity_score:desc' },
  { label: 'Newest', value: 'created_at:desc' },
  { label: 'Price: Low to High', value: 'price:asc' },
  { label: 'Price: High to Low', value: 'price:desc' },
  { label: 'Top Rated', value: 'rating:desc' },
];

export function SortSelect({ onChange, value }: SortSelectProps) {
  const selectId = useId();

  return (
    <label
      className="flex items-center gap-2 text-sm font-medium text-slate-700"
      htmlFor={selectId}
    >
      Sort
      <select
        className="rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-950 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-100"
        id={selectId}
        onChange={(event) => {
          onChange(
            event.target.value
              ? (event.target.value as SortOption)
              : undefined,
          );
        }}
        value={value ?? ''}
      >
        <option value="">Relevance</option>
        {sortOptions.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}
