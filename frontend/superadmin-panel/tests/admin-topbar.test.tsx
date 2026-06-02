import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { AdminTopbar } from "../src/components/layout/admin-topbar";
import { useAuthStore } from "../src/stores/auth-store";

describe("AdminTopbar", () => {
  it("shows admin identity and clears session on logout", () => {
    useAuthStore.getState().setSession({
      accessToken: "test-token",
      user: {
        id: "admin-1",
        email: "admin@example.com",
        name: "Admin User",
        roles: ["operations_admin"]
      }
    });

    render(
      <MemoryRouter>
        <AdminTopbar onOpenSidebar={() => undefined} />
      </MemoryRouter>
    );

    expect(screen.getByText("Admin User")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: /logout/i }));

    expect(useAuthStore.getState().accessToken).toBeNull();
    expect(useAuthStore.getState().user).toBeNull();
  });
});
