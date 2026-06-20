import { useEffect, useMemo, useState } from "react";
import { Plus, RotateCcw, Save, Trash2 } from "lucide-react";

import { ActionReasonDialog } from "../../../components/ui/action-reason-dialog";
import { DataState } from "../../../components/ui/data-state";
import { formatDateTime } from "../../../lib/format";
import { readFeatureFlags, settingConfigured } from "../setting-values";
import { useUpdatePlatformSetting } from "../hooks/use-update-platform-setting";
import type { FeatureFlagsValue, PlatformSetting } from "../types";
import { FEATURE_FLAG_KEY_PATTERN, validateFeatureFlags } from "../validators";
import { SettingChangeSummary } from "./setting-change-summary";

function serializeFlags(flags: Record<string, boolean>): string {
  return JSON.stringify(
    Object.keys(flags)
      .sort()
      .reduce<Record<string, boolean>>((accumulator, key) => {
        accumulator[key] = flags[key];
        return accumulator;
      }, {})
  );
}

export function FeatureFlagList({
  setting,
  canWrite
}: {
  setting?: PlatformSetting<unknown>;
  canWrite: boolean;
}) {
  const currentValue = useMemo(() => readFeatureFlags(setting), [setting]);
  const [flags, setFlags] = useState<Record<string, boolean>>(() => currentValue.flags);
  const [newFlagKey, setNewFlagKey] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const updateSetting = useUpdatePlatformSetting<FeatureFlagsValue>();

  useEffect(() => {
    setFlags(currentValue.flags);
  }, [currentValue]);

  const validation = validateFeatureFlags({ flags });
  const dirty = serializeFlags(flags) !== serializeFlags(currentValue.flags);
  const enabledCount = Object.values(flags).filter(Boolean).length;
  const sortedEntries = Object.entries(flags).sort(([left], [right]) => left.localeCompare(right));
  const normalizedNewFlagKey = newFlagKey.trim();
  const canAddFlag =
    FEATURE_FLAG_KEY_PATTERN.test(normalizedNewFlagKey) &&
    !Object.prototype.hasOwnProperty.call(flags, normalizedNewFlagKey);

  function toggleFlag(flagKey: string) {
    setFlags((current) => ({
      ...current,
      [flagKey]: !current[flagKey]
    }));
  }

  function addFlag() {
    if (!canAddFlag) {
      return;
    }

    setFlags((current) => ({
      ...current,
      [normalizedNewFlagKey]: false
    }));
    setNewFlagKey("");
  }

  function removeFlag(flagKey: string) {
    setFlags((current) => {
      const next = { ...current };
      delete next[flagKey];
      return next;
    });
  }

  async function submitChange({ reason }: { reason: string }) {
    if (!validation.valid) {
      return;
    }

    try {
      await updateSetting.mutateAsync({
        key: "platform.feature_flags",
        input: {
          value: { flags },
          reason
        }
      });
      setDialogOpen(false);
    } catch {
      // ActionReasonDialog displays mutation.error while preserving local edits.
    }
  }

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-base font-semibold text-slate-950">Feature Flags</h2>
            {!settingConfigured(setting) ? (
              <span className="rounded-full border border-amber-200 bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-800">
                Not configured
              </span>
            ) : null}
          </div>
          <p className="mt-1 text-sm text-slate-600">
            Enable or disable controlled platform rollout switches.
          </p>
          <p className="mt-1 text-xs text-slate-500">
            {enabledCount} enabled of {sortedEntries.length} flags. Last updated{" "}
            {formatDateTime(setting?.updated_at)}
          </p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => setFlags(currentValue.flags)}
            disabled={!dirty || updateSetting.isPending}
            className="inline-flex h-9 items-center justify-center gap-2 rounded-lg border border-slate-300 px-3 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <RotateCcw className="h-4 w-4" aria-hidden="true" />
            Reset
          </button>
          <button
            type="button"
            onClick={() => {
              updateSetting.reset();
              setDialogOpen(true);
            }}
            disabled={!canWrite || !dirty || !validation.valid || updateSetting.isPending}
            className="inline-flex h-9 items-center justify-center gap-2 rounded-lg bg-slate-950 px-3 text-sm font-medium text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
          >
            <Save className="h-4 w-4" aria-hidden="true" />
            Save
          </button>
        </div>
      </div>

      {!canWrite ? (
        <div className="mt-4">
          <DataState
            title="Read-only feature flags"
            description="Only superadmins can update feature flag state."
          />
        </div>
      ) : null}

      <div className="mt-4 grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto]">
        <label>
          <span className="sr-only">New feature flag key</span>
          <input
            value={newFlagKey}
            disabled={!canWrite || updateSetting.isPending}
            onChange={(event) => setNewFlagKey(event.target.value)}
            placeholder="new_feature_flag"
            className="h-10 w-full rounded-lg border border-slate-300 px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10 disabled:bg-slate-50 disabled:text-slate-500"
          />
        </label>
        <button
          type="button"
          onClick={addFlag}
          disabled={!canWrite || !canAddFlag || updateSetting.isPending}
          className="inline-flex h-10 items-center justify-center gap-2 rounded-lg border border-slate-300 px-3 text-sm font-medium text-slate-800 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
        >
          <Plus className="h-4 w-4" aria-hidden="true" />
          Add flag
        </button>
      </div>
      <p className="mt-2 text-xs text-slate-500">Flag keys must use lowercase snake_case.</p>

      <div className="mt-4 divide-y divide-slate-100 rounded-lg border border-slate-200">
        {sortedEntries.length === 0 ? (
          <div className="p-4 text-sm text-slate-500">No feature flags configured.</div>
        ) : (
          sortedEntries.map(([flagKey, enabled]) => (
            <div
              key={flagKey}
              className="flex flex-col gap-3 p-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <div className="min-w-0">
                <div className="break-words text-sm font-medium text-slate-950">{flagKey}</div>
                <span
                  className={
                    enabled
                      ? "mt-1 inline-flex rounded-full border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700"
                      : "mt-1 inline-flex rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-xs font-medium text-slate-600"
                  }
                >
                  {enabled ? "Enabled" : "Disabled"}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <label className="inline-flex h-9 items-center gap-2 rounded-lg border border-slate-300 px-3 text-sm font-medium text-slate-700">
                  <input
                    type="checkbox"
                    checked={enabled}
                    disabled={!canWrite || updateSetting.isPending}
                    onChange={() => toggleFlag(flagKey)}
                    className="h-4 w-4"
                  />
                  Toggle
                </label>
                <button
                  type="button"
                  onClick={() => removeFlag(flagKey)}
                  disabled={!canWrite || updateSetting.isPending}
                  className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-300 text-slate-600 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
                  aria-label={`Remove flag ${flagKey}`}
                >
                  <Trash2 className="h-4 w-4" aria-hidden="true" />
                </button>
              </div>
            </div>
          ))
        )}
      </div>

      {!validation.valid ? (
        <p className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {validation.errors.join(" ")}
        </p>
      ) : null}

      <ActionReasonDialog
        open={dialogOpen}
        title="Update feature flags"
        description="Confirm the feature flag change and provide an audit reason."
        confirmLabel="Confirm update"
        isSubmitting={updateSetting.isPending}
        error={updateSetting.error}
        onCancel={() => {
          if (!updateSetting.isPending) {
            setDialogOpen(false);
            updateSetting.reset();
          }
        }}
        onConfirm={submitChange}
      >
        <SettingChangeSummary
          title="Change summary"
          items={[
            { label: "Setting", value: "platform.feature_flags" },
            { label: "Current enabled", value: String(Object.values(currentValue.flags).filter(Boolean).length) },
            { label: "Next enabled", value: String(enabledCount) },
            { label: "Total flags", value: String(sortedEntries.length) }
          ]}
        />
      </ActionReasonDialog>
    </section>
  );
}
