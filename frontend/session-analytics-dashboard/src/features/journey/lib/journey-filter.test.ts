import { describe, expect, it } from "vitest";

import type { JourneyEvent } from "../../../api/session-api";
import { filterJourneyEvents } from "./journey-filter";

const events: JourneyEvent[] = [
  {
    eventId: "evt_1",
    eventType: "page_view",
    occurredAt: "2026-05-18T00:00:00Z",
    path: "/",
    properties: {}
  },
  {
    eventId: "evt_2",
    eventType: "add_to_cart",
    occurredAt: "2026-05-18T00:05:00Z",
    path: "/products/prod_123",
    properties: { product_id: "prod_123" }
  }
];

describe("filterJourneyEvents", () => {
  it("returns all events when no types are selected", () => {
    expect(filterJourneyEvents(events, [])).toEqual(events);
  });

  it("filters events by selected type", () => {
    expect(filterJourneyEvents(events, ["add_to_cart"])).toEqual([events[1]]);
  });
});
