import { http } from "../../../lib/http";
import type {
  FulfillmentUpdateInput,
  Order,
  OrderListResponse,
  SellerOrderFilters,
} from "../types";
import { normalizeOrder } from "../utils/order-mappers";

type RawOrderListResponse = {
  orders?: unknown[];
  total?: number;
};

function toQuery(params: Record<string, string | number | undefined>) {
  const search = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      search.set(key, String(value));
    }
  });

  return search.toString();
}

export async function listSellerOrders(
  filters: SellerOrderFilters,
): Promise<OrderListResponse> {
  const query = toQuery({
    status: filters.status === "all" ? undefined : filters.status,
    q: filters.q?.trim(),
    page: filters.page,
    page_size: filters.page_size,
    date_from: filters.date_from,
    date_to: filters.date_to,
  });
  const response = await http<RawOrderListResponse>(
    `/api/v1/seller/orders${query ? `?${query}` : ""}`,
  );
  const orders = Array.isArray(response.orders)
    ? response.orders.map(normalizeOrder)
    : [];

  return {
    orders,
    total: Number(response.total ?? orders.length),
  };
}

export async function updateOrderFulfillment(
  orderId: string,
  input: FulfillmentUpdateInput,
): Promise<Order> {
  const order = await http<unknown>(`/api/v1/seller/orders/${orderId}/fulfillment`, {
    method: "PATCH",
    body: JSON.stringify({
      order_id: orderId,
      ...input,
    }),
  });

  return normalizeOrder(order);
}
