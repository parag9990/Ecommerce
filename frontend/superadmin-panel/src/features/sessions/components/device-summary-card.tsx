import { Globe2, MonitorSmartphone, Shield } from "lucide-react";

import { MaskedIdentity } from "./masked-identity";
import type { AdminSession } from "../types";

function readable(value?: string | null): string {
  return value?.trim() || "Unknown";
}

export function DeviceSummaryCard({ session }: { session: AdminSession }) {
  const device = session.device ?? {};
  const operatingSystem = device.operating_system ?? device.os;
  const location = [device.city, device.country].filter(Boolean).join(", ") || "Unknown";

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4">
      <h3 className="text-base font-semibold text-slate-950">Device Summary</h3>
      <dl className="mt-4 grid gap-3 text-sm">
        <div className="flex items-start gap-3">
          <MonitorSmartphone className="mt-0.5 h-4 w-4 shrink-0 text-slate-500" aria-hidden="true" />
          <div>
            <dt className="font-medium text-slate-950">{readable(device.device_type)}</dt>
            <dd className="text-slate-600">
              {readable(device.browser)} on {readable(operatingSystem)}
            </dd>
          </div>
        </div>
        <div className="flex items-start gap-3">
          <Globe2 className="mt-0.5 h-4 w-4 shrink-0 text-slate-500" aria-hidden="true" />
          <div>
            <dt className="font-medium text-slate-950">{location}</dt>
            <dd className="text-slate-600">Language {readable(device.language)}</dd>
          </div>
        </div>
        <div className="flex items-start gap-3">
          <Shield className="mt-0.5 h-4 w-4 shrink-0 text-slate-500" aria-hidden="true" />
          <div className="min-w-0">
            <dt className="font-medium text-slate-950">Network and fingerprint</dt>
            <dd className="mt-1 flex min-w-0 flex-wrap gap-2 text-slate-600">
              <MaskedIdentity value={device.ip_hash} emptyLabel="No IP hash" label="IP hash" />
              <MaskedIdentity
                value={device.device_fingerprint_hash}
                emptyLabel="No fingerprint hash"
                label="Device fingerprint"
              />
            </dd>
          </div>
        </div>
      </dl>
    </section>
  );
}
