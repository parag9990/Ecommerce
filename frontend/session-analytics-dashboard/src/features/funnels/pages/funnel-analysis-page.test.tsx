import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";

import { FunnelAnalysisPage } from "./funnel-analysis-page";

vi.mock("../components/funnel-chart", () => ({
  FunnelChart: ({ report }: { report: { overallConversionRate: number } }) => (
    <section>
      Mock funnel chart {report.overallConversionRate.toFixed(1)}%
    </section>
  )
}));

let requestUrls: URL[] = [];

const server = setupServer(
  http.get("/api/v1/analytics/funnels", ({ request }) => {
    requestUrls.push(new URL(request.url));
    return HttpResponse.json(funnelResponse());
  })
);

beforeAll(() => server.listen());
afterEach(() => {
  requestUrls = [];
  server.resetHandlers();
});
afterAll(() => server.close());

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
    server.use(
      http.get("/api/v1/analytics/funnels", () =>
        HttpResponse.json({ steps: [] })
      )
    );

    renderPage();

    expect(
      await screen.findByText("No funnel data for this selection")
    ).toBeInTheDocument();
  });

  it("renders an error state when the funnel endpoint fails", async () => {
    server.use(
      http.get("/api/v1/analytics/funnels", () =>
        HttpResponse.json(
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
