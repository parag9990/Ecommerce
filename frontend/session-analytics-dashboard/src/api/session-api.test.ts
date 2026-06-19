import { describe, expect, it, vi } from "vitest";

import { ApiError } from "../lib/http";
import { defaultSegmentFilters } from "../features/shell/types";
import {
  createReportSchedule,
  createDeletionRequest,
  exportAnalyticsReport,
  getActiveSessions,
  getFunnelReport,
  getHeatmap,
  getLiveMetrics,
  getPrivacySettings,
  getReportSchedules,
  getRetentionSettings,
  getRetentionReport,
  getSessionJourney,
  listDeletionRequests,
  previewDeletion,
  updatePrivacySettings,
  updateRetentionSettings,
  updateReportScheduleStatus
} from "./session-api";

const request = {
  dateRange: {
    preset: "7d" as const,
    from: "2026-05-22",
    to: "2026-05-28"
  },
  filters: defaultSegmentFilters
};

describe("session analytics api", () => {
  it("fetches live metrics with date and segment query params", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        activeUsersNow: 42,
        averageSessionDurationSeconds: 245,
        bounceRate: 41.2,
        conversionRate: 3.8,
        productViewToCartRate: 12.6,
        sessionsToday: 1280
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(getLiveMetrics(request)).resolves.toEqual({
      activeUsersNow: 42,
      averageSessionDurationSeconds: 245,
      bounceRate: 41.2,
      conversionRate: 3.8,
      productViewToCartRate: 12.6,
      sessionsToday: 1280
    });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/analytics/live?");
    expect(url).toContain("from=2026-05-22");
    expect(url).toContain("to=2026-05-28");
    expect(url).toContain("device_type=all");
    expect(url).toContain("channel=all");
    expect(url).toContain("source=all");
    expect(url).toContain("user_type=all");
    expect(init.credentials).toBe("include");
  });

  it("unwraps API gateway response envelopes", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({
          data: {
            activeUsersNow: 1,
            averageSessionDurationSeconds: 10,
            bounceRate: 0,
            conversionRate: 0,
            productViewToCartRate: 0,
            sessionsToday: 2
          }
        })
      )
    );

    await expect(getLiveMetrics(request)).resolves.toMatchObject({
      activeUsersNow: 1,
      sessionsToday: 2
    });
  });

  it("fetches funnel reports with date, segment, and ordered step params", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        generated_at: "2026-05-28T10:00:00.000Z",
        min_segment_size: 5,
        steps: [
          { key: "product_view", count: 1000 },
          { key: "add_to_cart", sessions: "320" },
          { key: "checkout_started", unique_sessions: 180 },
          { key: "paid", count: 90 }
        ]
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      getFunnelReport({
        dateRange: request.dateRange,
        filters: {
          channel: "web",
          deviceType: "mobile",
          source: "paid",
          userType: "logged_in"
        },
        steps: ["product_view", "add_to_cart", "checkout_started", "paid"]
      })
    ).resolves.toMatchObject({
      generatedAt: "2026-05-28T10:00:00.000Z",
      minSegmentSize: 5,
      steps: [
        { count: 1000, key: "product_view" },
        { key: "add_to_cart", sessions: 320 },
        { key: "checkout_started", uniqueSessions: 180 },
        { count: 90, key: "paid" }
      ]
    });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    const params = new URL(url, "http://localhost").searchParams;

    expect(url).toContain("/api/v1/analytics/funnels?");
    expect(params.get("from")).toBe("2026-05-22");
    expect(params.get("to")).toBe("2026-05-28");
    expect(params.get("device_type")).toBe("mobile");
    expect(params.get("channel")).toBe("web");
    expect(params.get("source")).toBe("paid");
    expect(params.get("user_type")).toBe("logged_in");
    expect(params.get("steps")).toBe(
      "product_view,add_to_cart,checkout_started,paid"
    );
    expect(init.credentials).toBe("include");
  });

  it("rejects invalid funnel request inputs before calling fetch", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      getFunnelReport({
        dateRange: request.dateRange,
        filters: defaultSegmentFilters,
        steps: []
      })
    ).rejects.toMatchObject({
      code: "VALIDATION_ERROR",
      status: 400
    } satisfies Partial<ApiError>);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("fetches retention reports with date, segment, interval, and window params", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        cohorts: [
          {
            buckets: [
              { label: "W0", offset: 0, rate: 100, users: 1200 },
              { label: "W1", offset: 1, rate: "35.5", users: 426 }
            ],
            cohort_key: "2026-W21",
            cohort_label: "May 18 - May 24",
            cohort_size: 1200
          }
        ],
        meta: {
          from: "2026-05-22",
          interval: "week",
          small_count_threshold: 5,
          to: "2026-05-28",
          window: 8
        },
        new_vs_returning: [
          { bucket: "2026-W21", new_users: 1200, returning_users: 640 }
        ],
        summary: {
          average_retention: 35.5,
          best_cohort: "2026-W21",
          new_users: 1200,
          returning_rate: 34.78,
          returning_users: 640
        }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      getRetentionReport({
        dateRange: request.dateRange,
        filters: {
          channel: "web",
          deviceType: "mobile",
          source: "paid",
          userType: "logged_in"
        },
        interval: "week",
        window: 8
      })
    ).resolves.toMatchObject({
      cohorts: [
        {
          buckets: [
            { offset: 0, rate: 100, users: 1200 },
            { offset: 1, rate: 35.5, users: 426 }
          ],
          cohortKey: "2026-W21",
          cohortSize: 1200
        }
      ],
      meta: {
        interval: "week",
        smallCountThreshold: 5,
        window: 8
      },
      newVsReturning: [
        {
          bucket: "2026-W21",
          newUsers: 1200,
          returningUsers: 640,
          totalUsers: 1840
        }
      ],
      summary: {
        averageRetention: 35.5,
        bestCohort: "2026-W21",
        newUsers: 1200,
        returningRate: 34.78,
        returningUsers: 640
      }
    });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    const params = new URL(url, "http://localhost").searchParams;

    expect(url).toContain("/api/v1/analytics/retention?");
    expect(params.get("from")).toBe("2026-05-22");
    expect(params.get("to")).toBe("2026-05-28");
    expect(params.get("interval")).toBe("week");
    expect(params.get("window")).toBe("8");
    expect(params.get("device_type")).toBe("mobile");
    expect(params.get("channel")).toBe("web");
    expect(params.get("source")).toBe("paid");
    expect(params.get("user_type")).toBe("logged_in");
    expect(init.credentials).toBe("include");
  });

  it("rejects invalid retention windows before calling fetch", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      getRetentionReport({
        dateRange: request.dateRange,
        filters: defaultSegmentFilters,
        interval: "week",
        window: 30
      })
    ).rejects.toMatchObject({
      code: "VALIDATION_ERROR",
      status: 400
    } satisfies Partial<ApiError>);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("fetches heatmap aggregates with path, device, range, and mode params", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        average_scroll_depth: "64.5",
        generated_at: "2026-05-28T10:00:00.000Z",
        max_weight: 18,
        min_bucket_size: 5,
        partial: true,
        points: [
          { weight: 18, x: 42.5, y: 20 },
          { weight: 7, x: 60, y: 76 }
        ],
        total_events: 25
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      getHeatmap({
        deviceType: "mobile",
        from: "2026-05-22",
        mode: "scroll",
        path: " /products/prod_123 ",
        to: "2026-05-28"
      })
    ).resolves.toEqual({
      averageScrollDepth: 64.5,
      generatedAt: "2026-05-28T10:00:00.000Z",
      maxWeight: 18,
      minBucketSize: 5,
      partial: true,
      points: [
        { weight: 18, x: 42.5, y: 20 },
        { weight: 7, x: 60, y: 76 }
      ],
      suppressed: undefined,
      totalEvents: 25
    });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    const params = new URL(url, "http://localhost").searchParams;

    expect(url).toContain("/api/v1/analytics/heatmaps?");
    expect(params.get("path")).toBe("/products/prod_123");
    expect(params.get("device_type")).toBe("mobile");
    expect(params.get("from")).toBe("2026-05-22");
    expect(params.get("to")).toBe("2026-05-28");
    expect(params.get("mode")).toBe("scroll");
    expect(init.credentials).toBe("include");
  });

  it("rejects invalid heatmap request inputs before calling fetch", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      getHeatmap({
        deviceType: "desktop",
        from: "2026-05-22",
        mode: "click",
        path: "https://example.com/products",
        to: "2026-05-28"
      })
    ).rejects.toMatchObject({
      code: "VALIDATION_ERROR",
      status: 400
    } satisfies Partial<ApiError>);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("rejects invalid date ranges before calling fetch", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      getLiveMetrics({
        ...request,
        dateRange: {
          preset: "custom",
          from: "2026-05-29",
          to: "2026-05-28"
        }
      })
    ).rejects.toMatchObject({
      code: "VALIDATION_ERROR",
      status: 400
    } satisfies Partial<ApiError>);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("fetches active sessions with active status and lightweight filters", async () => {
    const fetchMock = vi.fn(async () => jsonResponse(activeSessionsResponse()));
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      getActiveSessions({
        country: "India",
        deviceType: "mobile",
        entryPage: "/products",
        limit: 25,
        q: "anon_123"
      })
    ).resolves.toMatchObject({
      activeSessions: 1,
      activeUsers: 1,
      sessions: [
        {
          anonymousId: "anon_1234567890",
          device: { type: "mobile" },
          entryPage: "/",
          sessionId: "sess_123"
        }
      ]
    });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/analytics/sessions?");
    expect(url).toContain("status=active");
    expect(url).toContain("device_type=mobile");
    expect(url).toContain("country=India");
    expect(url).toContain("entry_page=%2Fproducts");
    expect(url).toContain("q=anon_123");
    expect(url).toContain("limit=25");
    expect(init.credentials).toBe("include");
  });

  it("rejects invalid active session limits before calling fetch", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    await expect(getActiveSessions({ limit: 250 })).rejects.toMatchObject({
      code: "VALIDATION_ERROR",
      status: 400
    } satisfies Partial<ApiError>);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("fetches a session journey and derives summary when omitted", async () => {
    const fetchMock = vi.fn(async () => jsonResponse(journeyResponse()));
    vi.stubGlobal("fetch", fetchMock);

    await expect(getSessionJourney(" sess_123 ")).resolves.toMatchObject({
      events: [
        { eventId: "evt_1", eventType: "page_view", path: "/" },
        { eventId: "evt_2", eventType: "add_to_cart" },
        { eventId: "evt_3", eventType: "payment_result" }
      ],
      session: {
        anonymousId: "anon_1234567890",
        durationSeconds: 1320,
        entryPage: "/",
        exitPage: "/checkout",
        sessionId: "sess_123",
        status: "ended"
      },
      summary: {
        cartActions: 1,
        checkoutStarted: false,
        clicks: 0,
        pageViews: 1,
        paymentCompleted: true,
        totalEvents: 3
      }
    });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/analytics/sessions/sess_123/journey");
    expect(init.credentials).toBe("include");
  });

  it("rejects invalid journey session ids before calling fetch", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    await expect(getSessionJourney("sess/123")).rejects.toMatchObject({
      code: "VALIDATION_ERROR",
      status: 400
    } satisfies Partial<ApiError>);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("exports analytics reports as csv with report query params", async () => {
    const fetchMock = vi.fn(async () =>
      new Response("step,count\nproduct_view,1200\n", {
        headers: {
          "Content-Disposition": 'attachment; filename="funnel.csv"',
          "Content-Type": "text/csv; charset=utf-8"
        },
        status: 200
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    const result = await exportAnalyticsReport({
      channel: "web",
      deviceType: "mobile",
      format: "csv",
      from: "2026-05-22",
      reportType: "funnel",
      source: "paid",
      timezone: "UTC",
      to: "2026-05-28",
      userType: "logged_in"
    });

    await expect(result.blob.text()).resolves.toContain("product_view");
    expect(result.filename).toBe("funnel.csv");

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    const params = new URL(url, "http://localhost").searchParams;

    expect(url).toContain("/api/v1/analytics/reports/export?");
    expect(params.get("report_type")).toBe("funnel");
    expect(params.get("format")).toBe("csv");
    expect(params.get("from")).toBe("2026-05-22");
    expect(params.get("to")).toBe("2026-05-28");
    expect(params.get("timezone")).toBe("UTC");
    expect(params.get("device")).toBe("mobile");
    expect(params.get("channel")).toBe("web");
    expect(params.get("source")).toBe("paid");
    expect(params.get("user_type")).toBe("logged_in");
    expect((init.headers as Headers).get("Accept")).toBe("text/csv");
    expect(init.credentials).toBe("include");
  });

  it("skips all-valued export segment filters", async () => {
    const fetchMock = vi.fn(async () =>
      new Response("metric,value\nsessions,10\n", {
        headers: { "Content-Type": "text/csv; charset=utf-8" },
        status: 200
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await exportAnalyticsReport({
      channel: "all",
      deviceType: "all",
      format: "csv",
      from: "2026-05-22",
      reportType: "overview",
      source: "all",
      timezone: "UTC",
      to: "2026-05-28",
      userType: "all"
    });

    const [url] = fetchMock.mock.calls[0] as [string, RequestInit];
    const params = new URL(url, "http://localhost").searchParams;

    expect(params.has("device")).toBe(false);
    expect(params.has("channel")).toBe(false);
    expect(params.has("source")).toBe(false);
    expect(params.has("user_type")).toBe(false);
  });

  it("lists report schedules and normalizes snake case fields", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({
          items: [reportScheduleResponse()]
        })
      )
    );

    await expect(getReportSchedules()).resolves.toEqual({
      items: [
        expect.objectContaining({
          dayOfWeek: "monday",
          filters: expect.objectContaining({
            deviceType: "mobile",
            userType: "logged_in"
          }),
          format: "csv",
          frequency: "weekly",
          id: "rpt_sch_123",
          reportType: "retention",
          status: "active"
        })
      ]
    });
  });

  it("creates report schedules with csv-only snake case payload", async () => {
    const fetchMock = vi.fn(async () => jsonResponse(reportScheduleResponse()));
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      createReportSchedule({
        dayOfWeek: "monday",
        filters: {
          deviceType: "mobile",
          from: "2026-05-22",
          source: "paid",
          timezone: "UTC",
          to: "2026-05-28",
          userType: "logged_in"
        },
        format: "csv",
        frequency: "weekly",
        name: "Weekly retention report",
        recipients: ["Analytics@Example.com"],
        reportType: "retention",
        timeOfDay: "09:00",
        timezone: "UTC"
      })
    ).resolves.toMatchObject({
      id: "rpt_sch_123",
      reportType: "retention"
    });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    const body = JSON.parse(init.body as string) as Record<string, unknown>;

    expect(url).toBe("/api/v1/analytics/reports/schedules");
    expect(init.method).toBe("POST");
    expect(body).toMatchObject({
      day_of_week: "monday",
      format: "csv",
      frequency: "weekly",
      name: "Weekly retention report",
      recipients: ["analytics@example.com"],
      report_type: "retention",
      time_of_day: "09:00",
      timezone: "UTC"
    });
    expect(body.filters).toMatchObject({
      device: "mobile",
      from: "2026-05-22",
      source: "paid",
      timezone: "UTC",
      to: "2026-05-28",
      user_type: "logged_in"
    });
  });

  it("rejects invalid schedule recipients before calling fetch", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      createReportSchedule({
        dayOfWeek: "monday",
        filters: {
          from: "2026-05-22",
          timezone: "UTC",
          to: "2026-05-28"
        },
        format: "csv",
        frequency: "weekly",
        name: "Weekly report",
        recipients: ["not-an-email"],
        reportType: "funnel",
        timeOfDay: "09:00",
        timezone: "UTC"
      })
    ).rejects.toMatchObject({
      code: "VALIDATION_ERROR",
      status: 400
    } satisfies Partial<ApiError>);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("updates report schedule status with PATCH", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        ...reportScheduleResponse(),
        status: "paused"
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      updateReportScheduleStatus({
        id: "rpt_sch_123",
        status: "paused"
      })
    ).resolves.toMatchObject({
      id: "rpt_sch_123",
      status: "paused"
    });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/v1/analytics/reports/schedules/rpt_sch_123");
    expect(init.method).toBe("PATCH");
    expect(JSON.parse(init.body as string)).toEqual({ status: "paused" });
  });

  it("fetches privacy settings and normalizes snake case fields", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({
          data: {
            masking: {
              anonymous_id_mode: "masked",
              location_granularity: "country",
              session_id_mode: "masked",
              show_ip_hash: false,
              show_search_queries: false,
              user_id_mode: "hidden"
            },
            permissions: {
              can_request_deletion: true,
              can_update_masking: true,
              can_update_retention: false
            },
            updated_at: "2026-05-28T04:00:00Z",
            updated_by: "admin_123"
          }
        })
      )
    );

    await expect(getPrivacySettings()).resolves.toEqual({
      masking: {
        anonymousIdMode: "masked",
        locationGranularity: "country",
        sessionIdMode: "masked",
        showIpHash: false,
        showSearchQueries: false,
        userIdMode: "hidden"
      },
      permissions: {
        canRequestDeletion: true,
        canUpdateMasking: true,
        canUpdateRetention: false
      },
      updatedAt: "2026-05-28T04:00:00Z",
      updatedBy: "admin_123"
    });
  });

  it("updates privacy settings with a masking payload", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        masking: {
          anonymousIdMode: "masked",
          locationGranularity: "country",
          sessionIdMode: "masked",
          showIpHash: false,
          showSearchQueries: true,
          userIdMode: "masked"
        },
        permissions: {
          canRequestDeletion: true,
          canUpdateMasking: true,
          canUpdateRetention: true
        },
        updatedAt: "2026-05-28T04:00:00Z",
        updatedBy: "ops_1"
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await updatePrivacySettings({
      anonymousIdMode: "masked",
      locationGranularity: "country",
      sessionIdMode: "masked",
      showIpHash: false,
      showSearchQueries: true,
      userIdMode: "masked"
    });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/v1/analytics/privacy/settings");
    expect(init.method).toBe("PATCH");
    expect(JSON.parse(init.body as string)).toMatchObject({
      masking: {
        showSearchQueries: true,
        userIdMode: "masked"
      }
    });
  });

  it("fetches and updates retention privacy settings", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse({
          raw_events_days: 90,
          journey_summaries_days: 365,
          heatmap_aggregates_days: 365,
          analytics_aggregates_months: 36,
          active_session_ttl_minutes: 45,
          deletion_request_log_days: 730
        })
      )
      .mockResolvedValueOnce(
        jsonResponse({
          rawEventsDays: 60,
          journeySummariesDays: 365,
          heatmapAggregatesDays: 365,
          analyticsAggregatesMonths: 36,
          activeSessionTtlMinutes: 45,
          deletionRequestLogDays: 730
        })
      );
    vi.stubGlobal("fetch", fetchMock);

    await expect(getRetentionSettings()).resolves.toMatchObject({
      activeSessionTtlMinutes: 45,
      rawEventsDays: 90
    });

    await updateRetentionSettings(
      {
        activeSessionTtlMinutes: 45,
        analyticsAggregatesMonths: 36,
        deletionRequestLogDays: 730,
        heatmapAggregatesDays: 365,
        journeySummariesDays: 365,
        rawEventsDays: 60
      },
      "Reduce raw event storage after security review"
    );

    const [, init] = fetchMock.mock.calls[1] as [string, RequestInit];
    expect(init.method).toBe("PATCH");
    expect(JSON.parse(init.body as string)).toMatchObject({
      rawEventsDays: 60,
      reason: "Reduce raw event storage after security review"
    });
  });

  it("previews and creates analytics deletion requests safely", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse({
          aggregate_impact: "aggregates_anonymized_or_unchanged",
          estimated_completion_seconds: 20,
          matched_active_sessions: 1,
          matched_events: 1240,
          matched_journey_summaries: 18,
          matched_sessions: 18,
          target_type: "user_id",
          target_value_masked: "user_1...6789"
        })
      )
      .mockResolvedValueOnce(
        jsonResponse({
          completed_at: "2026-05-28T04:00:20Z",
          created_at: "2026-05-28T04:00:00Z",
          matched_active_sessions: 1,
          matched_events: 1240,
          matched_journey_summaries: 18,
          matched_sessions: 18,
          reason: "Support ticket SUP-123 user requested analytics deletion",
          requested_by: "ops_1",
          request_id: "delreq_123",
          status: "completed",
          target_type: "user_id",
          target_value_masked: "user_1...6789"
        })
      );
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      previewDeletion({
        targetType: "user_id",
        targetValue: " user_123456789 "
      })
    ).resolves.toMatchObject({
      matchedEvents: 1240,
      targetValueMasked: "user_1...6789"
    });

    await expect(
      createDeletionRequest({
        confirmed: true,
        reason: "Support ticket SUP-123 user requested analytics deletion",
        targetType: "user_id",
        targetValue: "user_123456789"
      })
    ).resolves.toMatchObject({
      requestId: "delreq_123",
      status: "completed",
      targetValueMasked: "user_1...6789"
    });

    const [, previewInit] = fetchMock.mock.calls[0] as [string, RequestInit];
    const [, createInit] = fetchMock.mock.calls[1] as [string, RequestInit];
    expect(JSON.parse(previewInit.body as string)).toEqual({
      targetType: "user_id",
      targetValue: "user_123456789"
    });
    expect(JSON.parse(createInit.body as string)).not.toHaveProperty(
      "targetValueMasked"
    );
  });

  it("lists deletion requests from an items response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({
          items: [
            {
              createdAt: "2026-05-28T04:00:00Z",
              matchedActiveSessions: 0,
              matchedEvents: 1,
              matchedJourneySummaries: 0,
              matchedSessions: 1,
              reason: "Support ticket SUP-123",
              requestedBy: "ops_1",
              requestId: "delreq_123",
              status: "queued",
              targetType: "session_id",
              targetValueMasked: "sess_1...6789"
            }
          ]
        })
      )
    );

    await expect(listDeletionRequests()).resolves.toEqual({
      items: [
        expect.objectContaining({
          requestId: "delreq_123",
          targetType: "session_id"
        })
      ]
    });
  });
});

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
    entryPageBreakdown: [{ count: 1, label: "/", percentage: 100 }],
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
        lastSeenAt: "2026-05-28T09:59:56.000Z",
        location: { city: "Delhi", country: "India" },
        sessionId: "sess_123",
        startedAt: "2026-05-28T09:58:00.000Z"
      }
    ]
  };
}

