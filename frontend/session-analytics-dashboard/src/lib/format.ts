export function formatNumber(value: number): string {
  return new Intl.NumberFormat("en-US").format(guardFiniteNumber(value));
}

export function formatPercent(value: number): string {
  return `${guardFiniteNumber(value).toFixed(1)}%`;
}

export function formatDuration(seconds: number): string {
  const safeSeconds = Math.max(0, Math.round(guardFiniteNumber(seconds)));
  const minutes = Math.floor(safeSeconds / 60);
  const remainingSeconds = safeSeconds % 60;

  if (minutes >= 60) {
    const hours = Math.floor(minutes / 60);
    const remainingMinutes = minutes % 60;
    return `${hours}h ${remainingMinutes}m`;
  }

  return `${minutes}m ${remainingSeconds}s`;
}

export function formatRelativeTime(value: string, now: Date = new Date()): string {
  const date = new Date(value);
  const timestamp = date.getTime();

  if (Number.isNaN(timestamp)) {
    return "Unknown";
  }

  const diffSeconds = Math.max(
    0,
    Math.floor((now.getTime() - timestamp) / 1000)
  );

  if (diffSeconds < 60) {
    return `${diffSeconds}s ago`;
  }

  const diffMinutes = Math.floor(diffSeconds / 60);
  if (diffMinutes < 60) {
    return `${diffMinutes}m ago`;
  }

  const diffHours = Math.floor(diffMinutes / 60);
  if (diffHours < 24) {
    return `${diffHours}h ago`;
  }

  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays}d ago`;
}

export function formatDateTime(value: string): string {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "Unknown";
  }

  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "medium"
  }).format(date);
}

export function formatTime(value: string): string {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "Unknown";
  }

  return new Intl.DateTimeFormat("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit"
  }).format(date);
}

function guardFiniteNumber(value: number): number {
  return Number.isFinite(value) ? value : 0;
}
