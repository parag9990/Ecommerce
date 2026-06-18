import { ShoppingCart } from 'lucide-react';
import { Link } from 'react-router-dom';

import { useCartQuery } from '../features/cart/hooks/use-cart-query';
import { hasAuthSession } from '../lib/auth-session';
import { routePaths } from '../routes/route-paths';
import { useAuthStore } from '../stores/auth-store';

type CartBadgeProps = {
  count?: number;
};

export function CartBadge({ count = 0 }: CartBadgeProps) {
  const user = useAuthStore((state) => state.user);
  const { data: cart } = useCartQuery({
    enabled: Boolean(user) || hasAuthSession(),
  });
  const itemCount =
    cart?.items?.reduce((sum, item) => sum + item.quantity, 0) ?? count;
  const displayCount = itemCount > 99 ? '99+' : String(itemCount);

  return (
    <Link
      aria-label={`Cart with ${itemCount} items`}
      className="relative inline-flex h-10 w-10 items-center justify-center rounded-md text-slate-700 transition hover:bg-slate-100 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
      to={routePaths.cart}
    >
      <ShoppingCart aria-hidden="true" className="h-5 w-5" />
      {itemCount > 0 ? (
        <span className="absolute -right-1 -top-1 min-w-5 rounded-full bg-blue-600 px-1.5 py-0.5 text-center text-xs font-semibold leading-none text-white">
          {displayCount}
        </span>
      ) : null}
    </Link>
  );
}
