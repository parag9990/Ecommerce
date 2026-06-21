import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { FunnelAnalysisPage } from "./funnel-analysis-page";

vi.mock("../components/funnel-chart", () => ({
  FunnelChart: ({ report }: { report: { overallConversionRate: number } }) => (
    <section>
      Mock funnel chart {report.overallConversionRate.toFixed(1)}%
    </section>
  )
}));

let requestUrls: URL[] = [];
let responseBody: unknown;
let responseStatus: number;

beforeEach(() => {
  requestUrls = [];
  responseBody = funnelResponse();
  responseStatus = 200;
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string | URL | Request) => {
      requestUrls.push(new URL(String(input)));
      return jsonResponse(responseBody, responseStatus);
    })
  );
});

describe("FunnelAnalysisPage", () => {
  it("renders the funnel chart, step cards, summary, and drop-off table", async () => {
    renderPage();

    expect(await screen.findByText("Product-to-paid funnel")).toBeInTheDocument();
    expect(screen.getAllByText("Product viewed").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Added to cart").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Checkout started").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Paid").length).toBeGreaterThan(0);
    expect(screen.getByText("Mock funnel chart 9.0%")).toBeInTheDocument();
    expect(screen.getByText("Step drop-off")).toBeInTheDocument();
    expect(screen.getByText("680")).toBeInTheDocument();
  });

  it("uses date range and segment filters in the funnel API request", async () => {
    renderPage();

    await screen.findByText("Product-to-paid funnel");
    await userEvent.selectOptions(screen.getByLabelText("Device"), "mobile");

    await waitFor(() => expect(requestUrls.length).toBeGreaterThanOrEqual(2));

    const params = requestUrls[requestUrls.length - 1].searchParams;
    expect(params.get("device_type")).toBe("mobile");
    expect(params.get("channel")).toBe("all");
    expect(params.get("source")).toBe("all");
    expect(params.get("user_type")).toBe("all");
    expect(params.get("steps")).toBe(
      "product_view,add_to_cart,checkout_started,paid"
    );
  });

  it("renders an empty state when the aggregate has no started sessions", async () => {
    responseBody = { steps: [] };

    renderPage();

    expect(
      await screen.findByText("No funnel data for this selection")
    ).toBeInTheDocument();
  });

  it("renders an error state when the funnel endpoint fails", async () => {
    responseStatus = 500;
    responseBody = {
            error: {
              code: "SESSION_ANALYTICS_UNAVAILABLE",
              message: "Session analytics unavailable."
            }
          };

    renderPage();

    expect(
      await screen.findByText("Funnel report unavailable")
    ).toBeInTheDocument();
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
      <FunnelAnalysisPage />
    </QueryClientProvider>
  );
}

function funnelResponse() {
  return {
    min_segment_size: 5,
    steps: [
      { key: "product_view", count: 1000 },
      { key: "add_to_cart", count: 320 },
      { key: "checkout_started", count: 180 },
      { key: "paid", count: 90 }
    ]
  };
}

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    headers: { "Content-Type": "application/json" },
    status
  });
}
