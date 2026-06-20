import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { SessionFilterBar } from "../../src/features/sessions/components/session-filter-bar";
import type { AdminSessionFilters } from "../../src/features/sessions/types";

const filters: AdminSessionFilters = {
  user_id: "",
  from: "",
  to: "",
  status: "all",
  risk_level: "all",
  page: 1,
  limit: 25
};

describe("SessionFilterBar", () => {
  it("emits user, status, risk, date, and reset changes", () => {
    const onChange = vi.fn();
    const onReset = vi.fn();

    render(<SessionFilterBar filters={filters} onChange={onChange} onReset={onReset} />);

    fireEvent.change(screen.getByPlaceholderText(/user id/i), {
      target: { value: "user_123" }
    });
    fireEvent.change(screen.getByLabelText(/session status/i), {
      target: { value: "revoked" }
    });
    fireEvent.change(screen.getByLabelText(/session risk level/i), {
      target: { value: "high" }
    });
    fireEvent.change(screen.getByLabelText(/session from date/i), {
      target: { value: "2026-06-01T10:00" }
    });
    fireEvent.change(screen.getByLabelText(/session to date/i), {
      target: { value: "2026-06-01T12:00" }
    });
    fireEvent.click(screen.getByRole("button", { name: /reset/i }));

    expect(onChange).toHaveBeenCalledWith({ user_id: "user_123" });
    expect(onChange).toHaveBeenCalledWith({ status: "revoked" });
    expect(onChange).toHaveBeenCalledWith({ risk_level: "high" });
    expect(onChange).toHaveBeenCalledWith({ from: "2026-06-01T10:00" });
    expect(onChange).toHaveBeenCalledWith({ to: "2026-06-01T12:00" });
    expect(onReset).toHaveBeenCalledTimes(1);
  });
});
