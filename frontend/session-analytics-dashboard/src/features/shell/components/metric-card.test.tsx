import { Activity } from "lucide-react";
import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

import { MetricCard } from "./metric-card";

describe("MetricCard", () => {
  it("renders the metric label, value, and helper", () => {
    render(
      <MetricCard
        helper="Currently active sessions"
        icon={Activity}
        label="Active Users Now"
        value="42"
      />
    );

    expect(screen.getByText("Active Users Now")).toBeInTheDocument();
    expect(screen.getByText("42")).toBeInTheDocument();
    expect(screen.getByText("Currently active sessions")).toBeInTheDocument();
  });
});
