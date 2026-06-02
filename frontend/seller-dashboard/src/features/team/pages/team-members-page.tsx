import { ShieldCheck, UserPlus, Users } from "lucide-react";
import { useMemo, useState } from "react";

import { CardsSkeleton, TableSkeleton } from "../../../components/state/loading-skeleton";
import { PermissionDeniedState } from "../../../components/state/permission-denied-state";
import { RefreshingNotice } from "../../../components/state/refreshing-notice";
import { getSafeErrorMessage } from "../../../lib/api-error";
import { useSellerStore } from "../../../stores/seller-store";
import { InviteStaffDialog } from "../components/invite-staff-dialog";
import { RolePermissionMatrix } from "../components/role-permission-matrix";
import { TeamEmptyState } from "../components/team-empty-state";
import { TeamErrorState } from "../components/team-error-state";
import { TeamMembersTable } from "../components/team-members-table";
import { useSellerPermissions } from "../hooks/use-seller-permissions";
import { useSellerTeam } from "../hooks/use-seller-team";
import type { SellerTeamFilters, SellerTeamStatusFilter } from "../types";

const defaultFilters: SellerTeamFilters = {
  page: 1,
  page_size: 20,
  status: "all",
};

const statusOptions: { label: string; value: SellerTeamStatusFilter }[] = [
  { label: "All statuses", value: "all" },
  { label: "Invited", value: "invited" },
  { label: "Active", value: "active" },
  { label: "Disabled", value: "disabled" },
];

export function TeamMembersPage() {
  const activeSeller = useSellerStore((state) => state.activeSeller);
  const permissions = useSellerPermissions();
  const [inviteOpen, setInviteOpen] = useState(false);
  const [filters, setFilters] = useState(defaultFilters);

  const canViewTeam = permissions.can("team:view");
  const canInvite = permissions.can("team:invite");
  const teamQuery = useSellerTeam(
    canViewTeam ? activeSeller?.seller_id : undefined,
    filters,
  );
  const members = teamQuery.data?.members ?? [];
  const statusCounts = useMemo(
    () => ({
      invited: members.filter((member) => member.status === "invited").length,
      active: members.filter((member) => member.status === "active").length,
      disabled: members.filter((member) => member.status === "disabled").length,
    }),
    [members],
  );

  function updateStatus(status: SellerTeamStatusFilter) {
    setFilters((current) => ({
      ...current,
      page: 1,
      status,
    }));
  }

  function updatePageSize(pageSize: number) {
    setFilters((current) => ({
      ...current,
      page: 1,
      page_size: pageSize,
    }));
  }

  if (!activeSeller) {
    return (
      <PermissionDeniedState
        title="Active seller unavailable"
        description="Team permissions load karne ke liye active seller context required hai."
      />
    );
  }

  if (!canViewTeam) {
    return (
      <PermissionDeniedState
        title="Team access unavailable"
        description="Aapke current seller role ke paas team management access nahi hai."
      />
    );
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <Users className="h-3.5 w-3.5" aria-hidden="true" />
            Team Management
          </p>
          <h1 className="mt-1 text-xl font-semibold text-slate-950">
            Team permissions
          </h1>
          <p className="mt-1 max-w-2xl text-sm text-slate-500">
            Invite staff, assign scoped roles, and keep dashboard access aligned
            with seller operations.
          </p>
        </div>

        <button
          type="button"
          disabled={!canInvite}
          title={!canInvite ? "You do not have permission to invite staff." : undefined}
          onClick={() => setInviteOpen(true)}
          className="inline-flex h-10 items-center gap-2 rounded-md bg-slate-950 px-3 text-sm font-medium text-white transition hover:bg-slate-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-950 disabled:cursor-not-allowed disabled:opacity-60"
        >
          <UserPlus className="h-4 w-4" aria-hidden="true" />
          Invite staff
        </button>
      </div>

      {teamQuery.isPending ? (
        <CardsSkeleton count={3} className="md:grid-cols-3 xl:grid-cols-3" />
      ) : teamQuery.data ? (
        <div className="grid gap-3 md:grid-cols-3">
          <div className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Invited
            </p>
            <p className="mt-2 text-2xl font-semibold text-slate-950">
              {statusCounts.invited}
            </p>
          </div>
          <div className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Active
            </p>
            <p className="mt-2 text-2xl font-semibold text-slate-950">
              {statusCounts.active}
            </p>
          </div>
          <div className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Disabled
            </p>
            <p className="mt-2 text-2xl font-semibold text-slate-950">
              {statusCounts.disabled}
            </p>
          </div>
        </div>
      ) : null}

      <div className="space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="inline-flex items-center gap-2 text-sm font-semibold text-slate-950">
            <ShieldCheck className="h-4 w-4 text-blue-600" aria-hidden="true" />
            Role access matrix
          </div>
        </div>
        <RolePermissionMatrix />
      </div>

      <div className="space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-slate-200 bg-white p-3 shadow-sm">
          <div>
            <h2 className="text-sm font-semibold text-slate-950">Staff members</h2>
            <p className="mt-0.5 text-sm text-slate-500">
              Backend remains the source of truth for every access decision.
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <label className="flex items-center gap-2 text-sm text-slate-600">
              Status
              <select
                value={filters.status}
                onChange={(event) =>
                  updateStatus(event.target.value as SellerTeamStatusFilter)
                }
                className="h-9 rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-900 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              >
                {statusOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex items-center gap-2 text-sm text-slate-600">
              Rows
              <select
                value={filters.page_size}
                onChange={(event) => updatePageSize(Number(event.target.value))}
                className="h-9 rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-900 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              >
                {[10, 20, 50].map((pageSize) => (
                  <option key={pageSize} value={pageSize}>
                    {pageSize}
                  </option>
                ))}
              </select>
            </label>
          </div>
        </div>

        <RefreshingNotice show={teamQuery.isFetching && !teamQuery.isPending} />

        {teamQuery.isPending ? <TableSkeleton columns={7} /> : null}

        {teamQuery.isError && members.length > 0 ? (
          <RefreshingNotice
            show
            failed
            message={getSafeErrorMessage(teamQuery.error)}
            onRetry={() => teamQuery.refetch()}
          />
        ) : null}

        {teamQuery.isError && members.length === 0 ? (
          <TeamErrorState error={teamQuery.error} onRetry={() => teamQuery.refetch()} />
        ) : null}

        {teamQuery.data && members.length === 0 ? (
          <TeamEmptyState canInvite={canInvite} onInvite={() => setInviteOpen(true)} />
        ) : null}

        {teamQuery.data && members.length > 0 ? (
          <TeamMembersTable
            members={members}
            pagination={teamQuery.data.pagination}
            filters={filters}
            onFiltersChange={setFilters}
          />
        ) : null}
      </div>

      <InviteStaffDialog open={inviteOpen} onOpenChange={setInviteOpen} />
    </section>
  );
}
