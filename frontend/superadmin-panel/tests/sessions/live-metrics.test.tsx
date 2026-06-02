import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { LiveMetricsCards } from "../../src/features/sessions/components/live-metrics-cards";

describe("LiveMetricsCards", () => {
  it("renders loading and loaded live traffic values", () => {
    const { rerender } = render(<LiveMetricsCards isLoading metrics={undefined} />);

    expect(screen.getAllByText("Loading")).toHaveLength(3);

    rerender(
      <LiveMetricsCards
        isLoading={false}
        metrics={{
          active_users: 1200,
          active_sessions: 1420,
          events_per_minute: 932
        }}
      />
    );

    expect(screen.getByText("1,200")).toBeTruthy();
    expect(screen.getByText("1,420")).toBeTruthy();
    expect(screen.getByText("932")).toBeTruthy();
  });
});
