export const ORDER_STATUSES = [
  "created",
  "pending_payment",
  "payment_failed",
  "paid",
  "packed",
  "shipped",
  "delivered",
  "cancelled",
  "refunded"
] as const;

export type OrderStatus = (typeof ORDER_STATUSES)[number];

export const REVIEW_STATUSES = ["none", "disputed", "manual_review", "resolved"] as const;

export type ReviewStatus = (typeof REVIEW_STATUSES)[number];

export type Money = {
  amount: number;
  currency: string;
};

export type AdminOrderItem = {
  item_id: string;
  product_id?: string | null;
  seller_id?: string | null;
  title: string;
  quantity: number;
  unit_price?: Money | null;
  fulfillment_status?: string | null;
};

export type OrderPaymentSummary = {
  payment_id?: string | null;
  provider?: string | null;
  status?: string | null;
  refund_status?: string | null;
  amount?: Money | null;
};

export type AdminOrder = {
  order_id: string;
  user_id: string;
  status: OrderStatus;
  review_status?: ReviewStatus;
  total: Money;
  items: AdminOrderItem[];
  payment?: OrderPaymentSummary | null;
  created_at?: string | null;
  updated_at?: string | null;
};

export type OrderStatusEvent = {
  status: OrderStatus | ReviewStatus | string;
  note?: string | null;
  actor_type: "system" | "seller" | "admin" | "payment" | "support";
  actor_id?: string | null;
  created_at?: string | null;
};

export type OrderShipment = {
  shipment_id: string;
  seller_id?: string | null;
  status: string;
  carrier?: string | null;
  tracking_number?: string | null;
  shipped_at?: string | null;
  delivered_at?: string | null;
};

export type OrderDispute = {
  dispute_id: string;
  order_id: string;
  type:
    | "delivery_issue"
    | "damaged_item"
    | "wrong_item"
    | "seller_claim"
    | "buyer_claim"
    | "status_mismatch"
    | string;
  status: "open" | "in_review" | "resolved" | "rejected" | string;
  opened_by: "buyer" | "seller" | "support" | "system" | string;
  summary: string;
  created_at?: string | null;
};

export type ManualReviewDecision = {
  decision: "mark_reviewing" | "resolve" | "escalate";
  reason: string;
  internal_note?: string;
};

export type AdminOrderFilters = {
  q?: string;
  status?: OrderStatus | "all";
  review_status?: ReviewStatus | "all";
  user_id?: string;
  seller_id?: string;
  from?: string;
  to?: string;
  page: number;
  limit: number;
};

export type AdminOrderListResponse = {
  orders: AdminOrder[];
  total?: number;
  page?: number;
  limit?: number;
};

export type AdminOrderDetailResponse = {
  order: AdminOrder;
  status_history: OrderStatusEvent[];
  shipments: OrderShipment[];
  payment?: OrderPaymentSummary | null;
};

export type OrderDisputeListResponse = {
  disputes: OrderDispute[];
};

export type SuccessResponse = {
  success: boolean;
};
