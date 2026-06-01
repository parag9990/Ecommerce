import { ShoppingCart, Trash2 } from 'lucide-react';
import { Link } from 'react-router-dom';

import { Button } from '../../../components/ui/button';
import { formatMoney } from '../../../lib/format-money';
import { routePaths } from '../../../routes/route-paths';
import type { WishlistItem } from '../types';

type WishlistItemCardProps = {
  busy?: boolean | undefined;
  item: WishlistItem;
  onMoveToCart: (productId: string) => void;
  onRemove: (productId: string) => void;
};

export function WishlistItemCard({
  busy = false,
  item,
  onMoveToCart,
  onRemove,
}: WishlistItemCardProps) {
  const isUnavailable =
    item.availability === 'out_of_stock' || item.availability === 'deleted';

  return (
    <article className="overflow-hidden rounded-md border border-slate-200 bg-white">
      <Link
        className="block focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
        to={routePaths.productDetail(item.product_id)}
      >
        <div className="aspect-square bg-slate-100">
          {item.image_url ? (
            <img
              alt={item.title ?? 'Wishlist product'}
              className="h-full w-full object-cover"
              src={item.image_url}
            />
          ) : (
            <div className="flex h-full items-center justify-center px-4 text-center text-sm text-slate-500">
              Product image unavailable
            </div>
          )}
        </div>
      </Link>

      <div className="space-y-3 p-4">
        <div>
          <Link
            className="line-clamp-2 font-semibold text-slate-950 underline-offset-4 hover:text-blue-700 hover:underline"
            to={routePaths.productDetail(item.product_id)}
          >
            {item.title ?? 'Saved product'}
          </Link>
          <p className="mt-1 text-sm font-semibold text-slate-700">
            {formatMoney(item.price)}
          </p>
          {isUnavailable ? (
            <p className="mt-2 text-sm text-red-700">Currently unavailable</p>
          ) : null}
        </div>

        <div className="flex gap-2">
          <Button
            className="flex-1"
            disabled={busy || isUnavailable}
            onClick={() => {
              onMoveToCart(item.product_id);
            }}
            type="button"
          >
            <ShoppingCart aria-hidden="true" className="mr-2 h-4 w-4" />
            Move to cart
          </Button>
          <button
            aria-label={`Remove ${item.title ?? 'product'} from wishlist`}
            className="inline-flex h-11 w-11 items-center justify-center rounded-md border border-slate-300 text-slate-700 transition hover:bg-red-50 hover:text-red-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600 disabled:cursor-not-allowed disabled:opacity-60"
            disabled={busy}
            onClick={() => {
              onRemove(item.product_id);
            }}
            title="Remove from wishlist"
            type="button"
          >
            <Trash2 aria-hidden="true" className="h-4 w-4" />
          </button>
        </div>
      </div>
    </article>
  );
}
