import type { JourneyEventProperties } from "../../../api/session-api";

const blockedPropertyFragments = [
  "address",
  "card",
  "cvv",
  "email",
  "otp",
  "pin_code",
  "password",
  "phone",
  "secret",
  "token",
  "upi_id"
];

const blockedPropertyKeys = [
  "anonymous_id",
  "anonymousid",
  "card_number",
  "access_token",
  "full_address",
  "input_value",
  "ip",
  "ip_address",
  "private_text",
  "refresh_token",
  "raw_ip",
  "raw_text",
  "upi_id",
  "user_id",
  "userid"
];

const cardPattern = /\b\d{12,19}\b/g;
const emailPattern = /[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/gi;
const ipPattern = /\b(?:\d{1,3}\.){3}\d{1,3}\b/g;
const phonePattern = /\b(?:\+?\d[\d\s().-]{7,}\d)\b/g;

export function sanitizeEventProperties(
  properties: JourneyEventProperties
): JourneyEventProperties {
  return sanitizeRecord(properties);
}

export function sanitizeDisplayText(value: string): string {
  return value
    .replace(emailPattern, "[masked-email]")
    .replace(ipPattern, "[masked-ip]")
    .replace(phonePattern, "[masked-phone]")
    .replace(cardPattern, "[masked-card]");
}

function sanitizeRecord(record: Record<string, unknown>): JourneyEventProperties {
  return Object.fromEntries(
    Object.entries(record).map(([key, value]) => [
      key,
      isSensitiveKey(key) ? "[masked]" : sanitizeValue(value)
    ])
  );
}

function sanitizeValue(value: unknown): unknown {
  if (typeof value === "string") {
    return sanitizeDisplayText(value);
  }

  if (Array.isArray(value)) {
    return value.map(sanitizeValue);
  }

  if (isRecord(value)) {
    return sanitizeRecord(value);
  }

  return value;
}

function isSensitiveKey(key: string): boolean {
  const normalized = key.toLowerCase();

  return (
    blockedPropertyKeys.includes(normalized) ||
    blockedPropertyFragments.some((fragment) => normalized.includes(fragment))
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
