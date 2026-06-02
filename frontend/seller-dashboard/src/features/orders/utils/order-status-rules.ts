import type { OrderStatus } from "../types";

const editableStatuses = new Set<OrderStatus>(["paid", "packed", "shipped"]);

export function canUpdateFulfillment(status: OrderStatus) {
  return editableStatuses.has(status);
}

export function nextFulfillmentStatuses(status: OrderStatus): OrderStatus[] {
  if (status === "paid") {
    return ["packed", "shipped"];
  }

  if (status === "packed") {
    return ["shipped"];
  }

  if (status === "shipped") {
    return ["delivered"];
  }

  return [];
}
