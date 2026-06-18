import type {
  CustomerSummary,
  Money,
  Order,
  OrderItem,
  OrderStatus,
  Refund,
  Shipment,
  ShippingAddress,
  StatusHistoryEntry,
} from "../types";
import { ORDER_STATUSES } from "../types";

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};
}

function normalizeString(value: unknown) {
  return typeof value === "string" ? value : "";
}

function normalizeOptionalString(value: unknown) {
  return typeof value === "string" && value.length > 0 ? value : undefined;
}

function normalizeNumber(value: unknown) {
  const numberValue = Number(value);
  return Number.isFinite(numberValue) ? numberValue : 0;
}

function normalizeArray(value: unknown) {
  return Array.isArray(value) ? value : [];
}

export function isOrderStatus(value: unknown): value is OrderStatus {
  return ORDER_STATUSES.some((status) => status === value);
}

function normalizeOrderStatus(value: unknown): OrderStatus {
  return isOrderStatus(value) ? value : "created";
}

export function normalizeMoney(value: unknown): Money {
  const candidate = asRecord(value);

  return {
    amount: normalizeNumber(candidate.amount),
    currency: normalizeOptionalString(candidate.currency) ?? "INR",
  };
}

function normalizeOrderItem(value: unknown): OrderItem {
  const candidate = asRecord(value);

  return {
    product_id: normalizeOptionalString(candidate.product_id),
    variant_id: normalizeOptionalString(candidate.variant_id),
    seller_id: normalizeOptionalString(candidate.seller_id),
    title:
      normalizeOptionalString(candidate.title) ??
      normalizeOptionalString(candidate.product_title) ??
      normalizeOptionalString(candidate.name),
    sku: normalizeOptionalString(candidate.sku),
    quantity: normalizeNumber(candidate.quantity || 1),
    unit_price: candidate.unit_price ? normalizeMoney(candidate.unit_price) : undefined,
    total: candidate.total ? normalizeMoney(candidate.total) : undefined,
  };
}

function normalizeShipment(value: unknown): Shipment {
  const candidate = asRecord(value);

  return {
    shipment_id: normalizeOptionalString(candidate.shipment_id),
    status: candidate.status ? normalizeOrderStatus(candidate.status) : undefined,
    tracking_number: normalizeOptionalString(candidate.tracking_number),
    carrier: normalizeOptionalString(candidate.carrier),
    shipped_at: normalizeOptionalString(candidate.shipped_at),
    delivered_at: normalizeOptionalString(candidate.delivered_at),
    updated_at: normalizeOptionalString(candidate.updated_at),
  };
}

function normalizeRefund(value: unknown): Refund {
  const candidate = asRecord(value);

  return {
    refund_id: normalizeOptionalString(candidate.refund_id),
    payment_id: normalizeOptionalString(candidate.payment_id),
    status: normalizeOptionalString(candidate.status) ?? "unknown",
    amount: normalizeMoney(candidate.amount),
    reason: normalizeOptionalString(candidate.reason),
    created_at: normalizeOptionalString(candidate.created_at),
  };
}

function normalizeStatusHistory(value: unknown): StatusHistoryEntry {
  const candidate = asRecord(value);

  return {
    status: normalizeOrderStatus(candidate.status),
    created_at: normalizeOptionalString(candidate.created_at) ?? "",
    note: normalizeOptionalString(candidate.note),
  };
}

function normalizeCustomer(value: unknown, userId?: string): CustomerSummary | undefined {
  const candidate = asRecord(value);
  const customer: CustomerSummary = {
    user_id: normalizeOptionalString(candidate.user_id) ?? userId,
    name: normalizeOptionalString(candidate.name),
    email: normalizeOptionalString(candidate.email),
    phone: normalizeOptionalString(candidate.phone),
  };

  return Object.values(customer).some(Boolean) ? customer : undefined;
}

function normalizeShippingAddress(value: unknown): ShippingAddress | undefined {
  const candidate = asRecord(value);
  const address: ShippingAddress = {
    name: normalizeOptionalString(candidate.name),
    phone: normalizeOptionalString(candidate.phone),
    line1:
      normalizeOptionalString(candidate.line1) ??
      normalizeOptionalString(candidate.address_line1),
    line2:
      normalizeOptionalString(candidate.line2) ??
      normalizeOptionalString(candidate.address_line2),
    city: normalizeOptionalString(candidate.city),
    state: normalizeOptionalString(candidate.state),
    country: normalizeOptionalString(candidate.country),
    postal_code:
      normalizeOptionalString(candidate.postal_code) ??
      normalizeOptionalString(candidate.zipcode),
  };

  return Object.values(address).some(Boolean) ? address : undefined;
}

export function normalizeOrder(value: unknown): Order {
  if (!value || typeof value !== "object") {
    throw new Error("Order response was empty or invalid.");
  }

  const candidate = value as Record<string, unknown>;
  const userId = normalizeOptionalString(candidate.user_id);

  return {
    order_id: normalizeString(candidate.order_id),
    user_id: userId,
    status: normalizeOrderStatus(candidate.status),
    items: normalizeArray(candidate.items).map(normalizeOrderItem),
    total: normalizeMoney(candidate.total),
    created_at: normalizeOptionalString(candidate.created_at) ?? "",
    updated_at: normalizeOptionalString(candidate.updated_at),
    customer: normalizeCustomer(candidate.customer, userId),
    shipping_address: normalizeShippingAddress(
      candidate.shipping_address ?? candidate.shipping ?? candidate.delivery_address,
    ),
    shipments: normalizeArray(candidate.shipments).map(normalizeShipment),
    refunds: normalizeArray(candidate.refunds).map(normalizeRefund),
    status_history: normalizeArray(candidate.status_history).map(normalizeStatusHistory),
  };
}
