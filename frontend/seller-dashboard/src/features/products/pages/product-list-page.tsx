import { Plus } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";

import { AsyncStateBoundary } from "../../../components/state/async-state-boundary";
import { EmptyState } from "../../../components/state/empty-state";
import { PermissionDeniedState } from "../../../components/state/permission-denied-state";
import { RefreshingNotice } from "../../../components/state/refreshing-notice";
import { StateShell } from "../../../components/state/state-shell";
import { TableSkeleton } from "../../../components/state/loading-skeleton";
import { useSellerStore } from "../../../stores/seller-store";
import { useSellerPermissions } from "../../team/hooks/use-seller-permissions";
import { ProductFiltersBar } from "../components/product-filters-bar";
import { ProductTable } from "../components/product-table";
import { useCategories } from "../hooks/use-categories";
import { useSellerProducts } from "../hooks/use-seller-products";
import type { ProductStatusFilter } from "../types";

export function ProductListPage() {
  const activeSeller = useSellerStore((state) => state.activeSeller);
  const permissions = useSellerPermissions();
  const canViewProducts = permissions.can("products:view");
  const canWriteProducts = permissions.can("products:write");
  const [status, setStatus] = useState<ProductStatusFilter>("all");
  const [categoryId, setCategoryId] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);

  const categoriesQuery = useCategories(canViewProducts);
  const productsQuery = useSellerProducts({
    seller_id: canViewProducts ? activeSeller?.seller_id ?? "" : "",
    category_id: categoryId || undefined,
    status,
    page,
    page_size: pageSize,
  });

  function resetPage(next: () => void) {
    setPage(1);
    next();
  }

  if (!activeSeller) {
    return (
      <PermissionDeniedState
        title="Active seller unavailable"
        description="Products load karne ke liye active seller context required hai."
      />
    );
  }

  if (!canViewProducts) {
    return (
      <PermissionDeniedState
        title="Products access unavailable"
        description="Aapke current seller role ke paas product catalog dekhne ka permission nahi hai."
      />
    );
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Catalog
          </p>
          <h1 className="mt-1 text-xl font-semibold text-slate-950">Products</h1>
        </div>

        {canWriteProducts ? (
          <Link
            to="/seller/products/new"
            className="inline-flex h-10 items-center gap-2 rounded-md bg-slate-950 px-3 text-sm font-medium text-white transition hover:bg-slate-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-950"
          >
            <Plus className="h-4 w-4" aria-hidden="true" />
            Add product
          </Link>
        ) : null}
      </div>

      <ProductFiltersBar
        status={status}
        categoryId={categoryId}
        pageSize={pageSize}
        categories={categoriesQuery.data ?? []}
        onStatusChange={(nextStatus) => resetPage(() => setStatus(nextStatus))}
        onCategoryChange={(nextCategoryId) => resetPage(() => setCategoryId(nextCategoryId))}
        onPageSizeChange={(nextPageSize) => resetPage(() => setPageSize(nextPageSize))}
      />

      {categoriesQuery.isError ? (
        <StateShell
          kind="unavailable"
          title="Category names unavailable"
          description="Products still load honge, lekin category labels fallback IDs se dikh sakte hain."
        />
      ) : null}

      <RefreshingNotice
        show={productsQuery.isFetching && !productsQuery.isPending}
      />

      <AsyncStateBoundary
        isLoading={productsQuery.isPending}
        isError={productsQuery.isError}
        error={productsQuery.error}
        data={productsQuery.data}
        isEmpty={(data) => data.products.length === 0}
        loadingFallback={<TableSkeleton columns={6} />}
        emptyFallback={
          <EmptyState
            title={status === "all" && !categoryId ? "No products yet" : "No products found"}
            description={
              status === "all" && !categoryId
                ? "Pehla product draft create karke catalog start karein."
                : "Selected filters ke liye products nahi mile. Filters clear karke retry karein."
            }
            action={
              status === "all" && !categoryId && canWriteProducts
                ? { label: "Create product", href: "/seller/products/new" }
                : {
                    label: "Clear filters",
                    onClick: () => {
                      setStatus("all");
                      setCategoryId("");
                      setPage(1);
                    },
                    variant: "secondary",
                  }
            }
          />
        }
        onRetry={() => productsQuery.refetch()}
      >
        {(data) => (
          <ProductTable
            products={data.products}
            total={data.total}
            page={page}
            pageSize={pageSize}
            categories={categoriesQuery.data ?? []}
            onPageChange={setPage}
          />
        )}
      </AsyncStateBoundary>
    </section>
  );
}
