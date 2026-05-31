import { AlertTriangle, RefreshCcw } from "lucide-react";

import { DeletionRequestPanel } from "../components/deletion-request-panel";
import { DeletionStatusTable } from "../components/deletion-status-table";
import { MaskingPolicyPanel } from "../components/masking-policy-panel";
import { PrivacySummaryCards } from "../components/privacy-summary-cards";
import { RetentionSettingsPanel } from "../components/retention-settings-panel";
import { usePrivacySettings } from "../hooks/use-privacy-settings";
import { useRetentionSettings } from "../hooks/use-retention-settings";

export function PrivacyControlsPage() {
  const privacy = usePrivacySettings();
  const retention = useRetentionSettings();

  if (privacy.isPending || retention.isPending) {
    return <PrivacyLoadingState />;
  }

  if (privacy.isError || retention.isError) {
    return (
      <PrivacyErrorState
        onRetry={() => {
          void privacy.refetch();
          void retention.refetch();
        }}
      />
    );
  }

  return (
    <div className="p-4 lg:p-6">
      <div className="mb-5 flex flex-col gap-3 xl:flex-row xl:items-end xl:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase text-emerald-700">
            Governance
          </p>
          <h1 className="mt-1 text-2xl font-semibold text-zinc-950">
            Privacy controls
          </h1>
          <p className="mt-1 max-w-3xl text-sm text-zinc-500">
            Manage masking policy, session analytics deletion requests, and
            retention limits.
          </p>
        </div>
      </div>

      <div className="space-y-5">
        <PrivacySummaryCards
          privacy={privacy.data}
          retention={retention.data}
        />

        <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_420px]">
          <div className="space-y-5">
            <MaskingPolicyPanel settings={privacy.data} />
            <RetentionSettingsPanel
              canUpdate={privacy.data.permissions.canUpdateRetention}
              settings={retention.data}
            />
          </div>

          <aside className="space-y-5">
            <DeletionRequestPanel
              canRequestDeletion={privacy.data.permissions.canRequestDeletion}
            />
            <DeletionStatusTable />
          </aside>
        </div>
      </div>
    </div>
  );
}

function PrivacyLoadingState() {
  return (
    <div className="space-y-5 p-4 lg:p-6">
      <div className="h-20 animate-pulse rounded-lg bg-zinc-100" />
      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        {Array.from({ length: 4 }).map((_, index) => (
          <div
            aria-hidden="true"
            className="h-24 animate-pulse rounded-lg bg-zinc-100"
            key={index}
          />
        ))}
      </div>
      <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_420px]">
        <div className="h-96 animate-pulse rounded-lg bg-zinc-100" />
        <div className="h-96 animate-pulse rounded-lg bg-zinc-100" />
      </div>
    </div>
  );
}

function PrivacyErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="p-4 lg:p-6">
      <section className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-900">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex gap-3">
            <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
            <div>
              <h1 className="text-sm font-semibold">
                Privacy controls unavailable
              </h1>
              <p className="mt-1 text-sm text-red-800">
                The analytics privacy settings could not be loaded.
              </p>
            </div>
          </div>
          <button
            className="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-red-300 bg-white px-3 text-sm font-medium text-red-900 transition-colors hover:bg-red-100"
            onClick={onRetry}
            type="button"
          >
            <RefreshCcw className="h-4 w-4" aria-hidden="true" />
            Retry
          </button>
        </div>
      </section>
    </div>
  );
}
