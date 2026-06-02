import { useEffect, useMemo, useState } from "react";
import { Plus, RotateCcw, Save, Trash2 } from "lucide-react";

import { ActionReasonDialog } from "../../../components/ui/action-reason-dialog";
import { DataState } from "../../../components/ui/data-state";
import { formatDateTime } from "../../../lib/format";
import {
  readCommissionCategoryOverrides,
  readCommissionDefaultRate,
  settingConfigured
} from "../setting-values";
import { useUpdatePlatformSetting } from "../hooks/use-update-platform-setting";
import type {
  CommissionCategoryOverride,
  CommissionCategoryOverridesValue,
  CommissionDefaultRate,
  PlatformSetting
} from "../types";
import {
  normalizeCommissionOverrides,
  validateCommissionCategoryOverrides,
  validateCommissionDefaultRate
} from "../validators";
import { SettingChangeSummary } from "./setting-change-summary";

type OverrideRow = {
  id: string;
  category_id: string;
  rate_percent: string;
};

type PendingAction = "default" | "overrides" | null;

function serialize(value: unknown): string {
  return JSON.stringify(value);
}

function toRateInput(rate: number): string {
  return Number.isFinite(rate) ? String(rate) : "";
}

function rowsFromOverrides(overrides: readonly CommissionCategoryOverride[]): OverrideRow[] {
  return overrides.map((override, index) => ({
    id: `${override.category_id}-${index}`,
    category_id: override.category_id,
    rate_percent: toRateInput(override.rate_percent)
  }));
}

function rowsToOverrides(rows: readonly OverrideRow[]): CommissionCategoryOverridesValue {
  return {
    overrides: rows.map((row) => ({
      category_id: row.category_id.trim(),
      rate_percent: row.rate_percent.trim() ? Number(row.rate_percent) : Number.NaN
    }))
  };
}

