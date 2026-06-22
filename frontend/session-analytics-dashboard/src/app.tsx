import { lazy, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import { AnalyticsLayout } from "./layout/analytics-layout";
import { AdminLoginPage } from "./features/auth/pages/admin-login-page";
import { RequireAdmin } from "./features/auth/components/require-admin";

const AnalyticsOverviewPage = lazy(() =>
  import("./features/shell/pages/analytics-overview-page").then((module) => ({
    default: module.AnalyticsOverviewPage
  }))
);
const LiveSessionsPage = lazy(() =>
  import("./features/live/pages/live-sessions-page").then((module) => ({
    default: module.LiveSessionsPage
  }))
);
const JourneyExplorerPage = lazy(() =>
  import("./features/journey/pages/journey-explorer-page").then((module) => ({
    default: module.JourneyExplorerPage
  }))
);
const FunnelAnalysisPage = lazy(() =>
  import("./features/funnels/pages/funnel-analysis-page").then((module) => ({
    default: module.FunnelAnalysisPage
  }))
);
const HeatmapPage = lazy(() =>
  import("./features/heatmaps/pages/heatmap-page").then((module) => ({
    default: module.HeatmapPage
  }))
);
const CohortRetentionPage = lazy(() =>
  import("./features/cohorts/pages/cohort-retention-page").then((module) => ({
    default: module.CohortRetentionPage
  }))
);
const ReportsExportPage = lazy(() =>
  import("./features/reports/pages/reports-export-page").then((module) => ({
    default: module.ReportsExportPage
  }))
);
const PrivacyControlsPage = lazy(() =>
  import("./features/privacy/pages/privacy-controls-page").then((module) => ({
    default: module.PrivacyControlsPage
  }))
);

export function App() {
  return (
    <Suspense fallback={<div className="p-6 text-sm text-zinc-600">Loading analytics...</div>}>
      <Routes>
        <Route path="login" element={<AdminLoginPage />} />
        <Route element={<RequireAdmin />}>
          <Route element={<AnalyticsLayout />}>
            <Route index element={<AnalyticsOverviewPage />} />
            <Route path="live" element={<LiveSessionsPage />} />
            <Route path="journey" element={<JourneyExplorerPage />} />
            <Route path="journey/:sessionId" element={<JourneyExplorerPage />} />
            <Route path="funnels" element={<FunnelAnalysisPage />} />
            <Route path="heatmaps" element={<HeatmapPage />} />
            <Route path="cohorts" element={<CohortRetentionPage />} />
            <Route path="reports" element={<ReportsExportPage />} />
            <Route path="privacy" element={<PrivacyControlsPage />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Suspense>
  );
}
