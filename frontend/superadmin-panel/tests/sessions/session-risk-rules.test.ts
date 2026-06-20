import { describe, expect, it } from "vitest";

import { calculateSessionRisk, riskLevelForScore } from "../../src/features/sessions/risk-rules";
import type { AdminSession, SessionEvent } from "../../src/features/sessions/types";

const baseSession: AdminSession = {
  session_id: "sess_1234567890",
  anonymous_id: "anon_1234567890",
  started_at: "2026-06-01T10:00:00Z",
  last_seen_at: "2026-06-01T10:05:00Z",
  device: {
    browser: "Chrome",
    operating_system: "Windows",
    device_type: "desktop"
  }
};

function event(event_type: string, occurred_at: string, properties?: Record<string, unknown>): SessionEvent {
  return {
    event_type,
    anonymous_id: "anon_1234567890",
    session_id: "sess_1234567890",
    occurred_at,
    properties
  };
}

describe("session risk rules", () => {
  it("derives low, medium, and high levels from scores", () => {
    expect(riskLevelForScore(10)).toBe("low");
    expect(riskLevelForScore(40)).toBe("medium");
    expect(riskLevelForScore(70)).toBe("high");
  });

  it("uses backend risk fields when they are present", () => {
    const risk = calculateSessionRisk({
      ...baseSession,
      risk_score: 82,
      risk_level: "high",
      risk_flags: ["payment_abuse", "bot_like_rate"]
    });

    expect(risk).toEqual({
      score: 82,
      level: "high",
      reasons: ["payment abuse", "bot like rate"]
    });
  });

  it("adds unknown device and payment failure risk reasons", () => {
    const risk = calculateSessionRisk(
      {
        ...baseSession,
        device: {
          device_type: "unknown"
        }
      },
      [
        event("payment_result", "2026-06-01T10:01:00Z", { status: "failed" }),
        event("payment_result", "2026-06-01T10:02:00Z", { status: "failed" }),
        event("payment_result", "2026-06-01T10:03:00Z", { status: "failed" })
      ]
    );

    expect(risk.level).toBe("medium");
    expect(risk.reasons).toContain("Unknown device metadata");
    expect(risk.reasons).toContain("Repeated failed payment attempts");
  });

  it("marks dense anonymous checkout/cart journeys as high risk", () => {
    const events = [
      ...Array.from({ length: 10 }, (_value, index) =>
        event("add_to_cart", `2026-06-01T10:00:${String(index).padStart(2, "0")}Z`)
      ),
      ...Array.from({ length: 5 }, (_value, index) =>
        event("checkout_step", `2026-06-01T10:01:${String(index).padStart(2, "0")}Z`)
      ),
      ...Array.from({ length: 3 }, (_value, index) =>
        event("payment_result", `2026-06-01T10:02:${String(index).padStart(2, "0")}Z`, {
          result: "failed"
        })
      )
    ];

    const risk = calculateSessionRisk(baseSession, events);

    expect(risk.level).toBe("high");
    expect(risk.reasons).toContain("Anonymous session reached checkout");
  });
});
