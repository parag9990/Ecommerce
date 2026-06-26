import { useMemo } from 'react';
import { Link, useSearchParams } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { EmptyState } from '../../../components/ui/empty-state';
import { routePaths } from '../../../routes/route-paths';
import { Pagination } from '../components/pagination';
import { ProductGrid } from '../components/product-grid';
import { ProductListSkeleton } from '../components/product-list-skeleton';
import { useProductList } from '../hooks/use-product-list';
import { readProductFilters, writeProductFilters } from '../product-url-state';
import type { ProductFilters } from '../types';

export function DealsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const filters = useMemo<ProductFilters>(() => {
    const parsedFilters = readProductFilters(searchParams);

    return {
      page: parsedFilters.page,
      pageSize: parsedFilters.pageSize,
      sort: 'popularity_score:desc',
    };
  }, [searchParams]);
  const products = useProductList(filters, 'search');

  function updatePage(page: number) {
    setSearchParams(
      writeProductFilters({
        page,
        pageSize: filters.pageSize,
        sort: filters.sort,
      }),
    );
  }

  return (
    <div className="space-y-6">
      <section className="space-y-2">
        <h1 className="text-2xl font-semibold text-slate-950">
          Deals and popular picks
        </h1>
        <p className="max-w-2xl text-sm leading-6 text-slate-600">
          Browse popular products and confirm each item&apos;s current price and
          availability before checkout.
        </p>
      </section>

      <section className="space-y-4" aria-labelledby="deals-products-title">
        <div>
          <h2
            className="text-lg font-semibold text-slate-950"
            id="deals-products-title"
          >
            Popular products
          </h2>
          <p className="text-sm text-slate-600">
            {products.data
              ? `${products.data.total} products found`
              : 'Loading products'}
          </p>
        </div>

        {products.isLoading ? <ProductListSkeleton /> : null}

        {products.error ? (
          <Alert title="Popular products could not be loaded" variant="error">
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
            description="Try searching for a product or check back once the catalog is updated."
            title="No popular products available yet"
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
