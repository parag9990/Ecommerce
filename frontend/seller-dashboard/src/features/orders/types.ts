export const ORDER_STATUSES = [
  "created",
  "pending_payment",
  "paid",
  "packed",
  "shipped",
  "delivered",
  "cancelled",
  "refunded",
] as const;

export type OrderStatus = (typeof ORDER_STATUSES)[number];
export type OrderStatusFilter = OrderStatus | "all";

export type Money = {
  amount: number;
  currency: string;
};

export type OrderItem = {
  product_id?: string;
  variant_id?: string;
  seller_id?: string;
  title?: string;
  sku?: string;
  quantity?: number;
  unit_price?: Money;
  total?: Money;
};

export type Shipment = {
  shipment_id?: string;
  status?: OrderStatus;
  tracking_number?: string;
  carrier?: string;
  shipped_at?: string;
  delivered_at?: string;
  updated_at?: string;
};

export type Refund = {
  refund_id?: string;
  payment_id?: string;
  status: string;
  amount: Money;
  reason?: string;
  created_at?: string;
};

export type StatusHistoryEntry = {
  status: OrderStatus;
  created_at: string;
  note?: string;
};

export type CustomerSummary = {
  user_id?: string;
  name?: string;
  email?: string;
  phone?: string;
};

export type ShippingAddress = {
  name?: string;
  phone?: string;
  line1?: string;
  line2?: string;
  city?: string;
  state?: string;
  country?: string;
  postal_code?: string;
};

export type Order = {
  order_id: string;
  user_id?: string;
  status: OrderStatus;
  items: OrderItem[];
  total: Money;
  created_at: string;
  updated_at?: string;
  customer?: CustomerSummary;
  shipping_address?: ShippingAddress;
  shipments?: Shipment[];
  refunds?: Refund[];
  status_history?: StatusHistoryEntry[];
};

export type SellerOrderFilters = {
  status?: OrderStatusFilter;
  q?: string;
  page: number;
  page_size: number;
  date_from?: string;
  date_to?: string;
};

export type OrderListResponse = {
  orders: Order[];
  total: number;
};

export type FulfillmentUpdateInput = {
  status?: OrderStatus;
  tracking_number?: string;
  carrier?: string;
};
