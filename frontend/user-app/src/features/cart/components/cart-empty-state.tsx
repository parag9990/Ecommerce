import { Link } from 'react-router-dom';

import { EmptyState } from '../../../components/ui/empty-state';
import { routePaths } from '../../../routes/route-paths';

export function CartEmptyState() {
  return (
    <EmptyState
      action={
        <Link
          className="inline-flex h-10 items-center justify-center rounded-md bg-blue-600 px-4 text-sm font-semibold text-white transition hover:bg-blue-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          to={routePaths.home}
        >
          Continue shopping
        </Link>
      }
      description="Add products to your cart and they will appear here."
      title="Your cart is empty"
    />
  );
}
