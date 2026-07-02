import { apiFetch } from "../../../lib/http";
import type {
  AdminOrder,
  AdminOrderDetailResponse,
  AdminOrderFilters,
  AdminOrderItem,
  AdminOrderListResponse,
  ManualReviewDecision,
  Money,
  OrderDispute,
  OrderDisputeListResponse,
  OrderPaymentSummary,
  OrderShipment,
  OrderStatus,
  OrderStatusEvent,
  ReviewStatus,
  SuccessResponse
} from "../types";
import { ORDER_STATUSES, REVIEW_STATUSES } from "../types";

const ORDERS_PATH = "/api/v1/admin/orders";
const DEFAULT_ORDER_STATUS: OrderStatus = "created";
const DEFAULT_REVIEW_STATUS: ReviewStatus = "none";
const DEFAULT_CURRENCY = "INR";

type RawOrder = Partial<Omit<AdminOrder, "items" | "status" | "review_status" | "total">> & {
  status?: string | null;
  review_status?: string | null;
  total?: Partial<Money> | null;
  items?: Array<Partial<AdminOrderItem>>;
  payment?: Partial<OrderPaymentSummary> | null;
};

type RawOrderDetailResponse = Partial<Omit<AdminOrderDetailResponse, "order">> & {
  order?: RawOrder;
};

function buildQueryString(values: Record<string, string | number | undefined>): string {
  const params = new URLSearchParams();

  Object.entries(values).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      params.set(key, String(value));
    }
  });

  return params.toString();
}

function isOrderStatus(status?: string | null): status is OrderStatus {
  return ORDER_STATUSES.includes(status as OrderStatus);
}

function isReviewStatus(status?: string | null): status is ReviewStatus {
  return REVIEW_STATUSES.includes(status as ReviewStatus);
}

function normalizeMoney(money?: Partial<Money> | null): Money {
  return {
    amount: typeof money?.amount === "number" ? money.amount : 0,
    currency: money?.currency?.trim() || DEFAULT_CURRENCY
  };
}

function normalizeOrderItem(item: Partial<AdminOrderItem>, index: number): AdminOrderItem {
  const productId = item.product_id?.trim();
  const itemId = item.item_id?.trim() || productId || `item_${index + 1}`;

  return {
    item_id: itemId,
    product_id: productId,
    seller_id: item.seller_id?.trim() || undefined,
    title: item.title?.trim() || productId || "Order item",
    quantity: typeof item.quantity === "number" && item.quantity > 0 ? item.quantity : 1,
    unit_price: item.unit_price ? normalizeMoney(item.unit_price) : null,
    fulfillment_status: item.fulfillment_status?.trim() || undefined
  };
}

function normalizePaymentSummary(payment?: Partial<OrderPaymentSummary> | null): OrderPaymentSummary | null {
  if (!payment) {
    return null;
  }

  return {
    payment_id: payment.payment_id?.trim() || undefined,
    provider: payment.provider?.trim() || undefined,
    status: payment.status?.trim() || undefined,
    refund_status: payment.refund_status?.trim() || undefined,
    amount: payment.amount ? normalizeMoney(payment.amount) : null
  };
}

function normalizeOrder(order: RawOrder): AdminOrder {
  const reviewStatus = isReviewStatus(order.review_status) ? order.review_status : DEFAULT_REVIEW_STATUS;

  return {
    ...order,
    order_id: order.order_id?.trim() || "unknown_order",
    user_id: order.user_id?.trim() || "unknown_user",
    status: isOrderStatus(order.status) ? order.status : DEFAULT_ORDER_STATUS,
    review_status: reviewStatus,
    total: normalizeMoney(order.total),
    items: (order.items ?? []).map(normalizeOrderItem),
    payment: normalizePaymentSummary(order.payment),
    created_at: order.created_at ?? null,
    updated_at: order.updated_at ?? null
  };
}

function normalizeStatusEvent(event: Partial<OrderStatusEvent>): OrderStatusEvent {
  return {
    status: event.status ?? "created",
    note: event.note ?? null,
    actor_type: event.actor_type ?? "system",
    actor_id: event.actor_id ?? null,
    created_at: event.created_at ?? null
  };
}

function normalizeShipment(shipment: Partial<OrderShipment>): OrderShipment {
  return {
    shipment_id: shipment.shipment_id?.trim() || "unknown_shipment",
    seller_id: shipment.seller_id?.trim() || undefined,
    status: shipment.status?.trim() || "unknown",
    carrier: shipment.carrier?.trim() || undefined,
    tracking_number: shipment.tracking_number?.trim() || undefined,
    shipped_at: shipment.shipped_at ?? null,
    delivered_at: shipment.delivered_at ?? null
  };
}

function normalizeDispute(dispute: Partial<OrderDispute>): OrderDispute {
  return {
    dispute_id: dispute.dispute_id?.trim() || "unknown_dispute",
    order_id: dispute.order_id?.trim() || "unknown_order",
    type: dispute.type?.trim() || "status_mismatch",
    status: dispute.status?.trim() || "open",
    opened_by: dispute.opened_by?.trim() || "system",
    summary: dispute.summary?.trim() || "No dispute summary provided.",
    created_at: dispute.created_at ?? null
  };
}

export function toAdminOrdersQueryString(filters: AdminOrderFilters): string {
  return buildQueryString({
    q: filters.q?.trim(),
    status: filters.status && filters.status !== "all" ? filters.status : undefined,
    review_status:
      filters.review_status && filters.review_status !== "all" ? filters.review_status : undefined,
    user_id: filters.user_id?.trim(),
    seller_id: filters.seller_id?.trim(),
    from: filters.from,
    to: filters.to,
    page: filters.page,
    page_size: filters.limit
  });
}

export async function listAdminOrders(filters: AdminOrderFilters): Promise<AdminOrderListResponse> {
  const query = toAdminOrdersQueryString(filters);
  const response = await apiFetch<AdminOrderListResponse>(`${ORDERS_PATH}?${query}`);

  return {
    ...response,
    orders: (response.orders ?? []).map((order) => normalizeOrder(order)),
    page: response.page ?? filters.page,
    limit: response.page_size ?? response.limit ?? filters.limit
  };
}

export async function getAdminOrder(orderId: string): Promise<AdminOrderDetailResponse> {
  const response = await apiFetch<RawOrderDetailResponse>(
    `${ORDERS_PATH}/${encodeURIComponent(orderId)}`
  );
  const order = normalizeOrder(response.order ?? { order_id: orderId });

  return {
    ...response,
    order,
    status_history: (response.status_history ?? []).map(normalizeStatusEvent),
    shipments: (response.shipments ?? []).map(normalizeShipment),
    payment: normalizePaymentSummary(response.payment ?? order.payment)
  };
}

export async function listOrderDisputes(orderId: string): Promise<OrderDisputeListResponse> {
  const response = await apiFetch<OrderDisputeListResponse>(
    `${ORDERS_PATH}/${encodeURIComponent(orderId)}/disputes`
  );

  return {
    ...response,
    disputes: (response.disputes ?? []).map((dispute) =>
      normalizeDispute({ ...dispute, order_id: dispute.order_id || orderId })
    )
  };
}

export async function submitOrderReview(
  orderId: string,
  input: ManualReviewDecision
): Promise<SuccessResponse> {
  return apiFetch<SuccessResponse>(`${ORDERS_PATH}/${encodeURIComponent(orderId)}/review`, {
    method: "POST",
    body: JSON.stringify({
      decision: input.decision,
      reason: input.reason,
      internal_note: input.internal_note
    })
  });
}
