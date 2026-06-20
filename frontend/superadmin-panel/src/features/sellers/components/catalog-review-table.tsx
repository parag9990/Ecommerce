import { ImageOff } from "lucide-react";

import { DataState } from "../../../components/ui/data-state";
import { StatusBadge } from "../../../components/ui/status-badge";
import { formatDateTime } from "../../../lib/format";
import { useSellerCatalog } from "../hooks/use-seller-catalog";

export function CatalogReviewTable({ sellerId, enabled }: { sellerId: string; enabled: boolean }) {
  const catalogQuery = useSellerCatalog(sellerId, enabled);

  if (!enabled) {
    return null;
  }

  if (catalogQuery.isLoading) {
    return <DataState title="Loading catalog" description="Fetching seller product moderation data." />;
  }

  if (catalogQuery.error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load catalog"
        description={
          catalogQuery.error instanceof Error
            ? catalogQuery.error.message
            : "The seller catalog request failed."
        }
        action={
          <button
            type="button"
            onClick={() => void catalogQuery.refetch()}
            className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
          >
            Retry
          </button>
        }
      />
    );
  }

  const products = catalogQuery.data?.products ?? [];

  return (
    <section className="space-y-3">
      <div>
        <h2 className="text-base font-semibold text-slate-950">Catalog Review</h2>
        <p className="mt-1 text-sm text-slate-600">Seller products in draft, pending, live, and rejected states.</p>
      </div>

      {products.length === 0 ? (
        <DataState title="No catalog items" description="This seller does not have products to review yet." />
      ) : (
        <div className="overflow-hidden rounded-lg border border-slate-200 bg-white">
          <div className="overflow-auto">
            <table className="min-w-full border-separate border-spacing-0 text-left text-sm">
              <thead className="bg-slate-100 text-xs uppercase text-slate-600">
                <tr>
                  <th className="border-b border-slate-200 px-4 py-3 font-semibold">Product</th>
                  <th className="border-b border-slate-200 px-4 py-3 font-semibold">Brand</th>
                  <th className="border-b border-slate-200 px-4 py-3 font-semibold">Category</th>
                  <th className="border-b border-slate-200 px-4 py-3 font-semibold">Status</th>
                  <th className="border-b border-slate-200 px-4 py-3 font-semibold">Updated</th>
                </tr>
              </thead>
              <tbody>
                {products.map((product) => (
                  <tr key={product.product_id} className="hover:bg-slate-50">
                    <td className="border-b border-slate-100 px-4 py-3">
                      <div className="flex min-w-0 items-center gap-3">
                        {product.image_url ? (
                          <img
                            src={product.image_url}
                            alt=""
                            className="h-10 w-10 shrink-0 rounded-md border border-slate-200 object-cover"
                          />
                        ) : (
                          <span className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md border border-slate-200 bg-slate-50 text-slate-500">
                            <ImageOff className="h-4 w-4" aria-hidden="true" />
                          </span>
                        )}
                        <div className="min-w-0">
                          <div className="truncate font-medium text-slate-950">{product.title}</div>
                          <div className="truncate text-xs text-slate-500">{product.product_id}</div>
                        </div>
                      </div>
                    </td>
                    <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                      {product.brand || "-"}
                    </td>
                    <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                      {product.category_id || "-"}
                    </td>
                    <td className="border-b border-slate-100 px-4 py-3">
                      <StatusBadge status={product.status} />
                    </td>
                    <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                      {formatDateTime(product.updated_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </section>
  );
}
