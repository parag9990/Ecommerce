import { Search } from 'lucide-react';
import { type FormEvent, useId, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

import { useProductAutocompleteQuery } from '../features/product/hooks/use-product-autocomplete-query';
import { routePaths } from '../routes/route-paths';

export function SearchBox() {
  const inputId = useId();
  const [searchParams] = useSearchParams();
  const currentQuery = searchParams.get('q') ?? '';
  const [queryDraft, setQueryDraft] = useState({
    sourceQuery: currentQuery,
    value: currentQuery,
  });
  const navigate = useNavigate();
  const query =
    queryDraft.sourceQuery === currentQuery ? queryDraft.value : currentQuery;
  const autocomplete = useProductAutocompleteQuery(query);
  const suggestions = autocomplete.data?.suggestions ?? [];

  function searchFor(value: string) {
    const trimmedQuery = value.trim();

    if (!trimmedQuery) {
      return;
    }

    const params = new URLSearchParams({ q: trimmedQuery });
    void navigate(`${routePaths.search}?${params.toString()}`);
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    searchFor(query);
  }

  function selectSuggestion(value: string) {
    setQueryDraft({
      sourceQuery: value,
      value,
    });
    searchFor(value);
  }

  return (
    <form
      className="relative w-full max-w-xl"
      onSubmit={handleSubmit}
      role="search"
    >
      <label className="sr-only" htmlFor={inputId}>
        Search products
      </label>
      <Search
        aria-hidden="true"
        className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400"
      />
      <input
        className="h-10 w-full rounded-md border border-slate-300 bg-white pl-10 pr-3 text-sm text-slate-950 shadow-sm outline-none transition placeholder:text-slate-400 focus:border-blue-600 focus:ring-2 focus:ring-blue-100"
        id={inputId}
        onChange={(event) => {
          setQueryDraft({
            sourceQuery: currentQuery,
            value: event.target.value,
          });
        }}
        placeholder="Search products, brands, categories"
        type="search"
        value={query}
      />
      {suggestions.length > 0 ? (
        <div className="absolute left-0 right-0 top-full z-40 mt-1 overflow-hidden rounded-md border border-slate-200 bg-white py-1 shadow-lg">
          {suggestions.map((suggestion) => (
            <button
              className="block w-full truncate px-3 py-2 text-left text-sm text-slate-700 transition hover:bg-slate-100 focus-visible:bg-slate-100 focus-visible:outline-none"
              key={suggestion}
              onMouseDown={(event) => {
                event.preventDefault();
                selectSuggestion(suggestion);
              }}
              type="button"
            >
              {suggestion}
            </button>
          ))}
        </div>
      ) : null}
    </form>
  );
}
