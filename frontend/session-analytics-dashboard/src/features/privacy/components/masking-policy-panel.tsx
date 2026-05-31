import { AlertTriangle, Save } from "lucide-react";
import { useEffect, useState } from "react";

import {
  locationGranularityOptions,
  maskingModeOptions,
  type LocationGranularity,
  type MaskingMode,
  type PrivacyMaskingSettings,
  type PrivacySettingsResponse
} from "../../../api/session-api";
import { formatDateTime } from "../../../lib/format";
import { useUpdatePrivacySettings } from "../hooks/use-privacy-settings";

type MaskingPolicyPanelProps = {
  settings: PrivacySettingsResponse;
};

export function MaskingPolicyPanel({ settings }: MaskingPolicyPanelProps) {
  const [draft, setDraft] = useState<PrivacyMaskingSettings>(settings.masking);
  const update = useUpdatePrivacySettings();
  const disabled = !settings.permissions.canUpdateMasking || update.isPending;

  useEffect(() => {
    setDraft(settings.masking);
  }, [settings.masking]);

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase text-emerald-700">
            Mask by default
          </p>
          <h2 className="mt-1 text-base font-semibold text-zinc-950">
            PII masking policy
          </h2>
          <p className="mt-1 text-sm text-zinc-500">
            Default dashboard visibility for identity-like analytics fields.
          </p>
        </div>
        <p className="text-xs text-zinc-500">
          Updated {formatDateTime(settings.updatedAt)}
        </p>
      </div>

      {!settings.permissions.canUpdateMasking ? (
        <PermissionNote message="You can view masking settings, but only operations admins and superadmins can update them." />
      ) : null}

      <div className="mt-4 grid gap-4 md:grid-cols-3">
        <MaskingSelect
          disabled={disabled}
          label="User ID"
          onChange={(userIdMode) => setDraft({ ...draft, userIdMode })}
          value={draft.userIdMode}
        />
        <MaskingSelect
          disabled={disabled}
          label="Anonymous ID"
          onChange={(anonymousIdMode) =>
            setDraft({ ...draft, anonymousIdMode })
          }
          value={draft.anonymousIdMode}
        />
        <MaskingSelect
          disabled={disabled}
          label="Session ID"
          onChange={(sessionIdMode) => setDraft({ ...draft, sessionIdMode })}
          value={draft.sessionIdMode}
        />
      </div>

      <div className="mt-4 grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <label className="text-sm font-medium text-zinc-700">
          Location granularity
          <select
            className="mt-1 h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10 disabled:cursor-not-allowed disabled:bg-zinc-100"
            disabled={disabled}
            onChange={(event) =>
              setDraft({
                ...draft,
                locationGranularity: event.target.value as LocationGranularity
              })
            }
            value={draft.locationGranularity}
          >
            {locationGranularityOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>

        <div className="space-y-2 text-sm text-zinc-700">
          <BooleanControl
            checked={draft.showSearchQueries}
            disabled={disabled}
            label="Show raw search queries"
            onChange={(showSearchQueries) =>
              setDraft({ ...draft, showSearchQueries })
            }
          />
          <BooleanControl
            checked={draft.showIpHash}
            disabled={disabled}
            label="Show IP hash to privileged admins"
            onChange={(showIpHash) => setDraft({ ...draft, showIpHash })}
          />
        </div>
      </div>

      {draft.userIdMode === "full" ||
      draft.anonymousIdMode === "full" ||
      draft.sessionIdMode === "full" ||
      draft.showIpHash ? (
        <div className="mt-4 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
          <div className="flex gap-2">
            <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
            <p>
              Full identifiers and IP hashes should stay limited to audited
              operational investigations.
            </p>
          </div>
        </div>
      ) : null}

      {update.isError ? (
        <div className="mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          Masking policy could not be saved. Check permissions and retry.
        </div>
      ) : null}

      <div className="mt-4 flex justify-end">
        <button
          className="inline-flex h-10 items-center gap-2 rounded-md bg-zinc-950 px-4 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:cursor-not-allowed disabled:bg-zinc-400"
          disabled={disabled}
          onClick={() => update.mutate(draft)}
          type="button"
        >
          <Save className="h-4 w-4" aria-hidden="true" />
          {update.isPending ? "Saving" : "Save masking policy"}
        </button>
      </div>
    </section>
  );
}

function MaskingSelect({
  disabled,
  label,
  onChange,
  value
}: {
  disabled: boolean;
  label: string;
  onChange: (value: MaskingMode) => void;
  value: MaskingMode;
}) {
  return (
    <label className="text-sm font-medium text-zinc-700">
      {label}
      <select
        className="mt-1 h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10 disabled:cursor-not-allowed disabled:bg-zinc-100"
        disabled={disabled}
        onChange={(event) => onChange(event.target.value as MaskingMode)}
        value={value}
      >
        {maskingModeOptions.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

function BooleanControl({
  checked,
  disabled,
  label,
  onChange
}: {
  checked: boolean;
  disabled: boolean;
  label: string;
  onChange: (checked: boolean) => void;
}) {
  return (
    <label className="flex items-center gap-2">
      <input
        checked={checked}
        className="h-4 w-4 rounded border-zinc-300 text-zinc-950 focus:ring-zinc-950"
        disabled={disabled}
        onChange={(event) => onChange(event.target.checked)}
        type="checkbox"
      />
      {label}
    </label>
  );
}

function PermissionNote({ message }: { message: string }) {
  return (
    <div className="mt-4 rounded-md border border-sky-200 bg-sky-50 px-3 py-2 text-sm text-sky-900">
      {message}
    </div>
  );
}
