import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

import { JourneyExplorerPage } from "./journey-explorer-page";

describe("JourneyExplorerPage", () => {
  it("renders session context, timeline, and sanitized event details", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse(journeyResponse()))
    );

    renderPage("/journey/sess_123");

    expect(await screen.findByText("sess_123")).toBeInTheDocument();
    expect(screen.getAllByText("Add to Cart").length).toBeGreaterThan(0);

    await userEvent.click(screen.getByText("Added product prod_123 to cart"));

    expect(screen.getAllByText(/product_id/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/prod_123/).length).toBeGreaterThan(0);
    expect(screen.getByText(/email/)).toBeInTheDocument();
    expect(screen.getByText(/\[masked\]/)).toBeInTheDocument();
  });

  it("renders the no-session state on the base route", () => {
    renderPage("/journey");

    expect(screen.getByText("Select a session")).toBeInTheDocument();
  });
});

function renderPage(initialEntry: string) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false
      }
    }
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[initialEntry]}>
        <Routes>
          <Route path="/journey" element={<JourneyExplorerPage />} />
          <Route path="/journey/:sessionId" element={<JourneyExplorerPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  );
}

function jsonResponse(body: unknown, init: ResponseInit = {}) {
  return new Response(JSON.stringify(body), {
    status: 200,
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init.headers
    }
  });
}

function journeyResponse() {
  return {
    events: [
      {
        event_id: "evt_1",
        event_type: "add_to_cart",
        occurred_at: "2026-05-18T00:06:30Z",
        path: "/products/prod_123",
        properties: {
          email: "buyer@example.com",
          product_id: "prod_123",
          quantity: 1
        },
        session_id: "sess_123"
      }
    ],
    session: {
      anonymous_id: "anon_1234567890",
      device: { browser: "Chrome", os: "Android", type: "mobile" },
      duration_seconds: 390,
      entry_page: "/",
      geo: { city: "Delhi", country: "India" },
      last_seen_at: "2026-05-18T00:06:30Z",
      session_id: "sess_123",
      started_at: "2026-05-18T00:00:00Z",
      status: "active"
    },
    summary: {
      cart_actions: 1,
      checkout_started: false,
      clicks: 0,
      page_views: 0,
      payment_completed: false,
      total_events: 1
    }
  };
}
