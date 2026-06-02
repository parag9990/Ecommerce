import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { FeatureFlagList } from "../../src/features/settings/components/feature-flag-list";
import type { PlatformSetting } from "../../src/features/settings/types";
import { useAuthStore } from "../../src/stores/auth-store";

const featureFlagSetting: PlatformSetting = {
  key: "platform.feature_flags",
  value: {
    flags: {
      new_checkout: false
    }
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
        key: "platform.feature_flags",
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

describe("FeatureFlagList", () => {
  it("submits toggled flags only after reason confirmation", async () => {
    const fetchMock = prepareMutationTest();
    const user = userEvent.setup();

    renderWithQueryClient(<FeatureFlagList setting={featureFlagSetting} canWrite />);

    await user.click(screen.getByLabelText(/toggle/i));
    await user.click(screen.getByRole("button", { name: /^save$/i }));

    const confirmButton = screen.getByRole("button", { name: /confirm update/i });
    expect((confirmButton as HTMLButtonElement).disabled).toBe(true);

    await user.type(screen.getByLabelText(/audit reason/i), "Enable checkout rollout safely");
    await user.click(confirmButton);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));

    const [, init = {}] = fetchMock.mock.calls[0] ?? [];

    expect(JSON.parse(init.body as string)).toEqual({
      value: {
        flags: {
          new_checkout: true
        }
      },
      reason: "Enable checkout rollout safely"
    });
  });
});
