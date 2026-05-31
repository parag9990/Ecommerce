import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { PrivacyControlsPage } from "./privacy-controls-page";

describe("PrivacyControlsPage", () => {
  it("renders privacy masking, retention, and deletion panels", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);

        if (url.endsWith("/api/v1/analytics/privacy/settings")) {
          return jsonResponse({
            masking: {
              anonymousIdMode: "masked",
              locationGranularity: "country",
              sessionIdMode: "masked",
              showIpHash: false,
              showSearchQueries: false,
              userIdMode: "masked"
            },
            permissions: {
              canRequestDeletion: true,
              canUpdateMasking: true,
              canUpdateRetention: true
            },
            updatedAt: "2026-05-28T04:00:00Z",
            updatedBy: "admin_123"
          });
        }

        if (url.endsWith("/api/v1/analytics/privacy/retention")) {
          return jsonResponse({
            activeSessionTtlMinutes: 45,
            analyticsAggregatesMonths: 36,
            deletionRequestLogDays: 730,
            heatmapAggregatesDays: 365,
            journeySummariesDays: 365,
            rawEventsDays: 90
          });
        }

        if (url.endsWith("/api/v1/analytics/privacy/deletion-requests")) {
          return jsonResponse({ items: [] });
        }

        return jsonResponse({}, { status: 404 });
      })
    );

    render(
      <QueryClientProvider client={newTestQueryClient()}>
        <PrivacyControlsPage />
      </QueryClientProvider>
    );

    expect(
      await screen.findByRole("heading", { name: "Privacy controls" })
    ).toBeInTheDocument();
    expect(screen.getByText("PII masking policy")).toBeInTheDocument();
    expect(screen.getByText("Retention settings")).toBeInTheDocument();
    expect(screen.getByText("Deletion request")).toBeInTheDocument();
    expect(screen.getByText("Recent deletion requests")).toBeInTheDocument();
  });
});

function newTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false
      }
    }
  });
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
