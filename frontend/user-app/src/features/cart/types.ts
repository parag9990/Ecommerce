export type Money = {
  amount?: number | undefined;
  currency?: string | undefined;
};

export type CartItem = {
  image_url?: string | undefined;
  item_id: string;
  line_total?: Money | undefined;
  product_id: string;
  quantity: number;
  stock_status?: 'in_stock' | 'low_stock' | 'out_of_stock' | undefined;
  title?: string | undefined;
  unit_price?: Money | undefined;
  variant_id: string;
  variant_label?: string | undefined;
};

export type Cart = {
  cart_id?: string | undefined;
  discount?: Money | undefined;
  items?: CartItem[] | undefined;
  subtotal?: Money | undefined;
  total?: Money | undefined;
  user_id?: string | undefined;
};

export type AddCartItemRequest = {
  product_id: string;
  quantity: number;
  variant_id: string;
};

export type UpdateCartItemRequest = {
  quantity: number;
};

export type CouponPreviewRequest = {
  cart_id?: string | undefined;
  coupon_code: string;
  order_id?: string | undefined;
};

export type CouponPreviewResponse = {
  coupon_id?: string | undefined;
  discount?: Money | undefined;
  reason?: string | undefined;
  valid?: boolean | undefined;
};
