export function maskPhone(phone?: string | null): string {
  const digits = phone?.replace(/\D/g, "") ?? "";

  if (digits.length < 4) {
    return "Not added";
  }

  return `${"*".repeat(Math.max(digits.length - 4, 0))}${digits.slice(-4)}`;
}

export function formatDateTime(value?: string | null): string {
  if (!value) {
    return "Unknown";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "Unknown";
  }

  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(date);
}

export function formatMoney(
  money?: { amount?: number | null; currency?: string | null } | null
): string {
  if (!money || typeof money.amount !== "number") {
    return "Unknown";
  }

  return new Intl.NumberFormat("en", {
    style: "currency",
    currency: money.currency?.trim() || "INR",
    maximumFractionDigits: 2
  }).format(money.amount / 100);
}

export function formatSessionDuration(startedAt?: string | null, lastSeenAt?: string | null): string {
  if (!startedAt || !lastSeenAt) {
    return "Unknown";
  }

  const started = new Date(startedAt).getTime();
  const lastSeen = new Date(lastSeenAt).getTime();

  if (Number.isNaN(started) || Number.isNaN(lastSeen) || lastSeen < started) {
    return "Unknown";
  }

  const totalSeconds = Math.round((lastSeen - started) / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;

  if (minutes >= 60) {
    const hours = Math.floor(minutes / 60);
    const remainingMinutes = minutes % 60;

    return remainingMinutes > 0 ? `${hours}h ${remainingMinutes}m` : `${hours}h`;
  }

  return minutes > 0 ? `${minutes}m ${seconds}s` : `${seconds}s`;
}
