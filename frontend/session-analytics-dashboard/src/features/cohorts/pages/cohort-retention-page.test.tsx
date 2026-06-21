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

import { CohortRetentionPage } from "./cohort-retention-page";

vi.mock("../components/new-returning-chart", () => ({
  NewReturningChart: ({
    data
  }: {
    data: Array<{ bucket: string; newUsers: number; returningUsers: number }>;
  }) => <section>Mock new returning chart {data.length}</section>
}));

let requestUrls: URL[] = [];
let responseBody: unknown;
let responseStatus: number;

beforeEach(() => {
  requestUrls = [];
  responseBody = retentionResponse();
  responseStatus = 200;
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string | URL | Request) => {
      requestUrls.push(new URL(String(input)));
      return jsonResponse(responseBody, responseStatus);
    })
  );
});

describe("CohortRetentionPage", () => {
  it("renders summary cards, chart, privacy notice, legend, and matrix", async () => {
    renderPage();

    expect(await screen.findByText("New users")).toBeInTheDocument();
    expect(screen.getAllByText("1,200").length).toBeGreaterThan(0);
    expect(screen.getByText("Returning users")).toBeInTheDocument();
    expect(screen.getByText("40.0%")).toBeInTheDocument();
    expect(screen.getByText("Mock new returning chart 2")).toBeInTheDocument();
    expect(screen.getByText("Privacy-safe aggregate")).toBeInTheDocument();
    expect(screen.getByText("Retention scale")).toBeInTheDocument();
    expect(screen.getByText("Cohort retention")).toBeInTheDocument();
    expect(screen.getAllByText("May 18 - May 24").length).toBeGreaterThan(0);
    expect(screen.getAllByText("35.0%").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Masked").length).toBeGreaterThan(0);
  });

  it("uses date range, segment filters, interval, and window in the API request", async () => {
    renderPage();

    await screen.findByText("Cohort retention");
    await userEvent.selectOptions(screen.getByLabelText("Device"), "mobile");
    await userEvent.selectOptions(screen.getByLabelText("Window"), "4");

    await waitFor(() =>
      expect(
        requestUrls.some((url) => url.searchParams.get("window") === "4")
      ).toBe(true)
    );

    const latest = requestUrls[requestUrls.length - 1];
    expect(latest.pathname).toBe("/api/v1/analytics/retention");
    expect(latest.searchParams.get("device_type")).toBe("mobile");
    expect(latest.searchParams.get("channel")).toBeNull();
    expect(latest.searchParams.get("source")).toBeNull();
    expect(latest.searchParams.get("user_type")).toBeNull();
    expect(latest.searchParams.get("interval")).toBe("week");
    expect(latest.searchParams.get("window")).toBe("4");
    expect(latest.searchParams.get("from")).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    expect(latest.searchParams.get("to")).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });

  it("updates window options when the interval changes", async () => {
    renderPage();

    await screen.findByText("Cohort retention");
    await userEvent.selectOptions(screen.getByLabelText("Interval"), "day");

    await waitFor(() =>
      expect(
        requestUrls.some(
          (url) =>
            url.searchParams.get("interval") === "day" &&
            url.searchParams.get("window") === "7"
        )
      ).toBe(true)
    );
  });

  it("renders an empty state when no cohorts are returned", async () => {
    responseBody = {
          cohorts: [],
          meta: { from: "2026-05-22", interval: "week", to: "2026-05-28", window: 8 },
          new_vs_returning: [],
          summary: {
            average_retention: 0,
            new_users: 0,
            returning_rate: 0,
            returning_users: 0
          }
        };

    renderPage();

    expect(
      await screen.findByText("No retention data for this selection")
    ).toBeInTheDocument();
  });

  it("renders an error state when the retention endpoint fails", async () => {
    responseStatus = 500;
    responseBody = {
            error: {
              code: "SESSION_ANALYTICS_UNAVAILABLE",
              message: "Session analytics unavailable."
            }
          };

    renderPage();

    expect(
      await screen.findByText("Retention report unavailable")
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
      <CohortRetentionPage />
    </QueryClientProvider>
  );
}

function retentionResponse() {
  return {
    cohorts: [
      {
        buckets: [
          { label: "W0", offset: 0, rate: 100, users: 1200 },
          { label: "W1", offset: 1, rate: 35, users: 420 },
          { label: "W2", offset: 2, rate: 24, users: 288 }
        ],
        cohort_key: "2026-W21",
        cohort_label: "May 18 - May 24",
        cohort_size: 1200
      },
      {
        buckets: [
          { label: "W0", offset: 0, rate: 100, users: 4 },
          { label: "W1", offset: 1, rate: 0, suppressed: true, users: 0 }
        ],
        cohort_key: "2026-W22",
        cohort_label: "May 25 - May 31",
        cohort_size: 4,
        suppressed: true
      }
    ],
    meta: {
      from: "2026-05-22",
      interval: "week",
      partial: true,
      small_count_threshold: 5,
      to: "2026-05-28",
      window: 8
    },
    new_vs_returning: [
      { bucket: "2026-W21", label: "May 18", new_users: 1200, returning_users: 800 },
      { bucket: "2026-W22", label: "May 25", new_users: 900, returning_users: 700 }
    ],
    summary: {
      average_retention: 28.5,
      best_cohort: "2026-W21",
      new_users: 1200,
      returning_rate: 40,
      returning_users: 800,
      worst_cohort: "2026-W22"
    }
  };
}

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    headers: { "Content-Type": "application/json" },
    status
  });
}
