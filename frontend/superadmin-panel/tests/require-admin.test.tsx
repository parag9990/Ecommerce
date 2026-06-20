import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { RequireAdmin, RequireAdminRoles } from "../src/routes/require-admin";
import { useAuthStore } from "../src/stores/auth-store";

function renderProtectedRoute(initialPath = "/admin") {
  const router = createMemoryRouter(
    [
      {
        element: <RequireAdmin />,
        children: [
          {
            path: "/admin",
            element: <div>Protected Admin Shell</div>
          }
        ]
      },
      {
        path: "/login",
        element: <div>Admin Login</div>
      }
    ],
    { initialEntries: [initialPath] }
  );

  render(<RouterProvider router={router} />);
}

describe("RequireAdmin", () => {
  it("redirects unauthenticated users to login", async () => {
    useAuthStore.setState({ isHydrated: true });

    renderProtectedRoute();

    expect(await screen.findByText("Admin Login")).toBeTruthy();
  });

  it("blocks authenticated non-admin users", async () => {
    useAuthStore.getState().setSession({
      accessToken: "test-token",
      user: {
        id: "buyer-1",
        email: "buyer@example.com",
        name: "Buyer User",
        roles: ["buyer"]
      }
    });

    renderProtectedRoute();

    expect(await screen.findByText("Permission denied")).toBeTruthy();
  });

  it("renders protected shell for admin users", async () => {
    useAuthStore.getState().setSession({
      accessToken: "test-token",
      user: {
        id: "admin-1",
        email: "admin@example.com",
        name: "Admin User",
        roles: ["superadmin"]
      }
    });

    renderProtectedRoute();

    expect(await screen.findByText("Protected Admin Shell")).toBeTruthy();
  });
});

describe("RequireAdminRoles", () => {
  it("denies direct module access when the role is not allowed", async () => {
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
      <RequireAdminRoles roles={["finance_admin"]}>
        <div>Payments Module</div>
      </RequireAdminRoles>
    );

    expect(await screen.findByText("Permission denied")).toBeTruthy();
    expect(screen.queryByText("Payments Module")).toBeNull();
  });

  it("allows direct module access when the role is allowed", () => {
    useAuthStore.getState().setSession({
      accessToken: "test-token",
      user: {
        id: "admin-1",
        email: "admin@example.com",
        name: "Admin User",
        roles: ["finance_admin"]
      }
    });

    render(
      <RequireAdminRoles roles={["finance_admin"]}>
        <div>Payments Module</div>
      </RequireAdminRoles>
    );

    expect(screen.getByText("Payments Module")).toBeTruthy();
  });
});
