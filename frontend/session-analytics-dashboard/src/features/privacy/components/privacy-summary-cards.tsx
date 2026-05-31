import { Clock3, DatabaseZap, MapPin, ShieldCheck } from "lucide-react";
import type { LucideIcon } from "lucide-react";

import type {
  PrivacySettingsResponse,
  RetentionSettings
} from "../../../api/session-api";
import {
  locationGranularityLabel,
  maskingModeLabel
} from "../lib/privacy-format";

type PrivacySummaryCardsProps = {
  privacy: PrivacySettingsResponse;
  retention: RetentionSettings;
};

export function PrivacySummaryCards({
  privacy,
  retention
}: PrivacySummaryCardsProps) {
  return (
    <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
      <SummaryCard
        icon={ShieldCheck}
        label="Identity display"
        value={`Users ${maskingModeLabel(privacy.masking.userIdMode).toLowerCase()}`}
      />
      <SummaryCard
        icon={MapPin}
        label="Location"
        value={locationGranularityLabel(privacy.masking.locationGranularity)}
      />
      <SummaryCard
        icon={DatabaseZap}
        label="Raw event retention"
        value={`${retention.rawEventsDays} days`}
      />
      <SummaryCard
        icon={Clock3}
        label="Deletion log"
        value={`${retention.deletionRequestLogDays} days`}
      />
    </div>
  );
}

function SummaryCard({
  icon: Icon,
  label,
  value
}: {
  icon: LucideIcon;
  label: string;
  value: string;
}) {
  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="flex items-center gap-3">
        <div className="flex h-9 w-9 items-center justify-center rounded-md bg-emerald-50 text-emerald-700">
          <Icon className="h-4 w-4" aria-hidden="true" />
        </div>
        <div className="min-w-0">
          <p className="text-xs font-medium uppercase text-zinc-500">{label}</p>
          <p className="mt-1 truncate text-sm font-semibold text-zinc-950">
            {value}
          </p>
        </div>
      </div>
    </section>
  );
}
