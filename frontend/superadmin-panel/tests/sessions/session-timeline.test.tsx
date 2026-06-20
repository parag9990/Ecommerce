import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { SessionEventTimeline } from "../../src/features/sessions/components/session-event-timeline";

describe("SessionEventTimeline", () => {
  it("renders safe journey context without sensitive event properties", () => {
    render(
      <SessionEventTimeline
        events={[
          {
            event_type: "search",
            anonymous_id: "anon_123",
            session_id: "sess_123",
            occurred_at: "2026-06-01T10:00:00Z",
            path: "/search?q=private-term",
            properties: {
              query: "private-term",
              product_id: "prod_sensitive_1234567890",
              card_token: "tok_should_not_render"
            }
          },
          {
            event_type: "payment_result",
            anonymous_id: "anon_123",
            session_id: "sess_123",
            occurred_at: "2026-06-01T10:05:00Z",
            properties: {
              status: "failed",
              failure_code: "insufficient_funds"
            }
          }
        ]}
      />
    );

    expect(screen.getByText("Search")).toBeTruthy();
    expect(screen.getByText(/Path: \/search/)).toBeTruthy();
    expect(screen.getByText(/Product: prod\.\.\.7890/)).toBeTruthy();
    expect(screen.getByText("Payment Result")).toBeTruthy();
    expect(screen.getByText(/Status: failed/)).toBeTruthy();
    expect(screen.queryByText(/private-term/)).toBeNull();
    expect(screen.queryByText(/tok_should_not_render/)).toBeNull();
  });
});
