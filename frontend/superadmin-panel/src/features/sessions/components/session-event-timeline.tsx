import { Timeline, type TimelineItem } from "../../../components/ui/timeline";
import { formatDateTime } from "../../../lib/format";
import { maskIdentifier } from "./masked-identity";
import type { SessionEvent } from "../types";

const SAFE_PROPERTY_LABELS: Record<string, string> = {
  status: "Status",
  result: "Result",
  step: "Step",
  checkout_step: "Step",
  product_id: "Product",
  sku: "SKU",
  order_id: "Order",
  payment_status: "Payment",
  failure_code: "Failure",
  cart_items: "Cart items",
  item_count: "Items"
};

const SENSITIVE_PROPERTY_PATTERN =
  /(password|otp|token|secret|card|cvv|email|phone|address|name|query|text|message)/i;

function formatEventType(eventType: string): string {
  return eventType
    .replace(/[_-]/g, " ")
    .replace(/\s+/g, " ")
    .trim()
    .replace(/\b\w/g, (character) => character.toUpperCase());
}

function safePath(path?: string | null): string | null {
  if (!path) {
    return null;
  }

  const [withoutQuery] = path.split(/[?#]/);

  return withoutQuery || null;
}

function formatSafeProperty(key: string, value: unknown): string | null {
  if (SENSITIVE_PROPERTY_PATTERN.test(key) || !(key in SAFE_PROPERTY_LABELS)) {
    return null;
  }

  if (typeof value !== "string" && typeof value !== "number" && typeof value !== "boolean") {
    return null;
  }

  const normalizedValue = String(value);
  const displayValue = key.endsWith("_id") ? maskIdentifier(normalizedValue) : normalizedValue;

  return `${SAFE_PROPERTY_LABELS[key]}: ${displayValue}`;
}

function eventDescription(event: SessionEvent): string | null {
  const details = Object.entries(event.properties ?? {})
    .map(([key, value]) => formatSafeProperty(key, value))
    .filter((value): value is string => Boolean(value))
    .slice(0, 3);
  const path = safePath(event.path);

  if (path) {
    details.unshift(`Path: ${path}`);
  }

  return details.length > 0 ? details.join(" | ") : null;
}

export function SessionEventTimeline({ events }: { events: SessionEvent[] }) {
  const items: TimelineItem[] = events.map((event, index) => ({
    id: `${event.session_id}-${event.occurred_at}-${index}`,
    title: formatEventType(event.event_type),
    description: eventDescription(event),
    timestamp: formatDateTime(event.occurred_at)
  }));

  return <Timeline items={items} emptyLabel="No journey events found." />;
}
