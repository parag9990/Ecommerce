import type {
  LocationGranularity,
  MaskingMode
} from "../../../api/session-api";

export function maskIdentifier(value: string | null | undefined): string {
  const trimmed = value?.trim() ?? "";
  if (!trimmed) {
    return "-";
  }
  if (trimmed.length <= 8) {
    return "****";
  }

  return `${trimmed.slice(0, 6)}...${trimmed.slice(-4)}`;
}

export function formatByMaskingMode(
  value: string | null | undefined,
  mode: MaskingMode
): string {
  if (mode === "hidden") {
    return "Hidden";
  }
  if (mode === "masked") {
    return maskIdentifier(value);
  }

  return value?.trim() || "-";
}

export function formatLocationByGranularity(
  geo: { city?: string; country?: string; region?: string },
  granularity: LocationGranularity
): string {
  if (granularity === "none") {
    return "Hidden";
  }
  if (granularity === "country") {
    return geo.country || "-";
  }

  return [geo.city, geo.region, geo.country].filter(Boolean).join(", ") || "-";
}

export function maskingModeLabel(mode: MaskingMode): string {
  const labels: Record<MaskingMode, string> = {
    full: "Full",
    hidden: "Hidden",
    masked: "Masked"
  };
  return labels[mode];
}

export function locationGranularityLabel(
  granularity: LocationGranularity
): string {
  const labels: Record<LocationGranularity, string> = {
    city: "City",
    country: "Country only",
    none: "Hidden"
  };
  return labels[granularity];
}
