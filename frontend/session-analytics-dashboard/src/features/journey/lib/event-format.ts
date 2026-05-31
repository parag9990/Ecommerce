import {
  CreditCard,
  Eye,
  MousePointerClick,
  PackageSearch,
  ScrollText,
  Search,
  ShoppingCart,
  Split,
  type LucideIcon
} from "lucide-react";

import type { JourneyEvent, JourneyEventType } from "../../../api/session-api";
import { sanitizeDisplayText } from "./event-privacy";

const eventLabels: Record<JourneyEventType, string> = {
  add_to_cart: "Add to Cart",
  checkout_step: "Checkout",
  click: "Click",
  page_view: "Page View",
  payment_result: "Payment",
  product_view: "Product View",
  scroll: "Scroll",
  search: "Search"
};

const eventIcons: Record<JourneyEventType, LucideIcon> = {
  add_to_cart: ShoppingCart,
  checkout_step: Split,
  click: MousePointerClick,
  page_view: Eye,
  payment_result: CreditCard,
  product_view: PackageSearch,
  scroll: ScrollText,
  search: Search
};

const eventTones: Record<JourneyEventType, string> = {
  add_to_cart: "border-emerald-200 bg-emerald-50 text-emerald-700",
  checkout_step: "border-orange-200 bg-orange-50 text-orange-700",
  click: "border-cyan-200 bg-cyan-50 text-cyan-700",
  page_view: "border-blue-200 bg-blue-50 text-blue-700",
  payment_result: "border-rose-200 bg-rose-50 text-rose-700",
  product_view: "border-violet-200 bg-violet-50 text-violet-700",
  scroll: "border-zinc-200 bg-zinc-50 text-zinc-700",
  search: "border-amber-200 bg-amber-50 text-amber-700"
};

export function getEventLabel(type: JourneyEventType): string {
  return eventLabels[type];
}

export function getEventIcon(type: JourneyEventType): LucideIcon {
  return eventIcons[type];
}

export function getEventTone(type: JourneyEventType): string {
  return eventTones[type];
}

export function describeEvent(event: JourneyEvent): string {
  if (event.eventType === "page_view") {
    return `Visited ${sanitizeDisplayText(event.path)}`;
  }

  if (event.eventType === "product_view") {
    return `Viewed product ${formatProperty(event.properties.product_id)}`;
  }

  if (event.eventType === "search") {
    return `Searched ${formatProperty(event.properties.query)}`;
  }

  if (event.eventType === "click") {
    return `Clicked ${formatProperty(
      event.properties.element_id ?? event.properties.element
    )}`;
  }

  if (event.eventType === "scroll") {
    return `Scrolled to ${formatProperty(event.properties.depth_percent)}%`;
  }

  if (event.eventType === "add_to_cart") {
    return `Added product ${formatProperty(event.properties.product_id)} to cart`;
  }

  if (event.eventType === "checkout_step") {
    return `Checkout step ${formatProperty(
      event.properties.step_name ?? event.properties.step
    )}`;
  }

  if (event.eventType === "payment_result") {
    return `Payment ${formatProperty(event.properties.status)}`;
  }

  return `${getEventLabel(event.eventType)} on ${sanitizeDisplayText(event.path)}`;
}

export function summarizeEventProperties(event: JourneyEvent): string[] {
  const keysByType: Partial<Record<JourneyEventType, string[]>> = {
    add_to_cart: ["product_id", "variant_id", "quantity"],
    checkout_step: ["step_name", "order_id"],
    click: ["element_id", "x", "y"],
    page_view: ["title", "referrer"],
    payment_result: ["status", "order_id"],
    product_view: ["product_id", "category_id", "seller_id"],
    scroll: ["depth_percent"],
    search: ["query", "result_count"]
  };

  return (keysByType[event.eventType] ?? [])
    .filter((key) => event.properties[key] !== undefined)
    .map((key) => `${key}: ${formatProperty(event.properties[key])}`);
}

function formatProperty(value: unknown): string {
  if (value === undefined || value === null || value === "") {
    return "unknown";
  }

  if (typeof value === "string") {
    return sanitizeDisplayString(value);
  }

  if (typeof value === "number") {
    return String(value);
  }

  if (typeof value === "boolean") {
    return value ? "true" : "false";
  }

  return "available";
}

function sanitizeDisplayString(value: string): string {
  const sanitized = sanitizeDisplayText(value);

  return sanitized.length > 80 ? `${sanitized.slice(0, 77)}...` : sanitized;
}
