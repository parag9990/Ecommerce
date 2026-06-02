import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { SessionTable } from "../../src/features/sessions/components/session-table";
import type { AdminSession } from "../../src/features/sessions/types";

const sessions: AdminSession[] = [
  {
    session_id: "sess_sensitive_1234567890",
    anonymous_id: "anon_sensitive_9876543210",
    user_id: "user_sensitive_11112222",
    started_at: "2026-06-01T10:00:00Z",
    last_seen_at: "2026-06-01T10:30:00Z",
    status: "active",
    device: {
      browser: "Chrome",
      operating_system: "Windows",
      device_type: "desktop"
    }
  }
];

describe("SessionTable", () => {
  it("masks identifiers and emits journey selection", () => {
    const onOpen = vi.fn();

    render(
      <SessionTable
        sessions={sessions}
        getRisk={() => ({ score: 75, level: "high", reasons: ["Repeated failed payment attempts"] })}
        isLoading={false}
        isFetching={false}
        error={null}
        page={1}
        limit={25}
        totalCount={1}
        onPageChange={vi.fn()}
        onRetry={vi.fn()}
        onOpen={onOpen}
      />
    );

    expect(screen.queryByText("sess_sensitive_1234567890")).toBeNull();
    expect(screen.getByLabelText(/session id: sess\.\.\.7890/i)).toBeTruthy();
    expect(screen.getByText(/high risk/i)).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: /view journey/i }));

    expect(onOpen).toHaveBeenCalledWith("sess_sensitive_1234567890");
  });
});