export function CommissionSettingsCard({
  defaultRateSetting,
  overridesSetting,
  canWrite
}: {
  defaultRateSetting?: PlatformSetting<unknown>;
  overridesSetting?: PlatformSetting<unknown>;
  canWrite: boolean;
}) {
  const currentDefaultRate = useMemo(
    () => readCommissionDefaultRate(defaultRateSetting),
    [defaultRateSetting]
  );
  const currentOverrides = useMemo(
    () => readCommissionCategoryOverrides(overridesSetting),
    [overridesSetting]
  );
  const [rateInput, setRateInput] = useState(() => toRateInput(currentDefaultRate.rate_percent));
  const [overrideRows, setOverrideRows] = useState<OverrideRow[]>(() =>
    rowsFromOverrides(currentOverrides.overrides)
  );
  const [pendingAction, setPendingAction] = useState<PendingAction>(null);
  const updateSetting = useUpdatePlatformSetting<CommissionDefaultRate | CommissionCategoryOverridesValue>();

  useEffect(() => {
    setRateInput(toRateInput(currentDefaultRate.rate_percent));
  }, [currentDefaultRate.rate_percent]);

  useEffect(() => {
    setOverrideRows(rowsFromOverrides(currentOverrides.overrides));
  }, [currentOverrides]);

  const nextDefaultRate: CommissionDefaultRate = {
    rate_percent: rateInput.trim() ? Number(rateInput) : Number.NaN,
    applies_to: "all_sellers"
  };
  const nextOverrides = rowsToOverrides(overrideRows);
  const defaultValidation = validateCommissionDefaultRate(nextDefaultRate);
  const overrideValidation = validateCommissionCategoryOverrides(nextOverrides);
  const defaultDirty = serialize(nextDefaultRate) !== serialize(currentDefaultRate);
  const overridesDirty =
    serialize(normalizeCommissionOverrides(nextOverrides.overrides)) !==
    serialize(normalizeCommissionOverrides(currentOverrides.overrides));
  const activeValidation = pendingAction === "default" ? defaultValidation : overrideValidation;

  function resetDefaultRate() {
    setRateInput(toRateInput(currentDefaultRate.rate_percent));
  }

  function resetOverrides() {
    setOverrideRows(rowsFromOverrides(currentOverrides.overrides));
  }

  function addOverride() {
    setOverrideRows((rows) => [
      ...rows,
      {
        id: crypto.randomUUID(),
        category_id: "",
        rate_percent: ""
      }
    ]);
  }

  function updateOverride(rowId: string, patch: Partial<Omit<OverrideRow, "id">>) {
    setOverrideRows((rows) =>
      rows.map((row) => (row.id === rowId ? { ...row, ...patch } : row))
    );
  }

  function removeOverride(rowId: string) {
    setOverrideRows((rows) => rows.filter((row) => row.id !== rowId));
  }

  async function submitChange({ reason }: { reason: string }) {
    if (!pendingAction || !activeValidation.valid) {
      return;
    }

    try {
      if (pendingAction === "default") {
        await updateSetting.mutateAsync({
          key: "commission.default_rate",
          input: {
            value: nextDefaultRate,
            reason
          }
        });
      } else {
        await updateSetting.mutateAsync({
          key: "commission.category_overrides",
          input: {
            value: normalizeCommissionOverrides(nextOverrides.overrides),
            reason
          }
        });
      }

      setPendingAction(null);
    } catch {
      // ActionReasonDialog receives mutation.error and keeps the form values intact.
    }
  }

  return (
    <section className="space-y-4">
      {!canWrite ? (
        <DataState
          title="Read-only commission settings"
          description="Only superadmins can update commission rates."
        />
      ) : null}

      <div className="rounded-lg border border-slate-200 bg-white p-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-base font-semibold text-slate-950">Default Commission</h2>
              {!settingConfigured(defaultRateSetting) ? (
                <span className="rounded-full border border-amber-200 bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-800">
                  Not configured
                </span>
              ) : null}
            </div>
            <p className="mt-1 text-sm text-slate-600">
              Platform-wide commission applied when no category override exists.
            </p>
            <p className="mt-1 text-xs text-slate-500">
              Last updated {formatDateTime(defaultRateSetting?.updated_at)}
            </p>
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={resetDefaultRate}
              disabled={!defaultDirty || updateSetting.isPending}
              className="inline-flex h-9 items-center justify-center gap-2 rounded-lg border border-slate-300 px-3 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <RotateCcw className="h-4 w-4" aria-hidden="true" />
              Reset
            </button>
            <button
              type="button"
              onClick={() => {
                updateSetting.reset();
                setPendingAction("default");
              }}
              disabled={!canWrite || !defaultDirty || !defaultValidation.valid || updateSetting.isPending}
              className="inline-flex h-9 items-center justify-center gap-2 rounded-lg bg-slate-950 px-3 text-sm font-medium text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
            >
              <Save className="h-4 w-4" aria-hidden="true" />
              Save
            </button>
          </div>
        </div>

        <div className="mt-4 grid gap-4 lg:grid-cols-[240px_1fr]">
          <label className="block">
            <span className="text-sm font-medium text-slate-700">Rate percent</span>
            <div className="mt-1 flex h-10 items-center rounded-lg border border-slate-300 bg-white px-3 focus-within:border-slate-900 focus-within:ring-2 focus-within:ring-slate-900/10">
              <input
                type="number"
                min={0}
                max={50}
                step={0.1}
                value={rateInput}
                disabled={!canWrite || updateSetting.isPending}
                onChange={(event) => setRateInput(event.target.value)}
                className="min-w-0 flex-1 border-0 bg-transparent text-sm text-slate-950 outline-none disabled:text-slate-500"
              />
              <span className="text-sm text-slate-500">%</span>
            </div>
          </label>

          <SettingChangeSummary
            title="Default rate change"
            items={[
              { label: "Current", value: `${currentDefaultRate.rate_percent}%` },
              {
                label: "Next",
                value: Number.isFinite(nextDefaultRate.rate_percent)
                  ? `${nextDefaultRate.rate_percent}%`
                  : "Invalid"
              }
            ]}
          />
        </div>

        {!defaultValidation.valid ? (
          <p className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {defaultValidation.errors.join(" ")}
          </p>
        ) : null}
      </div>

      <div className="rounded-lg border border-slate-200 bg-white p-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-base font-semibold text-slate-950">Category Overrides</h2>
              {!settingConfigured(overridesSetting) ? (
                <span className="rounded-full border border-amber-200 bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-800">
                  Not configured
                </span>
              ) : null}
            </div>
            <p className="mt-1 text-sm text-slate-600">
              Override the default commission for selected category ids.
            </p>
            <p className="mt-1 text-xs text-slate-500">
              Last updated {formatDateTime(overridesSetting?.updated_at)}
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={addOverride}
              disabled={!canWrite || updateSetting.isPending}
              className="inline-flex h-9 items-center justify-center gap-2 rounded-lg border border-slate-300 px-3 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <Plus className="h-4 w-4" aria-hidden="true" />
              Add
            </button>
            <button
              type="button"
              onClick={resetOverrides}
              disabled={!overridesDirty || updateSetting.isPending}
              className="inline-flex h-9 items-center justify-center gap-2 rounded-lg border border-slate-300 px-3 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <RotateCcw className="h-4 w-4" aria-hidden="true" />
              Reset
            </button>
            <button
              type="button"
              onClick={() => {
                updateSetting.reset();
                setPendingAction("overrides");
              }}
              disabled={!canWrite || !overridesDirty || !overrideValidation.valid || updateSetting.isPending}
              className="inline-flex h-9 items-center justify-center gap-2 rounded-lg bg-slate-950 px-3 text-sm font-medium text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
            >
              <Save className="h-4 w-4" aria-hidden="true" />
              Save
            </button>
          </div>
        </div>

        <div className="mt-4 overflow-x-auto">
          <table className="min-w-full text-left text-sm">
            <thead className="text-xs uppercase text-slate-500">
              <tr>
                <th className="w-2/3 border-b border-slate-200 pb-2 font-medium">Category id</th>
                <th className="w-36 border-b border-slate-200 pb-2 font-medium">Rate</th>
                <th className="w-16 border-b border-slate-200 pb-2 text-right font-medium">Action</th>
              </tr>
            </thead>
            <tbody>
              {overrideRows.length === 0 ? (
                <tr>
                  <td colSpan={3} className="py-4 text-sm text-slate-500">
                    No category overrides configured.
                  </td>
                </tr>
              ) : (
                overrideRows.map((row) => (
                  <tr key={row.id}>
                    <td className="border-b border-slate-100 py-2 pr-3">
                      <label>
                        <span className="sr-only">Category id</span>
                        <input
                          value={row.category_id}
                          disabled={!canWrite || updateSetting.isPending}
                          onChange={(event) =>
                            updateOverride(row.id, { category_id: event.target.value })
                          }
                          placeholder="cat_electronics"
                          className="h-9 w-full rounded-lg border border-slate-300 px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10 disabled:bg-slate-50 disabled:text-slate-500"
                        />
                      </label>
                    </td>
                    <td className="border-b border-slate-100 py-2 pr-3">
                      <label>
                        <span className="sr-only">Override rate percent</span>
                        <div className="flex h-9 items-center rounded-lg border border-slate-300 bg-white px-3 focus-within:border-slate-900 focus-within:ring-2 focus-within:ring-slate-900/10">
                          <input
                            type="number"
                            min={0}
                            max={50}
                            step={0.1}
                            value={row.rate_percent}
                            disabled={!canWrite || updateSetting.isPending}
                            onChange={(event) =>
                              updateOverride(row.id, { rate_percent: event.target.value })
                            }
                            className="min-w-0 flex-1 border-0 bg-transparent text-sm text-slate-950 outline-none disabled:text-slate-500"
                          />
                          <span className="text-sm text-slate-500">%</span>
                        </div>
                      </label>
                    </td>
                    <td className="border-b border-slate-100 py-2 text-right">
                      <button
                        type="button"
                        onClick={() => removeOverride(row.id)}
                        disabled={!canWrite || updateSetting.isPending}
                        className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-300 text-slate-600 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
                        aria-label={`Remove override ${row.category_id || "row"}`}
                      >
                        <Trash2 className="h-4 w-4" aria-hidden="true" />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {!overrideValidation.valid ? (
          <p className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {overrideValidation.errors.join(" ")}
          </p>
        ) : null}
      </div>

      <ActionReasonDialog
        open={pendingAction !== null}
        title={pendingAction === "default" ? "Save default commission" : "Save category overrides"}
        description="Confirm the commission setting change and provide an audit reason."
        confirmLabel="Confirm update"
        isSubmitting={updateSetting.isPending}
        error={updateSetting.error}
        onCancel={() => {
          if (!updateSetting.isPending) {
            setPendingAction(null);
            updateSetting.reset();
          }
        }}
        onConfirm={submitChange}
      >
        <SettingChangeSummary
          title="Change summary"
          items={
            pendingAction === "default"
              ? [
                  { label: "Setting", value: "commission.default_rate" },
                  { label: "Current", value: `${currentDefaultRate.rate_percent}%` },
                  {
                    label: "Next",
                    value: Number.isFinite(nextDefaultRate.rate_percent)
                      ? `${nextDefaultRate.rate_percent}%`
                      : "Invalid"
                  }
                ]
              : [
                  { label: "Setting", value: "commission.category_overrides" },
                  { label: "Current rows", value: String(currentOverrides.overrides.length) },
                  { label: "Next rows", value: String(nextOverrides.overrides.length) }
                ]
          }
        />
      </ActionReasonDialog>
    </section>
  );
}
