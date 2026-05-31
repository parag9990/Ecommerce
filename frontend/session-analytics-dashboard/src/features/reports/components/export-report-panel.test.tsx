import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { ReportFilters } from "../../../api/session-api";
import { ExportReportPanel } from "./export-report-panel";

const filters: ReportFilters = {
  channel: "web",
  deviceType: "mobile",
  from: "2026-05-22",
  source: "paid",
  timezone: "UTC",
  to: "2026-05-28",
  userType: "logged_in"
};

describe("ExportReportPanel", () => {
  it("downloads the selected report as csv", async () => {
    const fetchMock = vi.fn(async () =>
      new Response("step,count\nproduct_view,1200\n", {
        headers: {
          "Content-Disposition": 'attachment; filename="funnel.csv"',
          "Content-Type": "text/csv; charset=utf-8"
        }
      })
    );
    const click = vi.fn();
    const originalCreateElement = document.createElement.bind(document);
    const URLConstructor = globalThis.URL;
    const URLMock = function URL(value: string | URL, base?: string | URL) {
      return new URLConstructor(value, base);
    } as typeof globalThis.URL;
    Object.assign(URLMock, URLConstructor, {
      createObjectURL: vi.fn(() => "blob:report"),
      revokeObjectURL: vi.fn()
    });

    vi.stubGlobal("fetch", fetchMock);
    vi.stubGlobal("URL", URLMock);
    vi.spyOn(document, "createElement").mockImplementation((tagName) => {
      const element = originalCreateElement(tagName);
      if (tagName === "a") {
        Object.defineProperty(element, "click", {
          configurable: true,
          value: click
        });
      }

      return element;
    });

    render(
      <QueryClientProvider client={new QueryClient()}>
        <ExportReportPanel dateRangeValid filters={filters} />
      </QueryClientProvider>
    );

    await userEvent.click(
      screen.getByRole("button", { name: /download csv/i })
    );

    await waitFor(() => expect(fetchMock).toHaveBeenCalledOnce());
    await waitFor(() => expect(click).toHaveBeenCalledOnce());

    const [url] = fetchMock.mock.calls[0] as [string, RequestInit];
    const params = new URL(url, "http://localhost").searchParams;

    expect(params.get("report_type")).toBe("funnel");
    expect(params.get("device")).toBe("mobile");
    expect(params.get("source")).toBe("paid");
  });
});
