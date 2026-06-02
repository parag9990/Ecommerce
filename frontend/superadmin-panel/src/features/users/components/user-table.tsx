import { Lock, Unlock, UserRound } from "lucide-react";
import { Link } from "react-router-dom";

import { DataState } from "../../../components/ui/data-state";
import { StatusBadge } from "../../../components/ui/status-badge";
import { TablePagination } from "../../../components/ui/table-pagination";
import { maskPhone } from "../../../lib/format";
import type { AdminUser } from "../types";

function userDisplayName(user: AdminUser): string {
  return user.full_name?.trim() || user.email || user.phone || "Unnamed user";
}

function statusIconLabel(user: AdminUser): string {
  if (user.status === "blocked") {
    return "Blocked user";
  }

  if (user.status === "deleted") {
    return "Deleted user";
  }

  return "Active user";
}

export function UserTable({
  users,
  isLoading,
  isFetching,
  error,
  page,
  limit,
  totalCount,
  onPageChange,
  onRetry
}: {
  users: AdminUser[];
  isLoading: boolean;
  isFetching: boolean;
  error: unknown;
  page: number;
  limit: number;
  totalCount?: number;
  onPageChange: (page: number) => void;
  onRetry: () => void;
}) {
  if (isLoading) {
    return <DataState title="Loading users" description="Fetching the latest user records." />;
  }

  if (error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load users"
        description={error instanceof Error ? error.message : "The user list request failed."}
        action={
          <button
            type="button"
            onClick={onRetry}
            className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
          >
            Retry
          </button>
        }
      />
    );
  }

  if (users.length === 0) {
    return <DataState title="No users found" description="Try a different search or status filter." />;
  }

  return (
    <div className="min-h-0 flex-1 overflow-hidden rounded-lg border border-slate-200 bg-white">
      <div className="min-h-0 overflow-auto">
        <table className="min-w-full border-separate border-spacing-0 text-left text-sm">
          <thead className="sticky top-0 bg-slate-100 text-xs uppercase text-slate-600">
            <tr>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">User</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Contact</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Roles</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Status</th>
              <th className="border-b border-slate-200 px-4 py-3 text-right font-semibold">Signal</th>
            </tr>
          </thead>
          <tbody>
            {users.map((user) => {
              const StatusIcon = user.status === "blocked" ? Unlock : Lock;

              return (
                <tr key={user.user_id} className="hover:bg-slate-50">
                  <td className="border-b border-slate-100 px-4 py-3">
                    <div className="flex min-w-0 items-center gap-3">
                      <span className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                        <UserRound size={17} aria-hidden="true" />
                      </span>
                      <div className="min-w-0">
                        <Link
                          to={`/admin/users/${encodeURIComponent(user.user_id)}`}
                          className="block truncate font-medium text-slate-950 hover:text-blue-700"
                        >
                          {userDisplayName(user)}
                        </Link>
                        <div className="truncate text-xs text-slate-500">{user.user_id}</div>
                      </div>
                    </div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                    <div className="max-w-[260px] truncate">{user.email || "No email"}</div>
                    <div className="text-xs text-slate-500">{maskPhone(user.phone)}</div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                    {user.roles.join(", ")}
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3">
                    <StatusBadge status={user.status} />
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3 text-right">
                    <StatusIcon
                      className="ml-auto h-4 w-4 text-slate-500"
                      aria-label={statusIconLabel(user)}
                    />
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <TablePagination
        page={page}
        limit={limit}
        itemCount={users.length}
        totalCount={totalCount}
        isFetching={isFetching}
        onPageChange={onPageChange}
      />
    </div>
  );
}
