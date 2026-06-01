import { apiGet, apiPost } from '../../../lib/http';
import type { CancelOrderRequest, Order, OrderListResponse } from '../types';

type ListOrdersParams = {
  page?: number | undefined;
  page_size?: number | undefined;
};

export function listOrders(
  params: ListOrdersParams = {},
  signal?: AbortSignal,
) {
  const search = new URLSearchParams();

  if (params.page) {
    search.set('page', String(params.page));
  }

  if (params.page_size) {
    search.set('page_size', String(params.page_size));
  }

  const query = search.toString();

  return apiGet<OrderListResponse>(`/api/v1/orders${query ? `?${query}` : ''}`, {
    auth: true,
    signal,
  });
}

export function getOrder(orderId: string, signal?: AbortSignal) {
  return apiGet<Order>(`/api/v1/orders/${encodeURIComponent(orderId)}`, {
    auth: true,
    signal,
  });
}

export function cancelOrder(orderId: string, reason: string) {
  return apiPost<Order, CancelOrderRequest>(
    `/api/v1/orders/${encodeURIComponent(orderId)}/cancel`,
    { reason },
    { auth: true },
  );
}
