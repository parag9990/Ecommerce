import { useMemo } from 'react';
import { Link, useSearchParams } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { EmptyState } from '../../../components/ui/empty-state';
import { routePaths } from '../../../routes/route-paths';
import { CategoryStrip } from '../components/category-strip';
import { Pagination } from '../components/pagination';
import { ProductGrid } from '../components/product-grid';
import { ProductListSkeleton } from '../components/product-list-skeleton';
import { useCategories } from '../hooks/use-categories';
import { useProductList } from '../hooks/use-product-list';
import { readProductFilters, writeProductFilters } from '../product-url-state';
import type { ProductFilters } from '../types';

function CategoryStripSkeleton() {
  return (
    <div aria-label="Loading categories" className="flex gap-2 overflow-hidden">
      {Array.from({ length: 6 }).map((_, index) => (
        <div
          className="h-10 w-28 shrink-0 animate-pulse rounded-full bg-slate-100"
          key={`category-skeleton-${index}`}
        />
      ))}
    </div>
  );
}

export function HomePage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const filters = useMemo<ProductFilters>(() => {
    const parsedFilters = readProductFilters(searchParams);

    return {
      page: parsedFilters.page,
      pageSize: parsedFilters.pageSize,
    };
  }, [searchParams]);

  const categories = useCategories();
  const products = useProductList(filters, 'home');

  function updatePage(page: number) {
    setSearchParams(
      writeProductFilters({
        page,
        pageSize: filters.pageSize,
      }),
    );
  }

  return (
    <div className="space-y-6">
      <section className="space-y-2">
        <h1 className="text-2xl font-semibold text-slate-950">
          Shop products
        </h1>
        <p className="max-w-2xl text-sm leading-6 text-slate-600">
          Browse the latest published products, pick a category, or search for a
          specific item from the header.
        </p>
      </section>

      {categories.isLoading ? <CategoryStripSkeleton /> : null}
      {categories.error ? (
        <Alert title="Categories could not be loaded" variant="error">
          {categories.error}
        </Alert>
      ) : null}
      <CategoryStrip categories={categories.categories} />

      <section className="space-y-4" aria-labelledby="home-products-title">
        <div className="flex items-end justify-between gap-3">
          <div>
            <h2
              className="text-lg font-semibold text-slate-950"
              id="home-products-title"
            >
              Latest products
            </h2>
            <p className="text-sm text-slate-600">
              {products.data
                ? `${products.data.total} products found`
                : 'Loading products'}
            </p>
          </div>
        </div>

        {products.isLoading ? <ProductListSkeleton /> : null}

        {products.error ? (
          <Alert title="Products could not be loaded" variant="error">
            {products.error}
          </Alert>
        ) : null}

        {products.data && products.data.products.length === 0 ? (
          <EmptyState
            action={
              <Link
                className="text-sm font-semibold text-blue-700 underline-offset-4 hover:text-blue-800 hover:underline"
                to={routePaths.search}
              >
                Search products
              </Link>
            }
            description="Try searching for a specific product or check back once new products are published."
            title="No products available yet"
          />
        ) : null}

        {products.data && products.data.products.length > 0 ? (
          <>
            <ProductGrid products={products.data.products} />
            <Pagination
              onPageChange={updatePage}
              page={filters.page}
              pageSize={filters.pageSize}
              total={products.data.total}
            />
          </>
        ) : null}
      </section>
    </div>
  );
}
