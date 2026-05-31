import type {
  ActiveSessionDevice,
  ActiveSessionLocation
} from "../api/session-api";

export function maskAnonymousId(value: string): string {
  return maskIdentifier(value, 6, 4);
}

export function maskUserId(value?: string): string {
  if (!value) {
    return "Anonymous";
  }

  if (value.includes("...") || value.includes("*")) {
    return value;
  }

  return maskIdentifier(value, 5, 4);
}

export function maskSessionId(value: string): string {
  return maskIdentifier(value, 6, 4);
}

export function formatDeviceLabel(device: ActiveSessionDevice): string {
  const details = [device.browser, device.os].filter(Boolean).join(" / ");
  return details || "Unknown platform";
}

export function formatLocation(location: ActiveSessionLocation): string {
  const parts = [location.city, location.region, location.country].filter(Boolean);
  return parts.length > 0 ? parts.join(", ") : "Unknown";
}

export function normalizeBreakdownWidth(percentage: number): string {
  const safePercentage = Number.isFinite(percentage) ? percentage : 0;
  return `${Math.min(100, Math.max(safePercentage, safePercentage > 0 ? 4 : 0))}%`;
}

function maskIdentifier(
  value: string,
  visiblePrefix: number,
  visibleSuffix: number
): string {
  const trimmed = value.trim();

  if (!trimmed) {
    return "Unknown";
  }

  if (trimmed.length <= visiblePrefix + visibleSuffix + 2) {
    return `${trimmed.slice(0, Math.min(4, trimmed.length))}****`;
  }

  return `${trimmed.slice(0, visiblePrefix)}...${trimmed.slice(-visibleSuffix)}`;
}
