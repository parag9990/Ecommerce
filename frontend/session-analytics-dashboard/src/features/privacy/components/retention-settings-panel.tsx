import { Save } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import type { RetentionSettings } from "../../../api/session-api";
import { useUpdateRetentionSettings } from "../hooks/use-retention-settings";
import {
  validateDeletionReason,
  validateRetentionSettings
} from "../lib/privacy-validation";
import { PrivacyAuditNote } from "./privacy-audit-note";

type RetentionSettingsPanelProps = {
  canUpdate: boolean;
  settings: RetentionSettings;
};

export function RetentionSettingsPanel({
  canUpdate,
  settings
}: RetentionSettingsPanelProps) {
  const [draft, setDraft] = useState(settings);
  const [reason, setReason] = useState("");
  const update = useUpdateRetentionSettings();
  const errors = useMemo(() => validateRetentionSettings(draft), [draft]);
  const reasonError = reason ? validateDeletionReason(reason) : undefined;
  const hasErrors = Object.keys(errors).length > 0;
  const disabled = !canUpdate || update.isPending;

  useEffect(() => {
    setDraft(settings);
  }, [settings]);

  function updateNumber(key: keyof RetentionSettings, value: string) {
    setDraft({ ...draft, [key]: Number(value) });
  }

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div>
        <p className="text-xs font-semibold uppercase text-emerald-700">
          Retention
        </p>
        <h2 className="mt-1 text-base font-semibold text-zinc-950">
          Retention settings
        </h2>
        <p className="mt-1 text-sm text-zinc-500">
          Control how long session analytics data remains queryable.
        </p>
      </div>

      {!canUpdate ? (
        <div className="mt-4 rounded-md border border-sky-200 bg-sky-50 px-3 py-2 text-sm text-sky-900">
          Retention updates are restricted to superadmins.
        </div>
      ) : null}

      <div className="mt-4 grid gap-4 md:grid-cols-2">
        <RetentionInput
          disabled={disabled}
          error={errors.rawEventsDays}
          label="Raw events days"
          onChange={(value) => updateNumber("rawEventsDays", value)}
          value={draft.rawEventsDays}
        />
        <RetentionInput
          disabled={disabled}
          error={errors.journeySummariesDays}
          label="Journey summaries days"
          onChange={(value) => updateNumber("journeySummariesDays", value)}
          value={draft.journeySummariesDays}
        />
        <RetentionInput
          disabled={disabled}
          error={errors.heatmapAggregatesDays}
          label="Heatmap aggregates days"
          onChange={(value) => updateNumber("heatmapAggregatesDays", value)}
          value={draft.heatmapAggregatesDays}
        />
        <RetentionInput
          disabled={disabled}
          error={errors.analyticsAggregatesMonths}
          label="Analytics aggregates months"
          onChange={(value) => updateNumber("analyticsAggregatesMonths", value)}
          value={draft.analyticsAggregatesMonths}
        />
        <RetentionInput
          disabled={disabled}
          error={errors.activeSessionTtlMinutes}
          label="Active session TTL minutes"
          onChange={(value) => updateNumber("activeSessionTtlMinutes", value)}
          value={draft.activeSessionTtlMinutes}
        />
        <RetentionInput
          disabled={disabled}
          error={errors.deletionRequestLogDays}
          label="Deletion request log days"
          onChange={(value) => updateNumber("deletionRequestLogDays", value)}
          value={draft.deletionRequestLogDays}
        />
      </div>

      <div className="mt-4">
        <PrivacyAuditNote
          disabled={disabled}
          error={reasonError}
          id="retention-reason"
          onChange={setReason}
          placeholder="Why this retention policy is changing"
          value={reason}
        />
      </div>

      {update.isError ? (
        <div className="mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          Retention settings could not be saved. Check permissions and retry.
        </div>
      ) : null}

      <div className="mt-4 flex justify-end">
        <button
          className="inline-flex h-10 items-center gap-2 rounded-md bg-zinc-950 px-4 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:cursor-not-allowed disabled:bg-zinc-400"
          disabled={
            disabled ||
            hasErrors ||
            Boolean(validateDeletionReason(reason))
          }
          onClick={() =>
            update.mutate(
              { reason, settings: draft },
              { onSuccess: () => setReason("") }
            )
          }
          type="button"
        >
          <Save className="h-4 w-4" aria-hidden="true" />
          {update.isPending ? "Saving" : "Save retention settings"}
        </button>
      </div>
    </section>
  );
}

function RetentionInput({
  disabled,
  error,
  label,
  onChange,
  value
}: {
  disabled: boolean;
  error?: string;
  label: string;
  onChange: (value: string) => void;
  value: number;
}) {
  return (
    <label className="text-sm font-medium text-zinc-700">
      {label}
      <input
        className={[
          "mt-1 h-10 w-full rounded-md border bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10 disabled:cursor-not-allowed disabled:bg-zinc-100",
          error ? "border-red-300" : "border-zinc-300"
        ].join(" ")}
        disabled={disabled}
        min={1}
        onChange={(event) => onChange(event.target.value)}
        type="number"
        value={value}
      />
      {error ? <span className="mt-1 block text-xs text-red-600">{error}</span> : null}
    </label>
  );
}
