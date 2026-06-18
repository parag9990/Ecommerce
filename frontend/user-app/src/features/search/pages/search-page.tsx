import { useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { EmptyState } from '../../../components/ui/empty-state';
import { FilterPanel } from '../../product/components/filter-panel';
import { Pagination } from '../../product/components/pagination';
import { ProductGrid } from '../../product/components/product-grid';
import { ProductListSkeleton } from '../../product/components/product-list-skeleton';
import { SortSelect } from '../../product/components/sort-select';
import { useCategories } from '../../product/hooks/use-categories';
import { useProductList } from '../../product/hooks/use-product-list';
import {
  readProductFilters,
  writeProductFilters,
} from '../../product/product-url-state';
import type { ProductFilters } from '../../product/types';
import { FacetSummary } from '../components/facet-summary';

export function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const categories = useCategories();

  const filters = useMemo<ProductFilters>(() => {
    return readProductFilters(searchParams);
  }, [searchParams]);

  const products = useProductList(filters, 'search');
  const queryLabel = filters.q ? `"${filters.q}"` : 'all products';

  function updateFilters(nextFilters: ProductFilters) {
    setSearchParams(writeProductFilters(nextFilters));
  }

  function clearFilters() {
    setSearchParams(
      writeProductFilters({
        page: 1,
        pageSize: filters.pageSize,
        q: filters.q,
      }),
    );
  }

  return (
    <div className="space-y-6">
      <section className="space-y-1">
        <h1 className="text-2xl font-semibold text-slate-950">
          Search results for {queryLabel}
        </h1>
        <p className="text-sm text-slate-600">
          {products.data
            ? `${products.data.total} products found`
            : 'Loading products'}
        </p>
      </section>

      {categories.error ? (
        <Alert title="Categories could not be loaded" variant="error">
          {categories.error}
        </Alert>
      ) : null}

      <div className="grid gap-6 lg:grid-cols-[280px_1fr]">
        <FilterPanel
          categories={categories.categories}
          onChange={updateFilters}
          onClear={clearFilters}
          value={filters}
        />

        <section className="space-y-4" aria-label="Search results">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <FacetSummary
              categories={categories.categories}
              filters={filters}
              onChange={updateFilters}
            />
            <div className="sm:ml-auto">
              <SortSelect
                onChange={(sort) => {
                  updateFilters({ ...filters, page: 1, sort });
                }}
                value={filters.sort}
              />
            </div>
          </div>

          {products.isLoading ? <ProductListSkeleton /> : null}

          {products.error ? (
            <Alert title="Search results could not be loaded" variant="error">
              {products.error}
            </Alert>
          ) : null}

          {products.data && products.data.products.length === 0 ? (
            <EmptyState
              description="Check the spelling, clear filters, or try a shorter search query."
              title="No matching products"
            />
          ) : null}

          {products.data && products.data.products.length > 0 ? (
            <>
              <ProductGrid products={products.data.products} />
              <Pagination
                onPageChange={(page) => {
                  updateFilters({ ...filters, page });
                }}
                page={filters.page}
                pageSize={filters.pageSize}
                total={products.data.total}
              />
            </>
          ) : null}
        </section>
      </div>
    </div>
  );
}
