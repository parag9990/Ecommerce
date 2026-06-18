import type { WishlistItem } from '../types';
import { WishlistItemCard } from './wishlist-item-card';

type WishlistGridProps = {
  busyProductId?: string | undefined;
  items: WishlistItem[];
  onMoveToCart: (productId: string) => void;
  onRemove: (productId: string) => void;
};

export function WishlistGrid({
  busyProductId,
  items,
  onMoveToCart,
  onRemove,
}: WishlistGridProps) {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {items.map((item) => (
        <WishlistItemCard
          busy={busyProductId === item.product_id}
          item={item}
          key={item.product_id}
          onMoveToCart={onMoveToCart}
          onRemove={onRemove}
        />
      ))}
    </div>
  );
}
