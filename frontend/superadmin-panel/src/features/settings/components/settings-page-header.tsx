import { ShieldCheck } from "lucide-react";

export function SettingsPageHeader({
  canWrite,
  roles
}: {
  canWrite: boolean;
  roles: readonly string[];
}) {
  return (
    <header className="border-b border-slate-200 bg-white px-4 py-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-xl font-semibold text-slate-950">Platform Settings</h1>
          <p className="mt-1 text-sm text-slate-600">
            Commission, search synonyms, feature flags, and maintenance controls.
          </p>
        </div>
        <div className="inline-flex h-9 items-center gap-2 self-start rounded-lg border border-slate-200 bg-slate-50 px-3 text-xs font-medium text-slate-700">
          <ShieldCheck className="h-4 w-4" aria-hidden="true" />
          {canWrite ? "Superadmin write access" : "Read-only access"}
          {roles.length > 0 ? <span className="text-slate-400">({roles.join(", ")})</span> : null}
        </div>
      </div>
    </header>
  );
}
