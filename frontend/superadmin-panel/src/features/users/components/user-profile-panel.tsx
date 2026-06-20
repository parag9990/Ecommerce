import { StatusBadge } from "../../../components/ui/status-badge";
import { useCanMutateUserStatus } from "../../../lib/admin-permissions";
import { formatDateTime, maskPhone } from "../../../lib/format";
import type { AdminUser } from "../types";
import { UserActionDialog } from "./user-action-dialog";

function fieldValue(value?: string | null): string {
  return value?.trim() || "Not added";
}

export function UserProfilePanel({ user }: { user: AdminUser }) {
  const canMutateStatus = useCanMutateUserStatus();
  const nextStatus = user.status === "blocked" ? "active" : "blocked";
  const canShowAction = canMutateStatus && user.status !== "deleted";

  return (
    <aside className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-base font-semibold text-slate-950">Profile</h2>
          <p className="mt-1 truncate text-sm text-slate-600">{fieldValue(user.email)}</p>
        </div>
        <StatusBadge status={user.status} />
      </div>

      <dl className="mt-4 space-y-3 text-sm">
        <div>
          <dt className="text-xs font-semibold uppercase text-slate-500">User ID</dt>
          <dd className="break-all text-slate-950">{user.user_id}</dd>
        </div>
        <div>
          <dt className="text-xs font-semibold uppercase text-slate-500">Phone</dt>
          <dd className="text-slate-950">{maskPhone(user.phone)}</dd>
        </div>
        <div>
          <dt className="text-xs font-semibold uppercase text-slate-500">Roles</dt>
          <dd className="text-slate-950">{user.roles.join(", ")}</dd>
        </div>
        <div>
          <dt className="text-xs font-semibold uppercase text-slate-500">Created</dt>
          <dd className="text-slate-950">{formatDateTime(user.created_at)}</dd>
        </div>
        <div>
          <dt className="text-xs font-semibold uppercase text-slate-500">Last login</dt>
          <dd className="text-slate-950">{formatDateTime(user.last_login_at)}</dd>
        </div>
      </dl>

      {canShowAction ? (
        <div className="mt-5 border-t border-slate-200 pt-4">
          <UserActionDialog user={user} nextStatus={nextStatus} />
        </div>
      ) : null}
    </aside>
  );
}
