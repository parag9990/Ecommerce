import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { JourneyEvent } from "../../../api/session-api";
import { SessionTimeline } from "./session-timeline";

const events: JourneyEvent[] = [
  {
    eventId: "evt_1",
    eventType: "add_to_cart",
    occurredAt: "2026-05-18T00:06:30Z",
    path: "/products/prod_123",
    properties: { product_id: "prod_123", quantity: 1 }
  }
];

describe("SessionTimeline", () => {
  it("renders journey event labels and selects an event", async () => {
    const onSelectEvent = vi.fn();

    render(
      <SessionTimeline
        error={null}
        events={events}
        isLoading={false}
        onRetry={() => undefined}
        onSelectEvent={onSelectEvent}
      />
    );

    expect(screen.getByText("Add to Cart")).toBeInTheDocument();
    expect(screen.getAllByText(/prod_123/).length).toBeGreaterThan(0);

    await userEvent.click(screen.getByRole("button", { name: /add to cart/i }));

    expect(onSelectEvent).toHaveBeenCalledWith(events[0]);
  });

  it("renders an empty state when no events match", () => {
    render(
      <SessionTimeline
        error={null}
        events={[]}
        isLoading={false}
        onRetry={() => undefined}
        onSelectEvent={() => undefined}
      />
    );

    expect(screen.getByText("No journey events")).toBeInTheDocument();
  });
});
