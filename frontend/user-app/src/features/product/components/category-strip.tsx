import { Link } from 'react-router-dom';

import { routePaths } from '../../../routes/route-paths';
import type { Category } from '../types';

type CategoryStripProps = {
  categories: Category[];
};

export function CategoryStrip({ categories }: CategoryStripProps) {
  if (categories.length === 0) {
    return null;
  }

  return (
    <nav
      aria-label="Product categories"
      className="flex gap-2 overflow-x-auto pb-2"
    >
      {categories.map((category) => (
        <Link
          className="whitespace-nowrap rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-700 transition hover:border-slate-400 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          key={category.category_id}
          to={routePaths.category(category.category_id)}
        >
          {category.name}
        </Link>
      ))}
    </nav>
  );
}
