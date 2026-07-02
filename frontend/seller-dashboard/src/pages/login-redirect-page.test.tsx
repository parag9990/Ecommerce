import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { http } from "../lib/http";
import { clearSellerAuthSession, readSellerAuthSession } from "../lib/auth-session";
import { LoginRedirectPage } from "./login-redirect-page";

vi.mock("../lib/http", async () => {
  const actual = await vi.importActual<typeof import("../lib/http")>("../lib/http");

  return {
    ...actual,
    http: vi.fn(),
  };
});

const mockedHttp = vi.mocked(http);

function renderLoginPage() {
  const router = createMemoryRouter(
    [
      { path: "/login", element: <LoginRedirectPage /> },
      { path: "/seller", element: <div>Seller Home</div> },
    ],
    { initialEntries: ["/login"] },
  );

  render(<RouterProvider router={router} />);
}

describe("LoginRedirectPage", () => {
  beforeEach(() => {
    clearSellerAuthSession();
    mockedHttp.mockReset();
  });

  afterEach(() => {
    clearSellerAuthSession();
  });

  it("submits the auth login contract payload and stores a seller session", async () => {
    mockedHttp.mockResolvedValue({
      user: {
        user_id: "user_local_seller",
        roles: ["seller"],
        seller_id: "seller_local_demo",
      },
      tokens: {
        access_token: "seller-access-token",
        expires_in: 900,
      },
    });

    renderLoginPage();

    await userEvent.type(screen.getByLabelText(/email or phone/i), "seller.local@example.com");
    await userEvent.type(screen.getByLabelText(/password/i), "LocalDemo#2026!");
    await userEvent.click(screen.getByRole("button", { name: /continue/i }));

    await screen.findByText("Seller Home");

    const [path, options] = mockedHttp.mock.calls[0] ?? [];
    const body = JSON.parse(String(options?.body));

    expect(path).toBe("/api/v1/auth/login");
    expect(body.identifier).toBe("seller.local@example.com");
    expect(body.password).toBe("LocalDemo#2026!");
    expect(body.device).toMatchObject({
      channel: "seller-dashboard",
      user_agent: expect.any(String),
    });
    expect(body.device).not.toHaveProperty("source");

    await waitFor(() => {
      expect(readSellerAuthSession()?.sellerId).toBe("seller_local_demo");
    });
  });
});
