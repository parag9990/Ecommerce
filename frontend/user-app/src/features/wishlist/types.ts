import type { Cart } from '../cart/types';
import type { Money } from '../../lib/format-money';

export type WishlistItemAvailability =
  | 'deleted'
  | 'in_stock'
  | 'out_of_stock'
  | 'unknown';

export type WishlistItem = {
  availability?: WishlistItemAvailability | undefined;
  image_url?: string | undefined;
  price?: Money | undefined;
  product_id: string;
  title?: string | undefined;
  variant_id?: string | undefined;
};

export type Wishlist = {
  items?: WishlistItem[] | undefined;
  user_id?: string | undefined;
  wishlist_id?: string | undefined;
};

export type WishlistItemInput = {
  product_id: string;
  variant_id?: string | undefined;
};

export type WishlistMoveToCartResponse = Cart;
