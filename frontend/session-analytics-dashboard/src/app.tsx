import { Navigate, Route, Routes } from "react-router-dom";

import { CohortRetentionPage } from "./features/cohorts/pages/cohort-retention-page";
import { FunnelAnalysisPage } from "./features/funnels/pages/funnel-analysis-page";
import { HeatmapPage } from "./features/heatmaps/pages/heatmap-page";
import { JourneyExplorerPage } from "./features/journey/pages/journey-explorer-page";
import { LiveSessionsPage } from "./features/live/pages/live-sessions-page";
import { PrivacyControlsPage } from "./features/privacy/pages/privacy-controls-page";
import { ReportsExportPage } from "./features/reports/pages/reports-export-page";
import { AnalyticsOverviewPage } from "./features/shell/pages/analytics-overview-page";
import { AnalyticsLayout } from "./layout/analytics-layout";

export function App() {
  return (
    <Routes>
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
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
