import { ChevronLeft, ChevronRight } from "lucide-react";
import { useId, useState } from "react";

import { getSafeErrorMessage } from "../../../lib/api-error";
import type {
  AssignableSellerStaffRole,
  SellerStaffMember,
  SellerTeamFilters,
  SellerTeamPagination,
} from "../types";
import { useSellerPermissions } from "../hooks/use-seller-permissions";
import { useUpdateSellerStaffRole } from "../hooks/use-seller-team";
import {
  isAssignableSellerStaffRole,
  ROLE_LABELS,
} from "../utils/seller-permissions";
import { formatDateTime, formatMemberName } from "../utils/team-formatters";
import { RoleSelect } from "./role-select";
import { StaffActionsMenu } from "./staff-actions-menu";
import { StaffStatusBadge } from "./staff-status-badge";

type TeamMembersTableProps = {
  members: SellerStaffMember[];
  pagination: SellerTeamPagination;
  filters: SellerTeamFilters;
  onFiltersChange: (filters: SellerTeamFilters) => void;
};

function TeamMemberRoleCell({ member }: { member: SellerStaffMember }) {
  const labelId = useId();
  const permissions = useSellerPermissions();
  const updateRoleMutation = useUpdateSellerStaffRole();
  const [error, setError] = useState("");

  const canUpdateRole =
    permissions.can("team:update_role") &&
    isAssignableSellerStaffRole(member.role) &&
    member.status !== "disabled";

  async function updateRole(role: AssignableSellerStaffRole) {
    if (role === member.role) {
      return;
    }

    try {
      setError("");
      await updateRoleMutation.mutateAsync({
        staff_id: member.staff_id,
        role,
      });
    } catch (mutationError) {
      setError(getSafeErrorMessage(mutationError));
    }
  }

  if (!isAssignableSellerStaffRole(member.role)) {
    return (
      <div>
        <div className="font-medium text-slate-800">{ROLE_LABELS[member.role]}</div>
        <p className="mt-0.5 text-xs text-slate-500">Protected owner role</p>
      </div>
    );
  }

  return (
    <div className="max-w-xs space-y-1">
      <span id={labelId} className="sr-only">
        Role for {member.email}
      </span>
      <RoleSelect
        labelledBy={labelId}
        value={member.role}
        disabled={!canUpdateRole || updateRoleMutation.isPending}
        onChange={updateRole}
      />
      {error ? <p className="text-xs text-rose-600">{error}</p> : null}
      {!canUpdateRole ? (
        <p className="text-xs text-slate-500">
          {member.status === "disabled" ? "Disabled staff cannot be edited." : "Role locked."}
        </p>
      ) : null}
    </div>
  );
}

export function TeamMembersTable({
  members,
  pagination,
  filters,
  onFiltersChange,
}: TeamMembersTableProps) {
  const hasPreviousPage = filters.page > 1;
  const hasNextPage = filters.page * filters.page_size < pagination.total;

  return (
    <div className="overflow-hidden rounded-md border border-slate-200 bg-white shadow-sm">
      <div className="overflow-x-auto">
        <table className="w-full min-w-[1040px] border-collapse text-left text-sm">
          <thead className="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-4 py-3 font-semibold">Member</th>
              <th className="px-4 py-3 font-semibold">Role</th>
              <th className="px-4 py-3 font-semibold">Status</th>
              <th className="px-4 py-3 font-semibold">Invited by</th>
              <th className="px-4 py-3 font-semibold">Added</th>
              <th className="px-4 py-3 font-semibold">Updated</th>
              <th className="px-4 py-3 text-right font-semibold">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {members.map((member) => (
              <tr key={member.staff_id} className="align-middle">
                <td className="px-4 py-3">
                  <div className="font-medium text-slate-950">
                    {formatMemberName(member.full_name, member.email)}
                  </div>
                  <div className="mt-0.5 text-xs text-slate-500">{member.email}</div>
                  <div className="mt-0.5 text-xs text-slate-400">{member.staff_id}</div>
                </td>
                <td className="px-4 py-3">
                  <TeamMemberRoleCell member={member} />
                </td>
                <td className="px-4 py-3">
                  <StaffStatusBadge status={member.status} />
                </td>
                <td className="px-4 py-3 text-slate-600">
                  {member.invited_by ?? "System"}
                </td>
                <td className="px-4 py-3 text-slate-600">
                  {formatDateTime(member.created_at)}
                </td>
                <td className="px-4 py-3 text-slate-600">
                  {formatDateTime(member.updated_at)}
                </td>
                <td className="px-4 py-3 text-right">
                  <StaffActionsMenu member={member} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 px-4 py-3 text-sm">
        <span className="text-slate-500">
          Page {filters.page} - {pagination.total} total
        </span>
        <div className="flex items-center gap-2">
          <button
            type="button"
            aria-label="Previous page"
            title="Previous page"
            disabled={!hasPreviousPage}
            onClick={() => onFiltersChange({ ...filters, page: filters.page - 1 })}
            className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <ChevronLeft className="h-4 w-4" aria-hidden="true" />
          </button>
          <button
            type="button"
            aria-label="Next page"
            title="Next page"
            disabled={!hasNextPage}
            onClick={() => onFiltersChange({ ...filters, page: filters.page + 1 })}
            className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <ChevronRight className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
      </div>
    </div>
  );
}
