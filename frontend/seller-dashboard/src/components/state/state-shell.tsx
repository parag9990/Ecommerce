import { Link } from "react-router-dom";

import { cn } from "../../lib/cn";
import type { DashboardStateKind, StateAction, StateCopy } from "./state-types";

const stateClassNames: Record<DashboardStateKind, string> = {
  loading: "border-slate-200 bg-slate-50 text-slate-700",
  empty: "border-blue-200 bg-blue-50 text-blue-950",
  failed: "border-red-200 bg-red-50 text-red-950",
  permissionDenied: "border-amber-200 bg-amber-50 text-amber-950",
  unavailable: "border-violet-200 bg-violet-50 text-violet-950",
};

type StateShellProps = StateCopy & {
  kind: DashboardStateKind;
  className?: string;
};

export function StateShell({
  kind,
  title,
  description,
  icon,
  action,
  secondaryAction,
  requestId,
  className,
}: StateShellProps) {
  return (
    <section
      className={cn(
        "rounded-md border p-4 shadow-sm sm:p-5",
        "flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between",
        stateClassNames[kind],
        className,
      )}
      role={kind === "failed" ? "alert" : "status"}
      aria-live={kind === "failed" ? "assertive" : "polite"}
    >
      <div className="flex min-w-0 gap-3">
        {icon ? <div className="mt-0.5 shrink-0">{icon}</div> : null}
        <div className="min-w-0">
          <h2 className="text-sm font-semibold">{title}</h2>
          <p className="mt-1 max-w-2xl text-sm leading-6 opacity-80">
            {description}
          </p>
          {requestId ? (
            <p className="mt-2 break-all font-mono text-xs opacity-70">
              Request ID: {requestId}
            </p>
          ) : null}
        </div>
      </div>

      {secondaryAction || action ? (
        <div className="flex shrink-0 flex-wrap gap-2 sm:justify-end">
          {secondaryAction ? <StateActionButton action={secondaryAction} /> : null}
          {action ? <StateActionButton action={action} /> : null}
        </div>
      ) : null}
    </section>
  );
}

function getActionClassName(action: StateAction) {
  return action.variant === "secondary"
    ? "inline-flex h-9 items-center justify-center rounded-md border border-current bg-white/70 px-3 text-sm font-medium transition hover:bg-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-current disabled:cursor-not-allowed disabled:opacity-60"
    : "inline-flex h-9 items-center justify-center rounded-md bg-slate-950 px-3 text-sm font-medium text-white transition hover:bg-slate-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-950 disabled:cursor-not-allowed disabled:opacity-60";
}

function StateActionButton({ action }: { action: StateAction }) {
  const className = getActionClassName(action);

  if (action.href && !action.disabled) {
    return (
      <Link to={action.href} className={className} title={action.title}>
        {action.label}
      </Link>
    );
  }

  return (
    <button
      type="button"
      className={className}
      disabled={action.disabled}
      onClick={action.onClick}
      title={action.title}
    >
      {action.label}
    </button>
  );
}
