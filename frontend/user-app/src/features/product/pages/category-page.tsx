import { useMemo } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { EmptyState } from '../../../components/ui/empty-state';
import { FacetSummary } from '../../search/components/facet-summary';
import { FilterPanel } from '../components/filter-panel';
import { Pagination } from '../components/pagination';
import { ProductGrid } from '../components/product-grid';
import { ProductListSkeleton } from '../components/product-list-skeleton';
import { SortSelect } from '../components/sort-select';
import { useCategories } from '../hooks/use-categories';
import { useProductList } from '../hooks/use-product-list';
import { readProductFilters, writeProductFilters } from '../product-url-state';
import type { ProductFilters } from '../types';

export function CategoryPage() {
  const { categoryId } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const categories = useCategories();

  const filters = useMemo<ProductFilters>(() => {
    return {
      ...readProductFilters(searchParams),
      categoryId,
    };
  }, [categoryId, searchParams]);

  const products = useProductList(filters, 'search');
  const category = categories.categories.find(
    (item) => item.category_id === categoryId,
  );
  const title = category?.name ?? 'Category products';

  function updateFilters(nextFilters: ProductFilters) {
    setSearchParams(
      writeProductFilters({
        ...nextFilters,
        categoryId: undefined,
      }),
    );
  }

  function clearFilters() {
    setSearchParams(
      writeProductFilters({
        page: 1,
        pageSize: filters.pageSize,
      }),
    );
  }

  return (
    <div className="space-y-6">
      <section className="space-y-1">
        <h1 className="text-2xl font-semibold text-slate-950">{title}</h1>
        <p className="text-sm text-slate-600">
          {products.data
            ? `${products.data.total} products found`
            : 'Loading products'}
        </p>
      </section>

      <div className="grid gap-6 lg:grid-cols-[280px_1fr]">
        <FilterPanel
          hideCategoryFilter
          onChange={updateFilters}
          onClear={clearFilters}
          value={filters}
        />

        <section className="space-y-4" aria-label="Category product results">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <FacetSummary
              categories={categories.categories}
              filters={{ ...filters, categoryId: undefined }}
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
            <Alert title="Category products could not be loaded" variant="error">
              {products.error}
            </Alert>
          ) : null}

          {products.data && products.data.products.length === 0 ? (
            <EmptyState
              description="Clear filters or try another category to find more products."
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
