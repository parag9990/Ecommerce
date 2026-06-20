import { useMemo, useState } from "react";
import { Flag, Percent, Search, Wrench } from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { DataState } from "../../../components/ui/data-state";
import { PermissionDenied } from "../../../components/ui/permission-denied";
import { cn } from "../../../lib/classnames";
import { CommissionSettingsCard } from "../components/commission-settings-card";
import { FeatureFlagList } from "../components/feature-flag-list";
import { MaintenanceModePanel } from "../components/maintenance-mode-panel";
import { SearchSynonymTable } from "../components/search-synonym-table";
import { SettingsPageHeader } from "../components/settings-page-header";
import { usePlatformSettings } from "../hooks/use-platform-settings";
import { useSettingsPermissions } from "../permissions";
import type { PlatformSetting, SettingKey } from "../types";

type SettingsTab = "commission" | "synonyms" | "flags" | "maintenance";

const SETTINGS_TABS: Array<{
  id: SettingsTab;
  label: string;
  icon: LucideIcon;
}> = [
  { id: "commission", label: "Commission", icon: Percent },
  { id: "synonyms", label: "Search Synonyms", icon: Search },
  { id: "flags", label: "Feature Flags", icon: Flag },
  { id: "maintenance", label: "Maintenance", icon: Wrench }
];

export function PlatformSettingsPage() {
  const permissions = useSettingsPermissions();
  const [activeTab, setActiveTab] = useState<SettingsTab>("commission");
  const settingsQuery = usePlatformSettings();
  const settingsByKey = useMemo(() => {
    const settingsMap = new Map<SettingKey, PlatformSetting<unknown>>();

    (settingsQuery.data?.settings ?? []).forEach((setting) => {
      settingsMap.set(setting.key, setting);
    });

    return settingsMap;
  }, [settingsQuery.data?.settings]);

  if (!permissions.canViewPlatformSettings) {
    return <PermissionDenied compact />;
  }

  return (
    <section className="flex min-h-[calc(100vh-6.5rem)] flex-col overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
      <SettingsPageHeader canWrite={permissions.canWritePlatformSettings} roles={permissions.roles} />

      {settingsQuery.isLoading ? (
        <div className="p-4">
          <DataState title="Loading platform settings" description="Fetching current settings." />
        </div>
      ) : settingsQuery.isError ? (
        <div className="p-4">
          <DataState
            tone="danger"
            title="Platform settings could not be loaded"
            description={
              settingsQuery.error instanceof Error
                ? settingsQuery.error.message
                : "Settings API request failed."
            }
            action={
              <button
                type="button"
                onClick={() => void settingsQuery.refetch()}
                className="h-9 rounded-lg border border-red-200 bg-white px-3 text-sm font-medium text-red-700"
              >
                Retry
              </button>
            }
          />
        </div>
      ) : (
        <>
          {(settingsQuery.data?.settings.length ?? 0) === 0 ? (
            <div className="border-b border-slate-200 bg-slate-50 p-4">
              <DataState
                title="No platform settings returned"
                description="The page is showing safe defaults until the backend seeds platform settings."
              />
            </div>
          ) : null}

          <div className="border-b border-slate-200 bg-white px-4">
            <div className="flex gap-1 overflow-x-auto py-2">
              {SETTINGS_TABS.map((tab) => {
                const Icon = tab.icon;
                const selected = activeTab === tab.id;

                return (
                  <button
                    key={tab.id}
                    type="button"
                    onClick={() => setActiveTab(tab.id)}
                    className={cn(
                      "inline-flex h-10 shrink-0 items-center gap-2 rounded-lg px-3 text-sm font-medium",
                      selected
                        ? "bg-slate-950 text-white"
                        : "text-slate-600 hover:bg-slate-100 hover:text-slate-950"
                    )}
                  >
                    <Icon className="h-4 w-4" aria-hidden="true" />
                    {tab.label}
                  </button>
                );
              })}
            </div>
          </div>

          <div className="min-h-0 flex-1 overflow-auto p-4">
            {activeTab === "commission" ? (
              <CommissionSettingsCard
                defaultRateSetting={settingsByKey.get("commission.default_rate")}
                overridesSetting={settingsByKey.get("commission.category_overrides")}
                canWrite={permissions.canWritePlatformSettings}
              />
            ) : null}
            {activeTab === "synonyms" ? (
              <SearchSynonymTable canWrite={permissions.canManageSearchSynonyms} />
            ) : null}
            {activeTab === "flags" ? (
              <FeatureFlagList
                setting={settingsByKey.get("platform.feature_flags")}
                canWrite={permissions.canWritePlatformSettings}
              />
            ) : null}
            {activeTab === "maintenance" ? (
              <MaintenanceModePanel
                setting={settingsByKey.get("platform.maintenance_mode")}
                canWrite={permissions.canWritePlatformSettings}
              />
            ) : null}
          </div>
        </>
      )}
    </section>
  );
}
