import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { AdminSidebar } from "../src/components/layout/admin-sidebar";
import { useAuthStore } from "../src/stores/auth-store";

function renderSidebar(roles: string[]) {
  useAuthStore.getState().setSession({
    accessToken: "test-token",
    user: {
      id: "admin-1",
      email: "admin@example.com",
      name: "Admin User",
      roles
    }
  });

  render(
    <MemoryRouter initialEntries={["/admin"]}>
      <AdminSidebar isMobileOpen={false} onClose={() => undefined} />
    </MemoryRouter>
  );
}

describe("AdminSidebar", () => {
  it("renders only finance admin menu items", () => {
    renderSidebar(["finance_admin"]);

    expect(screen.getByText("Overview")).toBeTruthy();
    expect(screen.getByText("Payments")).toBeTruthy();
    expect(screen.queryByText("Audit Logs")).toBeNull();
    expect(screen.queryByText("Users")).toBeNull();
    expect(screen.queryByText("Platform Settings")).toBeNull();
  });

  it("renders all menu items for superadmin", () => {
    renderSidebar(["superadmin"]);

    expect(screen.getByText("Users")).toBeTruthy();
    expect(screen.getByText("Sellers")).toBeTruthy();
    expect(screen.getByText("Orders")).toBeTruthy();
    expect(screen.getByText("Payments")).toBeTruthy();
    expect(screen.getByText("Platform Settings")).toBeTruthy();
    expect(screen.getByText("Audit Logs")).toBeTruthy();
  });

  it("renders audit logs for readonly admins", () => {
    renderSidebar(["readonly_admin"]);

    expect(screen.getByText("Audit Logs")).toBeTruthy();
  });
});
