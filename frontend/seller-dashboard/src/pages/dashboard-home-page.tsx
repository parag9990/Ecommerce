import { StatusBadge } from "../components/ui/status-badge";
import { PermissionDeniedState } from "../components/state/permission-denied-state";
import { UnavailableState } from "../components/state/unavailable-state";
import { useSellerStore } from "../stores/seller-store";
import { useSellerPermissions } from "../features/team/hooks/use-seller-permissions";

export function DashboardHomePage() {
  const activeSeller = useSellerStore((state) => state.activeSeller);
  const permissions = useSellerPermissions();

  if (!permissions.can("dashboard:view")) {
    return (
      <PermissionDeniedState
        title="Dashboard access unavailable"
        description="Aapke current seller role ke paas dashboard overview dekhne ka permission nahi hai."
      />
    );
  }

  return (
    <section className="space-y-4">
      <div className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Active seller
            </p>
            <h2 className="mt-1 truncate text-lg font-semibold text-slate-950">
              {activeSeller?.display_name ?? "Seller not selected"}
            </h2>
          </div>

          {activeSeller ? <StatusBadge status={activeSeller.status} /> : null}
        </div>

        <p className="mt-3 max-w-3xl text-sm leading-6 text-slate-600">
          Product, order, offers, analytics, team, and audit workflows are
          available from the navigation according to your seller role.
        </p>
      </div>

      {!activeSeller ? (
        <PermissionDeniedState
          title="Active seller unavailable"
          description="Dashboard overview ke liye active seller context required hai."
        />
      ) : null}

      <UnavailableState
        title="Overview metrics unavailable"
        description="Dashboard summary cards ke liye dedicated overview data source current frontend contract me available nahi hai. Real module data navigation se accessible rahega."
      />

      <div className="grid gap-3 md:grid-cols-3">
        <div className="rounded-md border border-slate-200 bg-white p-4">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Shell
          </p>
          <p className="mt-1 text-sm font-medium text-slate-950">
            Protected layout
          </p>
        </div>
        <div className="rounded-md border border-slate-200 bg-white p-4">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Navigation
          </p>
          <p className="mt-1 text-sm font-medium text-slate-950">
            Active and reserved modules
          </p>
        </div>
        <div className="rounded-md border border-slate-200 bg-white p-4">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Context
          </p>
          <p className="mt-1 text-sm font-medium text-slate-950">
            Seller switcher
          </p>
        </div>
      </div>
    </section>
  );
}
