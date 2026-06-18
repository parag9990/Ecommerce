import { apiGet, apiPost } from '../../../lib/http';
import type { CheckoutRequest, CheckoutResponse, Order } from '../types';

export function createCheckout(body: CheckoutRequest) {
  return apiPost<CheckoutResponse, CheckoutRequest>(
    '/api/v1/orders/checkout',
    body,
    { auth: true },
  );
}

export function getOrder(orderId: string, signal?: AbortSignal) {
  return apiGet<Order>(`/api/v1/orders/${encodeURIComponent(orderId)}`, {
    auth: true,
    signal,
  });
}
