import { useEffect, useMemo, useState } from "react";

import { PermissionDenied } from "../../../components/ui/permission-denied";
import { SellerFilterBar } from "../components/seller-filter-bar";
import { SellerTable } from "../components/seller-table";
import { useAdminSellers } from "../hooks/use-admin-sellers";
import { useSellerPermissions } from "../permissions";
import type { SellerStatus } from "../types";

const SELLER_PAGE_SIZE = 20;

function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedValue(value), delayMs);

    return () => window.clearTimeout(timer);
  }, [delayMs, value]);

  return debouncedValue;
}

export function SellerListPage() {
  const permissions = useSellerPermissions();
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState<SellerStatus | "all">("pending_review");
  const [page, setPage] = useState(1);
  const debouncedQuery = useDebouncedValue(query.trim(), 350);
  const filters = useMemo(
    () => ({
      q: debouncedQuery,
      status,
      page,
      page_size: SELLER_PAGE_SIZE
    }),
    [debouncedQuery, page, status]
  );
  const sellersQuery = useAdminSellers(filters);

  if (!permissions.canViewSellers) {
    return <PermissionDenied compact />;
  }

  return (
    <section className="flex min-h-[calc(100vh-6.5rem)] flex-col overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
      <header className="border-b border-slate-200 bg-white px-4 py-4">
        <h1 className="text-xl font-semibold text-slate-950">Sellers</h1>
        <p className="mt-1 text-sm text-slate-600">
          Review KYC applications, seller status, suspension controls, and catalog moderation signals.
        </p>
      </header>

      <SellerFilterBar
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
        <SellerTable
          sellers={sellersQuery.data?.sellers ?? []}
          canViewKyc={permissions.canReviewSellerKyc}
          isLoading={sellersQuery.isLoading}
          isFetching={sellersQuery.isFetching}
          error={sellersQuery.error}
          page={page}
          pageSize={SELLER_PAGE_SIZE}
          totalCount={sellersQuery.data?.total}
          onPageChange={setPage}
          onRetry={() => void sellersQuery.refetch()}
        />
      </div>
    </section>
  );
}
