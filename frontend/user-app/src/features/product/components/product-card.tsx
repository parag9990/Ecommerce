import { Link } from 'react-router-dom';

import { AddToCartButton } from '../../cart/components/add-to-cart-button';
import { WishlistButton } from '../../wishlist/components/wishlist-button';
import { routePaths } from '../../../routes/route-paths';
import type { Product } from '../types';
import { Price } from './price';
import { Rating } from './rating';

type ProductCardProps = {
  product: Product;
};

function getStockLabel(stock: number | undefined) {
  if (stock === undefined) {
    return { className: 'text-slate-500', label: 'Stock unavailable' };
  }

  if (stock > 0) {
    return { className: 'text-emerald-700', label: 'In stock' };
  }

  return { className: 'text-red-700', label: 'Out of stock' };
}

export function ProductCard({ product }: ProductCardProps) {
  const primaryVariant = product.variants?.[0];
  const imageUrl = product.images?.[0];
  const stock = getStockLabel(primaryVariant?.stock_quantity);

  return (
    <article className="group overflow-hidden rounded-md border border-slate-200 bg-white transition hover:border-slate-300 hover:shadow-sm">
      <div className="relative">
        <Link
          className="block focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          to={routePaths.productDetail(product.product_id)}
        >
          <div className="aspect-square bg-slate-100">
            {imageUrl ? (
              <img
                alt={product.title}
                className="h-full w-full object-cover transition duration-200 group-hover:scale-[1.02]"
                loading="lazy"
                src={imageUrl}
              />
            ) : (
              <div className="flex h-full w-full items-center justify-center px-4 text-center text-sm text-slate-500">
                Product image coming soon
              </div>
            )}
          </div>
        </Link>
        <div className="absolute right-3 top-3">
          <WishlistButton
            iconOnly
            productId={product.product_id}
            variantId={primaryVariant?.sku}
          />
        </div>
      </div>

      <Link
        className="block focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
        to={routePaths.productDetail(product.product_id)}
      >
        <div className="space-y-2 p-3">
          <div className="space-y-1">
            {product.brand ? (
              <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                {product.brand}
              </p>
            ) : null}
            <h3 className="line-clamp-2 text-sm font-medium text-slate-950">
              {product.title}
            </h3>
          </div>

          <div className="flex items-center justify-between gap-2">
            <Price value={primaryVariant?.price} />
            <span className={`text-xs ${stock.className}`}>{stock.label}</span>
          </div>

          <Rating value={product.rating} />
        </div>
      </Link>

      <div className="border-t border-slate-100 p-3">
        <AddToCartButton
          compact
          disabled={!primaryVariant || primaryVariant.stock_quantity === 0}
          productId={product.product_id}
          showViewCartLink={false}
          variantId={primaryVariant?.sku}
        />
      </div>
    </article>
  );
}
