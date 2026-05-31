import {
  CheckCircle2,
  Circle,
  Clock3,
  MapPin,
  MonitorSmartphone,
  MousePointerClick,
  ShoppingCart
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import type { JourneySession, JourneySummary } from "../../../api/session-api";
import {
  formatDateTime,
  formatDuration,
  formatNumber
} from "../../../lib/format";
import {
  formatDeviceLabel,
  formatLocation,
  maskAnonymousId,
  maskSessionId,
  maskUserId
} from "../../../lib/session-view";
import { sanitizeDisplayText } from "../lib/event-privacy";

type SessionContextCardProps = {
  session: JourneySession;
  summary: JourneySummary;
};

export function SessionContextCard({
  session,
  summary
}: SessionContextCardProps) {
  const actor = session.userId
    ? maskUserId(session.userId)
    : maskAnonymousId(session.anonymousId ?? "");
  const exitPage = session.exitPage ?? (session.status === "active" ? "Active" : "-");

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
        <div className="min-w-0">
          <p className="font-mono text-xs font-medium text-emerald-700">
            {maskSessionId(session.sessionId)}
          </p>
          <h2 className="mt-1 truncate text-base font-semibold text-zinc-950">
            {sanitizeDisplayText(session.entryPage)} -&gt;{" "}
            {sanitizeDisplayText(exitPage)}
          </h2>
          <p className="mt-1 text-sm text-zinc-500">User {actor}</p>
        </div>

        <div className="flex flex-wrap gap-2">
          <StatusBadge status={session.status} />
          <BooleanBadge
            active={summary.checkoutStarted}
            label="Checkout"
          />
          <BooleanBadge active={summary.paymentCompleted} label="Payment" />
        </div>
      </div>

      <div className="mt-4 grid gap-3 border-t border-zinc-100 pt-4 sm:grid-cols-2 xl:grid-cols-5">
        <Metric
          icon={Clock3}
          label="Duration"
          value={formatDuration(session.durationSeconds)}
        />
        <Metric
          icon={MonitorSmartphone}
          label="Device"
          value={`${session.device.type} / ${formatDeviceLabel(session.device)}`}
        />
        <Metric
          icon={MousePointerClick}
          label="Events"
          value={formatNumber(summary.totalEvents)}
        />
        <Metric
          icon={ShoppingCart}
          label="Cart"
          value={formatNumber(summary.cartActions)}
        />
        <Metric
          icon={MapPin}
          label="Location"
          value={formatLocation(session.geo ?? {})}
        />
      </div>

      <div className="mt-4 flex flex-wrap gap-x-4 gap-y-1 border-t border-zinc-100 pt-3 text-xs text-zinc-500">
        <span>Started {formatDateTime(session.startedAt)}</span>
        <span>Last seen {formatDateTime(session.lastSeenAt)}</span>
      </div>
    </section>
  );
}

type MetricProps = {
  icon: LucideIcon;
  label: string;
  value: string;
};

function Metric({ icon: Icon, label, value }: MetricProps) {
  return (
    <div className="min-w-0">
      <div className="flex items-center gap-2 text-xs font-medium uppercase text-zinc-500">
        <Icon className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
        {label}
      </div>
      <div className="mt-1 truncate text-sm font-semibold text-zinc-950">
        {value}
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: JourneySession["status"] }) {
  const className =
    status === "active"
      ? "border-emerald-200 bg-emerald-50 text-emerald-700"
      : "border-zinc-200 bg-zinc-50 text-zinc-700";

  return (
    <span
      className={[
        "inline-flex h-8 items-center rounded-md border px-3 text-xs font-medium capitalize",
        className
      ].join(" ")}
    >
      {status}
    </span>
  );
}

function BooleanBadge({ active, label }: { active: boolean; label: string }) {
  const Icon = active ? CheckCircle2 : Circle;

  return (
    <span
      className={[
        "inline-flex h-8 items-center gap-1.5 rounded-md border px-3 text-xs font-medium",
        active
          ? "border-emerald-200 bg-emerald-50 text-emerald-700"
          : "border-zinc-200 bg-zinc-50 text-zinc-500"
      ].join(" ")}
    >
      <Icon className="h-3.5 w-3.5" aria-hidden="true" />
      {label}
    </span>
  );
}
