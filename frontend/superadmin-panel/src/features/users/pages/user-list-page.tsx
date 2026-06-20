import { useEffect, useMemo, useState } from "react";

import { UserFilterBar } from "../components/user-filter-bar";
import { UserTable } from "../components/user-table";
import { useAdminUsers } from "../hooks/use-admin-users";
import type { UserStatus } from "../types";

const USER_PAGE_LIMIT = 25;

function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedValue(value), delayMs);

    return () => window.clearTimeout(timer);
  }, [delayMs, value]);

  return debouncedValue;
}

export function UserListPage() {
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState<UserStatus | "all">("all");
  const [page, setPage] = useState(1);
  const debouncedQuery = useDebouncedValue(query.trim(), 350);
  const filters = useMemo(
    () => ({
      q: debouncedQuery,
      status,
      page,
      limit: USER_PAGE_LIMIT
    }),
    [debouncedQuery, page, status]
  );
  const usersQuery = useAdminUsers(filters);

  return (
    <section className="flex min-h-[calc(100vh-6.5rem)] flex-col overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
      <header className="border-b border-slate-200 bg-white px-4 py-4">
        <h1 className="text-xl font-semibold text-slate-950">Users</h1>
        <p className="mt-1 text-sm text-slate-600">
          Search users, inspect profile status, and manage account access.
        </p>
      </header>

      <UserFilterBar
        query={query}
        status={status}
        onQueryChange={(nextQuery) => {
          setQuery(nextQuery);
          setPage(1);
        }}
        onStatusChange={(nextStatus) => {
          setStatus(nextStatus);
          setPage(1);
        }}
      />

      <div className="min-h-0 flex-1 p-4">
        <UserTable
          users={usersQuery.data?.users ?? []}
          isLoading={usersQuery.isLoading}
          isFetching={usersQuery.isFetching}
          error={usersQuery.error}
          page={page}
          limit={USER_PAGE_LIMIT}
          totalCount={usersQuery.data?.total}
          onPageChange={setPage}
          onRetry={() => void usersQuery.refetch()}
        />
      </div>
    </section>
  );
}
