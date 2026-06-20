import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { MaintenanceModePanel } from "../../src/features/settings/components/maintenance-mode-panel";
import type { PlatformSetting } from "../../src/features/settings/types";
import { useAuthStore } from "../../src/stores/auth-store";

const maintenanceSetting: PlatformSetting = {
  key: "platform.maintenance_mode",
  value: {
    enabled: false,
    message: "",
    starts_at: null,
    ends_at: null,
    allow_admin_bypass: true
  },
  updated_at: "2026-06-02T10:00:00Z"
};

function renderWithQueryClient(ui: ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false }
    }
  });

  render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>);
}

function prepareMutationTest() {
  vi.stubEnv("VITE_API_BASE_URL", "https://api.example.test");
  useAuthStore.getState().setSession({
    accessToken: "admin-token",
    user: {
      id: "admin-1",
      email: "admin@example.com",
      name: "Admin User",
      roles: ["superadmin"]
    }
  });

  const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) =>
    new Response(
      JSON.stringify({
        key: "platform.maintenance_mode",
        value: JSON.parse(String(init?.body)).value,
        updated_at: "2026-06-02T11:00:00Z"
      }),
      {
        status: 200,
        headers: { "Content-Type": "application/json" }
      }
    )
  );
  vi.stubGlobal("fetch", fetchMock);

  return fetchMock;
}

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("MaintenanceModePanel", () => {
  it("requires a valid message and audit reason before enabling maintenance", async () => {
    const fetchMock = prepareMutationTest();
    const user = userEvent.setup();

    renderWithQueryClient(<MaintenanceModePanel setting={maintenanceSetting} canWrite />);

    await user.click(screen.getByLabelText(/enable maintenance mode/i));

    expect((screen.getByRole("button", { name: /^save$/i }) as HTMLButtonElement).disabled).toBe(
      true
    );

    await user.type(
      screen.getByLabelText(/maintenance message/i),
      "Scheduled platform maintenance"
    );
    await user.click(screen.getByRole("button", { name: /^save$/i }));

    const confirmButton = screen.getByRole("button", { name: /confirm update/i });
    expect((confirmButton as HTMLButtonElement).disabled).toBe(true);

    await user.type(screen.getByLabelText(/audit reason/i), "Planned database maintenance window");
    expect((confirmButton as HTMLButtonElement).disabled).toBe(false);

    await user.click(confirmButton);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));

    const [input, init = {}] = fetchMock.mock.calls[0] ?? [];

    expect(new URL(String(input)).pathname).toBe(
      "/api/v1/admin/settings/platform.maintenance_mode"
    );
    expect(JSON.parse(init.body as string)).toMatchObject({
      reason: "Planned database maintenance window",
      value: {
        enabled: true,
        message: "Scheduled platform maintenance",
        allow_admin_bypass: true
      }
    });
  });
});
