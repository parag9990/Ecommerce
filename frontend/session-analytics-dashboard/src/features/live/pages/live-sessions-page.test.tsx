import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

import { LiveSessionsPage } from "./live-sessions-page";

describe("LiveSessionsPage", () => {
  it("renders active sessions, summary, and breakdown panels", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse(activeSessionsResponse()))
    );

    renderPage();

    expect(await screen.findByText("anon_1...7890")).toBeInTheDocument();
    expect(screen.getByText("sess****")).toBeInTheDocument();
    expect(screen.getByText("Chrome / Android")).toBeInTheDocument();
    expect(screen.getAllByText("/products").length).toBeGreaterThan(0);
    expect(screen.getByText("Device mix")).toBeInTheDocument();
    expect(screen.getByText("Top entry pages")).toBeInTheDocument();
    expect(screen.getAllByText("Delhi, India").length).toBeGreaterThan(0);
    expect(
      screen.getByRole("link", { name: "Open journey for selected session" })
    ).toHaveAttribute("href", "/journey/sess_123");
  });

  it("renders an empty state when no sessions match", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({
          ...activeSessionsResponse(),
          activeSessions: 0,
          activeUsers: 0,
          deviceBreakdown: [],
          entryPageBreakdown: [],
          locationBreakdown: [],
          sessions: []
        })
      )
    );

    renderPage();

    expect(
      await screen.findByText("No active sessions match these filters")
    ).toBeInTheDocument();
  });

  it("renders an error state when the API request fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse(
          {
            error: {
              code: "SESSION_ANALYTICS_UNAVAILABLE",
              message: "Session analytics unavailable."
            }
          },
          { status: 500 }
        )
      )
    );

    renderPage();

    expect(
      await screen.findByText("Active sessions unavailable")
    ).toBeInTheDocument();
  });

  it("sends filter changes as active sessions query params", async () => {
    const fetchMock = vi.fn(async () => jsonResponse(activeSessionsResponse()));
    vi.stubGlobal("fetch", fetchMock);

    renderPage();

    await screen.findByText("sess****");
    await userEvent.selectOptions(screen.getByLabelText("Device"), "mobile");

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));

    const [url] = fetchMock.mock.calls[1] as [string, RequestInit];
    expect(url).toContain("/api/v1/analytics/sessions?");
    expect(url).toContain("status=active");
    expect(url).toContain("device_type=mobile");
    expect(url).toContain("page_size=50");
  });
});

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false
      }
    }
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter
        future={{ v7_relativeSplatPath: true, v7_startTransition: true }}
      >
        <LiveSessionsPage />
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

function activeSessionsResponse() {
  return {
    activeSessions: 1,
    activeUsers: 1,
    deviceBreakdown: [{ count: 1, label: "mobile", percentage: 100 }],
    entryPageBreakdown: [{ count: 1, label: "/products", percentage: 100 }],
    eventsPerMinute: 4,
    locationBreakdown: [{ count: 1, label: "Delhi, India", percentage: 100 }],
    refreshedAt: "2026-05-28T10:00:00.000Z",
    sessions: [
      {
        anonymousId: "anon_1234567890",
        currentPage: "/products",
        device: { browser: "Chrome", os: "Android", type: "mobile" },
        durationSeconds: 120,
        entryPage: "/",
        eventCount: 9,
        lastSeenAt: new Date().toISOString(),
        location: { city: "Delhi", country: "India" },
        sessionId: "sess_123",
        startedAt: "2026-05-28T09:58:00.000Z"
      }
    ]
  };
}
