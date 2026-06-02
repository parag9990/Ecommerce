import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { UserProfilePanel } from "../../src/features/users/components/user-profile-panel";
import type { AdminUser } from "../../src/features/users/types";
import { useAuthStore } from "../../src/stores/auth-store";

const activeUser: AdminUser = {
  user_id: "user_123",
  email: "rahul@example.com",
  phone: "+919999999999",
  full_name: "Rahul Sharma",
  status: "active",
  roles: ["buyer"]
};

function setAdminSession(roles: string[]) {
  useAuthStore.getState().setSession({
    accessToken: "admin-token",
    user: {
      id: "admin-1",
      email: "admin@example.com",
      name: "Admin User",
      roles
    }
  });
}

function renderWithQueryClient(ui: ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false }
    }
  });

  render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>);
}

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("UserProfilePanel status action", () => {
  it("hides block and unblock actions for readonly admins", () => {
    setAdminSession(["readonly_admin"]);

    renderWithQueryClient(<UserProfilePanel user={activeUser} />);

    expect(screen.queryByRole("button", { name: /block user/i })).toBeNull();
  });

  it("requires an audit reason before submitting a block action", async () => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.test");
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      new Response(JSON.stringify({ success: true }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      })
    );
    vi.stubGlobal("fetch", fetchMock);
    setAdminSession(["superadmin"]);
    const user = userEvent.setup();

    renderWithQueryClient(<UserProfilePanel user={activeUser} />);

    await user.click(screen.getByRole("button", { name: /block user/i }));

    const confirmButton = screen.getByRole("button", { name: /confirm/i });
    expect((confirmButton as HTMLButtonElement).disabled).toBe(true);

    await user.type(screen.getByLabelText(/audit reason/i), "short");
    expect((confirmButton as HTMLButtonElement).disabled).toBe(true);

    await user.clear(screen.getByLabelText(/audit reason/i));
    await user.type(screen.getByLabelText(/audit reason/i), "Repeated suspicious checkout attempts");
    expect((confirmButton as HTMLButtonElement).disabled).toBe(false);

    await user.click(confirmButton);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));

    const fetchCall = fetchMock.mock.calls[0];

    if (!fetchCall) {
      throw new Error("Expected fetch to be called.");
    }

    const [input, requestInit = {}] = fetchCall;
    const requestUrl = new URL(String(input));

    expect(requestUrl.pathname).toBe("/api/v1/admin/users/user_123/status");
    expect(JSON.parse(requestInit.body as string)).toEqual({
      status: "blocked",
      reason: "Repeated suspicious checkout attempts"
    });
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });
});
