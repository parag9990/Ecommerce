import { adminMenu } from "../../../config/admin-menu";
import { filterAdminMenu } from "../../../lib/admin-rbac";
import { useAdminRoles } from "../../../stores/auth-store";

export function AdminHomePage() {
  const roles = useAdminRoles();
  const visibleMenu = filterAdminMenu(roles, adminMenu);

  return (
    <section>
      <div className="mb-4">
        <h1 className="text-xl font-semibold text-slate-950">Overview</h1>
        <p className="mt-1 text-sm text-slate-500">Secure platform control shell is ready.</p>
      </div>

      <div className="grid gap-3 md:grid-cols-3">
        <div className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
          <p className="text-xs font-medium uppercase text-slate-500">Shell</p>
          <p className="mt-2 text-lg font-semibold text-slate-950">Protected</p>
        </div>
        <div className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
          <p className="text-xs font-medium uppercase text-slate-500">Navigation</p>
          <p className="mt-2 text-lg font-semibold text-slate-950">Role-based</p>
        </div>
        <div className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
          <p className="text-xs font-medium uppercase text-slate-500">Visible Modules</p>
          <p className="mt-2 text-lg font-semibold text-slate-950">{visibleMenu.length}</p>
        </div>
      </div>
    </section>
  );
}
