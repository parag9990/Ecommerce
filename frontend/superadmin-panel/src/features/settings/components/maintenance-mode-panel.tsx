import { useEffect, useMemo, useState } from "react";
import { AlertTriangle, RotateCcw, Save } from "lucide-react";

import { ActionReasonDialog } from "../../../components/ui/action-reason-dialog";
import { DataState } from "../../../components/ui/data-state";
import { cn } from "../../../lib/classnames";
import { formatDateTime } from "../../../lib/format";
import { readMaintenanceMode, settingConfigured } from "../setting-values";
import { useUpdatePlatformSetting } from "../hooks/use-update-platform-setting";
import type { MaintenanceModeStatus, MaintenanceModeValue, PlatformSetting } from "../types";
import { getMaintenanceModeStatus, validateMaintenanceMode } from "../validators";
import { SettingChangeSummary } from "./setting-change-summary";

function serialize(value: unknown): string {
  return JSON.stringify(value);
}

function pad(value: number): string {
  return String(value).padStart(2, "0");
}

function toDateTimeLocal(value: string | null): string {
  if (!value) {
    return "";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "";
  }

  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(
    date.getHours()
  )}:${pad(date.getMinutes())}`;
}

function fromDateTimeLocal(value: string): string | null {
  if (!value) {
    return null;
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toISOString();
}

function statusLabel(status: MaintenanceModeStatus): string {
  switch (status) {
    case "active":
      return "Active";
    case "scheduled":
      return "Scheduled";
    case "expired":
      return "Expired";
    case "disabled":
      return "Disabled";
  }
}

function statusClass(status: MaintenanceModeStatus): string {
  switch (status) {
    case "active":
      return "border-red-200 bg-red-50 text-red-700";
    case "scheduled":
      return "border-amber-200 bg-amber-50 text-amber-800";
    case "expired":
      return "border-orange-200 bg-orange-50 text-orange-800";
    case "disabled":
      return "border-slate-200 bg-slate-50 text-slate-600";
  }
}

export function MaintenanceModePanel({
  setting,
  canWrite
}: {
  setting?: PlatformSetting<unknown>;
  canWrite: boolean;
}) {
  const currentValue = useMemo(() => readMaintenanceMode(setting), [setting]);
  const [value, setValue] = useState<MaintenanceModeValue>(() => currentValue);
  const [dialogOpen, setDialogOpen] = useState(false);
  const updateSetting = useUpdatePlatformSetting<MaintenanceModeValue>();
  const status = getMaintenanceModeStatus(value);
  const currentStatus = getMaintenanceModeStatus(currentValue);
  const validation = validateMaintenanceMode(value);
  const dirty = serialize(value) !== serialize(currentValue);

  useEffect(() => {
    setValue(currentValue);
  }, [currentValue]);

  async function submitChange({ reason }: { reason: string }) {
    if (!validation.valid) {
      return;
    }

    try {
      await updateSetting.mutateAsync({
        key: "platform.maintenance_mode",
        input: {
          value,
          reason
        }
      });
      setDialogOpen(false);
    } catch {
      // ActionReasonDialog displays mutation.error while preserving local edits.
    }
  }

  return (
    <section className="rounded-lg border border-amber-200 bg-amber-50 p-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-base font-semibold text-slate-950">Maintenance Mode</h2>
            <span
              className={cn(
                "rounded-full border px-2 py-0.5 text-xs font-medium",
                statusClass(status)
              )}
            >
              {statusLabel(status)}
            </span>
            {!settingConfigured(setting) ? (
              <span className="rounded-full border border-amber-300 bg-white px-2 py-0.5 text-xs font-medium text-amber-800">
                Not configured
              </span>
            ) : null}
          </div>
          <p className="mt-1 text-sm text-amber-900">
            Controls platform maintenance banner and access behavior for buyer and seller traffic.
          </p>
          <p className="mt-1 text-xs text-amber-800">
            Last updated {formatDateTime(setting?.updated_at)}
          </p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => setValue(currentValue)}
            disabled={!dirty || updateSetting.isPending}
            className="inline-flex h-9 items-center justify-center gap-2 rounded-lg border border-amber-300 bg-white px-3 text-sm font-medium text-amber-900 hover:bg-amber-100 disabled:cursor-not-allowed disabled:opacity-50"
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
            title="Read-only maintenance mode"
            description="Only superadmins can update maintenance mode."
          />
        </div>
      ) : null}

      <div className="mt-4 grid gap-4 lg:grid-cols-[260px_1fr]">
        <div className="space-y-3">
          <label className="flex min-h-10 items-center gap-3 rounded-lg border border-amber-200 bg-white px-3 text-sm font-medium text-slate-800">
            <input
              type="checkbox"
              checked={value.enabled}
              disabled={!canWrite || updateSetting.isPending}
              onChange={(event) => setValue((current) => ({ ...current, enabled: event.target.checked }))}
              className="h-4 w-4"
            />
            Enable maintenance mode
          </label>

          <label className="flex min-h-10 items-center gap-3 rounded-lg border border-amber-200 bg-white px-3 text-sm font-medium text-slate-800">
            <input
              type="checkbox"
              checked={value.allow_admin_bypass}
              disabled={!canWrite || updateSetting.isPending}
              onChange={(event) =>
                setValue((current) => ({ ...current, allow_admin_bypass: event.target.checked }))
              }
              className="h-4 w-4"
            />
            Allow admin bypass
          </label>
        </div>

        <div className="space-y-3">
          <label className="block">
            <span className="text-sm font-medium text-slate-800">Maintenance message</span>
            <textarea
              value={value.message}
              disabled={!canWrite || updateSetting.isPending}
              rows={3}
              onChange={(event) => setValue((current) => ({ ...current, message: event.target.value }))}
              placeholder="Scheduled maintenance from 01:00 to 02:00 UTC"
              className="mt-1 w-full resize-none rounded-lg border border-amber-200 bg-white p-3 text-sm text-slate-950 outline-none focus:border-amber-700 focus:ring-2 focus:ring-amber-900/10 disabled:bg-amber-100 disabled:text-slate-500"
            />
          </label>

          <div className="grid gap-3 sm:grid-cols-2">
            <label className="block">
              <span className="text-sm font-medium text-slate-800">Starts at</span>
              <input
                type="datetime-local"
                value={toDateTimeLocal(value.starts_at)}
                disabled={!canWrite || updateSetting.isPending}
                onChange={(event) =>
                  setValue((current) => ({
                    ...current,
                    starts_at: fromDateTimeLocal(event.target.value)
                  }))
                }
                className="mt-1 h-10 w-full rounded-lg border border-amber-200 bg-white px-3 text-sm text-slate-950 outline-none focus:border-amber-700 focus:ring-2 focus:ring-amber-900/10 disabled:bg-amber-100 disabled:text-slate-500"
              />
            </label>
            <label className="block">
              <span className="text-sm font-medium text-slate-800">Ends at</span>
              <input
                type="datetime-local"
                value={toDateTimeLocal(value.ends_at)}
                disabled={!canWrite || updateSetting.isPending}
                onChange={(event) =>
                  setValue((current) => ({
                    ...current,
                    ends_at: fromDateTimeLocal(event.target.value)
                  }))
                }
                className="mt-1 h-10 w-full rounded-lg border border-amber-200 bg-white px-3 text-sm text-slate-950 outline-none focus:border-amber-700 focus:ring-2 focus:ring-amber-900/10 disabled:bg-amber-100 disabled:text-slate-500"
              />
            </label>
          </div>
        </div>
      </div>

      {status === "expired" ? (
        <div className="mt-4 flex gap-2 rounded-lg border border-orange-200 bg-white px-3 py-2 text-sm text-orange-800">
          <AlertTriangle className="h-4 w-4 shrink-0" aria-hidden="true" />
          This maintenance window has ended while maintenance remains enabled.
        </div>
      ) : null}

      {!validation.valid ? (
        <p className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {validation.errors.join(" ")}
        </p>
      ) : null}

      <ActionReasonDialog
        open={dialogOpen}
        title={value.enabled ? "Update maintenance mode" : "Disable maintenance mode"}
        description="Confirm the maintenance setting change and provide an audit reason."
        confirmLabel="Confirm update"
        tone={value.enabled ? "danger" : "neutral"}
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
          tone={value.enabled ? "warning" : "neutral"}
          items={[
            { label: "Setting", value: "platform.maintenance_mode" },
            { label: "Current status", value: statusLabel(currentStatus) },
            { label: "Next status", value: statusLabel(status) },
            { label: "Starts", value: formatDateTime(value.starts_at) },
            { label: "Ends", value: formatDateTime(value.ends_at) },
            { label: "Admin bypass", value: value.allow_admin_bypass ? "Allowed" : "Blocked" }
          ]}
        />
      </ActionReasonDialog>
    </section>
  );
}
