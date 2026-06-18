import { Link } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { EmptyState } from '../../../components/ui/empty-state';
import { routePaths } from '../../../routes/route-paths';
import { useCategories } from '../hooks/use-categories';

export function CategoriesPage() {
  const { categories, error, isLoading } = useCategories();

  return (
    <div className="space-y-6">
      <section className="space-y-2">
        <h1 className="text-2xl font-semibold text-slate-950">
          Shop by category
        </h1>
        <p className="max-w-2xl text-sm leading-6 text-slate-600">
          Choose a category to browse matching products with filters, sort, and
          pagination.
        </p>
      </section>

      {isLoading ? (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, index) => (
            <div
              className="h-24 animate-pulse rounded-md bg-slate-100"
              key={`category-card-skeleton-${index}`}
            />
          ))}
        </div>
      ) : null}

      {error ? (
        <Alert title="Categories could not be loaded" variant="error">
          {error}
        </Alert>
      ) : null}

      {!isLoading && !error && categories.length === 0 ? (
        <EmptyState
          description="Categories will appear here once the catalog team publishes them."
          title="No categories available"
        />
      ) : null}

      {categories.length > 0 ? (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {categories.map((category) => (
            <Link
              className="rounded-md border border-slate-200 bg-white p-4 transition hover:border-slate-300 hover:shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
              key={category.category_id}
              to={routePaths.category(category.category_id)}
            >
              <h2 className="font-semibold text-slate-950">
                {category.name}
              </h2>
              <p className="mt-1 text-sm text-slate-600">
                Browse products in this category
              </p>
            </Link>
          ))}
        </div>
      ) : null}
    </div>
  );
}
