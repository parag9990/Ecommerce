import { Heart } from 'lucide-react';
import { useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

import { ApiError } from '../../../lib/http';
import { hasAuthSession } from '../../../lib/auth-session';
import { routePaths } from '../../../routes/route-paths';
import { useAuthStore } from '../../../stores/auth-store';
import { useWishlistMutations } from '../hooks/use-wishlist-mutations';
import { useWishlistQuery } from '../hooks/use-wishlist-query';

type WishlistButtonProps = {
  className?: string | undefined;
  iconOnly?: boolean | undefined;
  initialInWishlist?: boolean | undefined;
  productId: string;
  variantId?: string | undefined;
};

export function WishlistButton({
  className = '',
  iconOnly = false,
  initialInWishlist = false,
  productId,
  variantId,
}: WishlistButtonProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const [error, setError] = useState<string>();
  const user = useAuthStore((state) => state.user);
  const wishlistQuery = useWishlistQuery({
    enabled: Boolean(user) || hasAuthSession(),
  });
  const { addItem, removeItem } = useWishlistMutations();
  const cachedItem = wishlistQuery.data?.items?.find(
    (item) => item.product_id === productId,
  );
  const inWishlist = wishlistQuery.data ? Boolean(cachedItem) : initialInWishlist;
  const isSaving = addItem.isPending || removeItem.isPending;

  async function toggleWishlist() {
    setError(undefined);

    try {
      if (inWishlist) {
        await removeItem.mutateAsync(productId);
      } else {
        await addItem.mutateAsync({ productId, variantId });
      }
    } catch (apiError) {
      if (apiError instanceof ApiError && apiError.status === 401) {
        const redirectTo = encodeURIComponent(location.pathname + location.search);
        void navigate(`${routePaths.login}?redirect=${redirectTo}`);
        return;
      }

      setError(
        apiError instanceof Error
          ? apiError.message
          : 'Wishlist could not be updated.',
      );
    }
  }

  const label = inWishlist ? 'Remove from wishlist' : 'Add to wishlist';

  return (
    <div className={iconOnly ? 'inline-flex' : ''}>
      <button
        aria-pressed={inWishlist}
        className={[
          iconOnly
            ? 'inline-flex h-10 w-10 items-center justify-center rounded-full border border-slate-200 bg-white/95 text-slate-700 shadow-sm transition hover:bg-white hover:text-red-700'
            : 'inline-flex h-11 items-center justify-center rounded-md border border-slate-300 bg-white px-4 text-sm font-semibold text-slate-700 transition hover:bg-slate-50 hover:text-red-700',
          'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-60',
          inWishlist ? 'text-red-700' : '',
          className,
        ].join(' ')}
        disabled={isSaving}
        onClick={() => {
          void toggleWishlist();
        }}
        title={label}
        type="button"
      >
        <Heart
          aria-hidden="true"
          className={['h-5 w-5', inWishlist ? 'fill-current' : ''].join(' ')}
        />
        {iconOnly ? <span className="sr-only">{label}</span> : label}
      </button>
      {error && !iconOnly ? (
        <p className="mt-2 text-sm text-red-700">{error}</p>
      ) : null}
    </div>
  );
}
