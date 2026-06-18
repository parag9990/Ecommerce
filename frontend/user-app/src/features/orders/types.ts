import type { Money } from '../../lib/format-money';

export type OrderItem = {
  image_url?: string | undefined;
  price?: Money | undefined;
  product_id?: string | undefined;
  quantity?: number | undefined;
  title?: string | undefined;
  variant_id?: string | undefined;
  variant_label?: string | undefined;
};

export type OrderAddressSnapshot = {
  city?: string | undefined;
  country?: string | undefined;
  line1?: string | undefined;
  line2?: string | undefined;
  name?: string | undefined;
  phone?: string | undefined;
  postal_code?: string | undefined;
  state?: string | undefined;
};

export type OrderStatusEvent = {
  created_at?: string | undefined;
  note?: string | undefined;
  status?: string | undefined;
};

export type Order = {
  created_at?: string | undefined;
  fulfillment_status?: string | undefined;
  items?: OrderItem[] | undefined;
  order_id?: string | undefined;
  payment_status?: string | undefined;
  shipping_address?: OrderAddressSnapshot | undefined;
  status?: string | undefined;
  status_history?: OrderStatusEvent[] | undefined;
  total?: Money | undefined;
  user_id?: string | undefined;
};

export type OrderListResponse = {
  orders?: Order[] | undefined;
  total?: number | undefined;
};

export type CancelOrderRequest = {
  reason: string;
};
