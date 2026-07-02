import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { AdminLoginPage } from "../src/features/auth/pages/admin-login-page";
import { useAuthStore } from "../src/stores/auth-store";

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

function renderLoginPage() {
  useAuthStore.setState({ isHydrated: true });

  const router = createMemoryRouter(
    [
      { path: "/login", element: <AdminLoginPage /> },
      { path: "/admin", element: <div>Admin Home</div> }
    ],
    { initialEntries: ["/login"] }
  );

  render(<RouterProvider router={router} />);
}

describe("AdminLoginPage", () => {
  it("submits the auth login contract payload and stores a superadmin session", async () => {
    vi.stubEnv("VITE_API_BASE_URL", "https://gateway.example.test");
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      new Response(
        JSON.stringify({
          user: {
            user_id: "user_local_superadmin",
            roles: ["superadmin"]
          },
          tokens: {
            access_token: "access-token",
            refresh_token: "refresh-token",
            expires_in: 900
          },
          session_id: "sess_123"
        }),
        {
          status: 200,
          headers: { "Content-Type": "application/json" }
        }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    renderLoginPage();

    await userEvent.type(screen.getByLabelText(/email or phone/i), "superadmin.local@example.com");
    await userEvent.type(screen.getByLabelText(/password/i), "LocalDemo#2026!");
    await userEvent.click(screen.getByRole("button", { name: /continue/i }));

    await screen.findByText("Admin Home");

    const [, init] = fetchMock.mock.calls[0] ?? [];
    const body = JSON.parse(String(init?.body));

    expect(body.identifier).toBe("superadmin.local@example.com");
    expect(body.password).toBe("LocalDemo#2026!");
    expect(body.device).toMatchObject({
      channel: "superadmin-panel"
    });
    expect(body.device).not.toHaveProperty("source");

    await waitFor(() => {
      expect(useAuthStore.getState().user?.roles).toEqual(["superadmin"]);
      expect(useAuthStore.getState().accessToken).toBe("access-token");
    });
  });
});