function journeyResponse() {
  return {
    events: [
      {
        _id: "evt_1",
        event_type: "page_view",
        occurred_at: "2026-05-18T00:00:02Z",
        path: "/",
        properties: { title: "Home" },
        session_id: "sess_123"
      },
      {
        _id: "evt_2",
        event_type: "add_to_cart",
        occurred_at: "2026-05-18T00:06:30Z",
        path: "/products/prod_123",
        properties: { product_id: "prod_123", quantity: 1 },
        session_id: "sess_123"
      },
      {
        _id: "evt_3",
        event_type: "payment_result",
        occurred_at: "2026-05-18T00:21:00Z",
        path: "/checkout",
        properties: { order_id: "ord_123", status: "success" },
        session_id: "sess_123"
      }
    ],
    session: {
      anonymous_id: "anon_1234567890",
      device: { browser: "Chrome", os: "Android", type: "mobile" },
      duration_seconds: 1320,
      ended_at: "2026-05-18T00:22:00Z",
      entry_page: "/",
      exit_page: "/checkout",
      geo: { city: "Delhi", country: "India" },
      last_seen_at: "2026-05-18T00:22:00Z",
      session_id: "sess_123",
      started_at: "2026-05-18T00:00:00Z",
      status: "ended",
      user_id: "user_1234567890"
    }
  };
}

function reportScheduleResponse() {
  return {
    created_at: "2026-05-18T12:00:00Z",
    day_of_week: "monday",
    filters: {
      device: "mobile",
      source: "paid",
      user_type: "logged_in"
    },
    format: "csv",
    frequency: "weekly",
    id: "rpt_sch_123",
    last_run_at: "2026-05-25T09:00:00Z",
    name: "Weekly retention report",
    next_run_at: "2026-06-01T09:00:00Z",
    recipients: ["analytics@example.com"],
    report_type: "retention",
    status: "active",
    time_of_day: "09:00",
    timezone: "UTC"
  };
}
