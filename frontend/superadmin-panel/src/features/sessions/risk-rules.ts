import type { AdminSession, SessionEvent, SessionRisk, SessionRiskLevel } from "./types";

const VALID_RISK_LEVELS = new Set<SessionRiskLevel>(["low", "medium", "high"]);

function clampScore(score: number): number {
  if (!Number.isFinite(score)) {
    return 0;
  }

  return Math.min(Math.max(Math.round(score), 0), 100);
}

export function riskLevelForScore(score: number): SessionRiskLevel {
  if (score >= 70) {
    return "high";
  }

  if (score >= 40) {
    return "medium";
  }

  return "low";
}

function hasKnownDevice(session: AdminSession): boolean {
  const deviceType = session.device?.device_type?.toLowerCase();
  const browser = session.device?.browser?.trim();
  const operatingSystem = session.device?.operating_system?.trim() || session.device?.os?.trim();

  return Boolean(deviceType && deviceType !== "unknown" && (browser || operatingSystem));
}

function eventType(event: SessionEvent): string {
  return event.event_type.toLowerCase();
}

function propertyString(event: SessionEvent, key: string): string {
  const value = event.properties?.[key];

  return typeof value === "string" ? value.toLowerCase() : "";
}

function isPaymentFailure(event: SessionEvent): boolean {
  const type = eventType(event);
  const status = propertyString(event, "status") || propertyString(event, "payment_status");
  const result = propertyString(event, "result");

  return (
    type.includes("payment_failed") ||
    (type.includes("payment") && (status.includes("fail") || result.includes("fail")))
  );
}

function isCheckoutEvent(event: SessionEvent): boolean {
  return eventType(event).includes("checkout");
}

function isCartEvent(event: SessionEvent): boolean {
  return eventType(event).includes("cart");
}

function eventRatePerMinute(events: SessionEvent[]): number {
  if (events.length < 2) {
    return 0;
  }

  const timestamps = events
    .map((event) => new Date(event.occurred_at).getTime())
    .filter((timestamp) => !Number.isNaN(timestamp))
    .sort((left, right) => left - right);

  if (timestamps.length < 2) {
    return 0;
  }

  const minutes = Math.max((timestamps[timestamps.length - 1] - timestamps[0]) / 60_000, 1);

  return events.length / minutes;
}

function normalizedFlags(flags?: string[] | null): string[] {
  return (flags ?? [])
    .map((flag) => flag.replace(/_/g, " ").trim())
    .filter((flag) => flag.length > 0);
}

export function calculateSessionRisk(
  session: AdminSession,
  events: SessionEvent[] = []
): SessionRisk {
  if (typeof session.risk_score === "number" && session.risk_level && VALID_RISK_LEVELS.has(session.risk_level)) {
    const score = clampScore(session.risk_score);

    return {
      score,
      level: session.risk_level,
      reasons: normalizedFlags(session.risk_flags)
    };
  }

  let score = 0;
  const reasons: string[] = [];
  const checkoutEvents = events.filter(isCheckoutEvent).length + (session.checkout_step_count ?? 0);
  const failedPayments = events.filter(isPaymentFailure).length + (session.payment_failure_count ?? 0);
  const cartEvents = events.filter(isCartEvent).length + (session.cart_action_count ?? 0);
  const eventsPerMinute = Math.max(eventRatePerMinute(events), session.events_per_minute ?? 0);

  if (!hasKnownDevice(session)) {
    score += 15;
    reasons.push("Unknown device metadata");
  }

  if (eventsPerMinute >= 120) {
    score += 45;
    reasons.push("Very high event rate");
  } else if (eventsPerMinute >= 60) {
    score += 30;
    reasons.push("High event rate");
  }

  if (cartEvents >= 10) {
    score += 20;
    reasons.push("High cart activity in one session");
  }

  if (checkoutEvents >= 5) {
    score += 20;
    reasons.push("Repeated checkout steps");
  }

  if (failedPayments >= 3) {
    score += 35;
    reasons.push("Repeated failed payment attempts");
  }

  if (!session.user_id && checkoutEvents > 0) {
    score += 10;
    reasons.push("Anonymous session reached checkout");
  }

  if ((session.active_session_count ?? 0) >= 4) {
    score += 20;
    reasons.push("Multiple active sessions for same account");
  }

  const cappedScore = clampScore(score);

  return {
    score: cappedScore,
    level: riskLevelForScore(cappedScore),
    reasons
  };
}
