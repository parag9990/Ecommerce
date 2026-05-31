import { HelpCircle, Monitor, Smartphone, Tablet } from "lucide-react";

import type { ActiveSessionBreakdownItem } from "../../../api/session-api";
import { normalizeBreakdownWidth } from "../../../lib/session-view";

type DeviceBreakdownPanelProps = {
  devices: ActiveSessionBreakdownItem[];
};

const deviceIcons = {
  desktop: Monitor,
  mobile: Smartphone,
  tablet: Tablet,
  unknown: HelpCircle
};

export function DeviceBreakdownPanel({ devices }: DeviceBreakdownPanelProps) {
  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <h2 className="mb-4 text-sm font-semibold text-zinc-950">Device mix</h2>

      {devices.length === 0 ? (
        <p className="text-sm text-zinc-500">No device data yet.</p>
      ) : (
        <div className="space-y-3">
          {devices.map((device) => {
            const Icon =
              deviceIcons[device.label as keyof typeof deviceIcons] ?? HelpCircle;

            return (
              <div
                className="rounded-md border border-zinc-200 bg-zinc-50 p-3"
                key={device.label}
              >
                <div className="mb-2 flex items-center justify-between gap-3">
                  <div className="flex min-w-0 items-center gap-2">
                    <Icon
                      className="h-4 w-4 shrink-0 text-emerald-600"
                      aria-hidden="true"
                    />
                    <span className="truncate text-sm capitalize text-zinc-700">
                      {device.label}
                    </span>
                  </div>
                  <span className="shrink-0 text-xs text-zinc-500">
                    {device.count} / {device.percentage}%
                  </span>
                </div>
                <div className="h-1.5 rounded-full bg-white">
                  <div
                    className="h-1.5 rounded-full bg-zinc-900"
                    style={{ width: normalizeBreakdownWidth(device.percentage) }}
                  />
                </div>
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
}
