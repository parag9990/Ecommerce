import { ClipboardCheck, Store } from "lucide-react";
import { Link } from "react-router-dom";

import { DataState } from "../../../components/ui/data-state";
import { StatusBadge } from "../../../components/ui/status-badge";
import { TablePagination } from "../../../components/ui/table-pagination";
import { formatDateTime } from "../../../lib/format";
import type { AdminSeller } from "../types";

function sellerDisplayName(seller: AdminSeller): string {
  return seller.store_name?.trim() || seller.display_name?.trim() || "Unnamed store";
}

function kycSummary(seller: AdminSeller, canViewKyc: boolean): string {
  if (!canViewKyc) {
    return "Masked";
  }

  const pendingCount = seller.pending_document_count ?? 0;
  const totalCount = seller.document_count ?? 0;

  if (totalCount > 0) {
    return `${pendingCount} pending / ${totalCount} total`;
  }

  return seller.kyc_status?.replace(/_/g, " ") ?? "No documents";
}

export function SellerTable({
  sellers,
  canViewKyc,
  isLoading,
  isFetching,
  error,
  page,
  pageSize,
  totalCount,
  onPageChange,
  onRetry
}: {
  sellers: AdminSeller[];
  canViewKyc: boolean;
  isLoading: boolean;
  isFetching: boolean;
  error: unknown;
  page: number;
  pageSize: number;
  totalCount?: number;
  onPageChange: (page: number) => void;
  onRetry: () => void;
}) {
  if (isLoading) {
    return <DataState title="Loading sellers" description="Fetching seller applications and status signals." />;
  }

  if (error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load sellers"
        description={error instanceof Error ? error.message : "The seller list request failed."}
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

  if (sellers.length === 0) {
    return <DataState title="No sellers found" description="Try a different search or seller status filter." />;
  }

  return (
    <div className="min-h-0 flex-1 overflow-hidden rounded-lg border border-slate-200 bg-white">
      <div className="min-h-0 overflow-auto">
        <table className="min-w-full border-separate border-spacing-0 text-left text-sm">
          <thead className="sticky top-0 bg-slate-100 text-xs uppercase text-slate-600">
            <tr>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Store</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Status</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">KYC</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Catalog</th>
              <th className="border-b border-slate-200 px-4 py-3 font-semibold">Updated</th>
              <th className="border-b border-slate-200 px-4 py-3 text-right font-semibold">Review</th>
            </tr>
          </thead>
          <tbody>
            {sellers.map((seller) => (
              <tr key={seller.seller_id} className="hover:bg-slate-50">
                <td className="border-b border-slate-100 px-4 py-3">
                  <div className="flex min-w-0 items-center gap-3">
                    <span className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                      <Store size={17} aria-hidden="true" />
                    </span>
                    <div className="min-w-0">
                      <Link
                        to={`/admin/sellers/${encodeURIComponent(seller.seller_id)}`}
                        className="block truncate font-medium text-slate-950 hover:text-blue-700"
                      >
                        {sellerDisplayName(seller)}
                      </Link>
                      <div className="truncate text-xs text-slate-500">{seller.seller_id}</div>
                    </div>
                  </div>
                </td>
                <td className="border-b border-slate-100 px-4 py-3">
                  <StatusBadge status={seller.status} />
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                  <div className="max-w-[190px] truncate">{kycSummary(seller, canViewKyc)}</div>
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                  <div className="flex items-center gap-2">
                    <ClipboardCheck className="h-4 w-4 text-slate-500" aria-hidden="true" />
                    <span>{seller.pending_catalog_count ?? 0} pending</span>
                  </div>
                  <div className="text-xs text-slate-500">{seller.product_count ?? 0} products</div>
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                  {formatDateTime(seller.updated_at)}
                </td>
                <td className="border-b border-slate-100 px-4 py-3 text-right">
                  <Link
                    to={`/admin/sellers/${encodeURIComponent(seller.seller_id)}`}
                    className="font-medium text-blue-700 hover:text-blue-900"
                  >
                    Review
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <TablePagination
        page={page}
        limit={pageSize}
        itemCount={sellers.length}
        totalCount={totalCount}
        isFetching={isFetching}
        onPageChange={onPageChange}
      />
    </div>
  );
}
