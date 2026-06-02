import type { SellerStaffStatus } from "../types";

export const STAFF_STATUS_LABELS: Record<SellerStaffStatus, string> = {
  invited: "Invited",
  active: "Active",
  disabled: "Disabled",
};

export function formatDateTime(value: string) {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "Unknown";
  }

  return new Intl.DateTimeFormat("en-IN", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export function formatMemberName(fullName: string | null, email: string) {
  const trimmedName = fullName?.trim();

  return trimmedName || email;
}
