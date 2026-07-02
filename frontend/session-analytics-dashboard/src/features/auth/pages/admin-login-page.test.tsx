import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { clearAdminSession, readAdminSession } from "../../../lib/auth-session";
import { sendJSON } from "../../../lib/http";
import { AdminLoginPage } from "./admin-login-page";

vi.mock("../../../lib/http", async () => {
  const actual = await vi.importActual<typeof import("../../../lib/http")>(
    "../../../lib/http"
  );

  return {
    ...actual,
    sendJSON: vi.fn()
  };
});

const mockedSendJSON = vi.mocked(sendJSON);

function renderLoginPage() {
  render(
    <MemoryRouter initialEntries={["/login"]}>
      <Routes>
        <Route path="/login" element={<AdminLoginPage />} />
        <Route path="/" element={<div>Analytics Home</div>} />
      </Routes>
    </MemoryRouter>
  );
}

describe("AdminLoginPage", () => {
  beforeEach(() => {
    clearAdminSession();
    mockedSendJSON.mockReset();
  });

  afterEach(() => {
    clearAdminSession();
  });

  it("submits the auth login contract payload and stores an admin session", async () => {
    mockedSendJSON.mockResolvedValue({
      user: {
        user_id: "user_local_analytics_admin",
        roles: ["operations_admin"]
      },
      tokens: {
        access_token: "analytics-access-token",
        expires_in: 900
      }
    });

    renderLoginPage();

    await userEvent.type(
      screen.getByLabelText(/email or phone/i),
      "analytics.admin.local@example.com"
    );
    await userEvent.type(screen.getByLabelText(/password/i), "LocalDemo#2026!");
    await userEvent.click(screen.getByRole("button", { name: /continue/i }));

    await screen.findByText("Analytics Home");

    const [path, options] = mockedSendJSON.mock.calls[0] ?? [];
    const body = options?.body as Record<string, unknown>;
    const device = body.device as Record<string, unknown>;

    expect(path).toBe("/api/v1/auth/login");
    expect(body.identifier).toBe("analytics.admin.local@example.com");
    expect(body.password).toBe("LocalDemo#2026!");
    expect(device).toMatchObject({
      channel: "session-analytics-dashboard",
      user_agent: expect.any(String)
    });
    expect(device).not.toHaveProperty("source");

    await waitFor(() => {
      expect(readAdminSession()?.userId).toBe("user_local_analytics_admin");
    });
  });
});
