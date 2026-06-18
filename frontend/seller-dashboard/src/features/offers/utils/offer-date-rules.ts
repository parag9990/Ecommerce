const dateTimeLocalFormatter = new Intl.DateTimeFormat("sv-SE", {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
  hour12: false,
});

const shortDateFormatter = new Intl.DateTimeFormat("en-IN", {
  day: "2-digit",
  month: "short",
});

const monthFormatter = new Intl.DateTimeFormat("en-IN", {
  month: "long",
  year: "numeric",
});

export function parseDate(value?: string) {
  if (!value) {
    return null;
  }

  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

export function toDateTimeLocal(value?: string) {
  const date = parseDate(value);

  if (!date) {
    return "";
  }

  return dateTimeLocalFormatter.format(date).replace(" ", "T");
}

export function fromDateTimeLocal(value?: string) {
  const trimmed = value?.trim();

  if (!trimmed) {
    return undefined;
  }

  const date = parseDate(trimmed);
  return date ? date.toISOString() : undefined;
}

export function formatShortDate(value?: string) {
  const date = parseDate(value);
  return date ? shortDateFormatter.format(date) : "";
}

export function formatMonthLabel(date: Date) {
  return monthFormatter.format(date);
}

export function formatMonthParam(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");

  return `${year}-${month}`;
}

export function getMonthStart(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), 1);
}

export function getMonthEnd(date: Date) {
  return new Date(date.getFullYear(), date.getMonth() + 1, 0, 23, 59, 59, 999);
}

export function addMonths(date: Date, months: number) {
  return new Date(date.getFullYear(), date.getMonth() + months, 1);
}

export function getCalendarDays(month: Date) {
  const start = getMonthStart(month);
  const end = getMonthEnd(month);
  const leadingBlanks = start.getDay();
  const days: Array<Date | null> = Array.from({ length: leadingBlanks }, () => null);

  for (let day = 1; day <= end.getDate(); day += 1) {
    days.push(new Date(month.getFullYear(), month.getMonth(), day));
  }

  while (days.length % 7 !== 0) {
    days.push(null);
  }

  return days;
}

function startOfDay(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime();
}

export function dateFallsInsideRange(day: Date, startsAt: string, endsAt: string) {
  const start = parseDate(startsAt);
  const end = parseDate(endsAt);

  if (!start || !end) {
    return false;
  }

  const dayTime = startOfDay(day);
  return dayTime >= startOfDay(start) && dayTime <= startOfDay(end);
}

export function isExpired(endsAt?: string, now = new Date()) {
  const end = parseDate(endsAt);
  return Boolean(end && end.getTime() < now.getTime());
}
