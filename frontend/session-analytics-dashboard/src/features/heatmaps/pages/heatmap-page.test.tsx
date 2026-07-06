import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import {
  beforeEach,
  describe,
  expect,
  it,
  vi
} from "vitest";

import { HeatmapPage } from "./heatmap-page";

vi.mock("../components/heatmap-canvas", () => ({
  HeatmapCanvas: ({ mode }: { mode: string }) => (
    <div data-testid="heatmap-canvas">Mock {mode} heatmap canvas</div>
  )
}));

let requestUrls: URL[] = [];
let responseBody: unknown;
let responseStatus: number;

beforeEach(() => {
  requestUrls = [];
  responseBody = heatmapResponse();
  responseStatus = 200;
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string | URL | Request) => {
      requestUrls.push(new URL(String(input)));
      return jsonResponse(responseBody, responseStatus);
    })
  );
});

describe("HeatmapPage", () => {
  it("renders summary, privacy state, preview, legend, and top click buckets", async () => {
    renderPage();

    expect(await screen.findByText("Total events")).toBeInTheDocument();
    expect(screen.getByText("48")).toBeInTheDocument();
    expect(screen.getByText("Heat buckets")).toBeInTheDocument();
    expect(screen.getByText("Page preview")).toBeInTheDocument();
    expect(screen.getByText("Intensity")).toBeInTheDocument();
    expect(screen.getByText("Privacy-safe aggregate")).toBeInTheDocument();
    expect(screen.getByTestId("heatmap-canvas")).toHaveTextContent(
      "Mock click heatmap canvas"
    );
    expect(screen.getByText("Top aggregate areas")).toBeInTheDocument();
  });

  it("sends selected page, device, date range, and mode to the heatmap API", async () => {
    renderPage();

    await screen.findByText("Total events");
    await userEvent.selectOptions(screen.getByLabelText("Page"), "/checkout");
    await userEvent.selectOptions(screen.getByLabelText("Device"), "mobile");
    await userEvent.click(screen.getByRole("button", { name: /scroll/i }));

    await waitFor(() =>
      expect(
        requestUrls.some(
          (url) => url.searchParams.get("heatmap_type") === "scroll"
        )
      ).toBe(true)
    );

    const latest = requestUrls[requestUrls.length - 1];
    expect(latest.pathname).toBe("/api/v1/analytics/heatmaps");
    expect(latest.searchParams.get("path")).toBe("/checkout");
    expect(latest.searchParams.get("device_type")).toBe("mobile");
    expect(latest.searchParams.get("heatmap_type")).toBe("scroll");
    expect(latest.searchParams.get("from")).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    expect(latest.searchParams.get("to")).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    expect(await screen.findByText("Scroll reach")).toBeInTheDocument();
  });

  it("renders an empty state when no aggregate points are returned", async () => {
    responseBody = { points: [] };

    renderPage();

    expect(
      await screen.findByText("No heatmap data for this selection")
    ).toBeInTheDocument();
  });

  it("renders an error state when the heatmap endpoint fails", async () => {
    responseStatus = 500;
    responseBody = {
            error: {
              code: "SESSION_ANALYTICS_UNAVAILABLE",
              message: "Session analytics unavailable."
            }
          };

    renderPage();

    expect(await screen.findByText("Heatmap unavailable")).toBeInTheDocument();
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
      <HeatmapPage />
    </QueryClientProvider>
  );
}

function heatmapResponse() {
  return {
    average_scroll_depth: 68,
    max_weight: 22,
    min_bucket_size: 5,
    partial: true,
    points: [
      { weight: 22, x: 42, y: 24 },
      { weight: 16, x: 61, y: 44 },
      { weight: 10, x: 48, y: 82 }
    ],
    total_events: 48
  };
}

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    headers: { "Content-Type": "application/json" },
    status
  });
}
