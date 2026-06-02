import { ChevronLeft, ChevronRight, ImageIcon, Pencil } from "lucide-react";
import { Link } from "react-router-dom";

import { EmptyState } from "../../../components/state/empty-state";
import type { Category, Product } from "../types";
import { ProductStatusBadge } from "./product-status-badge";

type ProductTableProps = {
  products: Product[];
  total: number;
  page: number;
  pageSize: number;
  categories: Category[];
  onPageChange: (page: number) => void;
};

const moneyFormatter = new Intl.NumberFormat("en-IN", {
  style: "currency",
  currency: "INR",
  maximumFractionDigits: 0,
});

function formatPriceRange(product: Product) {
  const prices = product.variants
    .map((variant) => variant.price.amount)
    .filter((price) => Number.isFinite(price) && price > 0);

  if (!prices.length) {
    return "No price";
  }

  const min = Math.min(...prices);
  const max = Math.max(...prices);

  if (min === max) {
    return moneyFormatter.format(min);
  }

  return `${moneyFormatter.format(min)} - ${moneyFormatter.format(max)}`;
}

function getTotalStock(product: Product) {
  return product.variants.reduce((total, variant) => total + variant.stock_quantity, 0);
}

function getCategoryName(categories: Category[], categoryId: string) {
  return categories.find((category) => category.category_id === categoryId)?.name ?? categoryId;
}

export function ProductTable({
  products,
  total,
  page,
  pageSize,
  categories,
  onPageChange,
}: ProductTableProps) {
  const hasPreviousPage = page > 1;
  const hasNextPage = page * pageSize < total;

  if (products.length === 0) {
    return (
      <EmptyState
        title="No products"
        description="Create a product to start the catalog."
      />
    );
  }

  return (
    <div className="overflow-hidden rounded-md border border-slate-200 bg-white shadow-sm">
      <div className="overflow-x-auto">
        <table className="w-full min-w-[820px] border-collapse text-left text-sm">
          <thead className="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-4 py-3 font-semibold">Product</th>
              <th className="px-4 py-3 font-semibold">Category</th>
              <th className="px-4 py-3 font-semibold">Price</th>
              <th className="px-4 py-3 font-semibold">Stock</th>
              <th className="px-4 py-3 font-semibold">Status</th>
              <th className="px-4 py-3 text-right font-semibold">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {products.map((product) => (
              <tr key={product.product_id} className="align-middle">
                <td className="px-4 py-3">
                  <div className="flex min-w-0 items-center gap-3">
                    <div className="flex h-12 w-12 shrink-0 items-center justify-center overflow-hidden rounded-md border border-slate-200 bg-slate-50">
                      {product.images[0] ? (
                        <img
                          src={product.images[0]}
                          alt=""
                          className="h-full w-full object-cover"
                        />
                      ) : (
                        <ImageIcon className="h-5 w-5 text-slate-400" aria-hidden="true" />
                      )}
                    </div>
                    <div className="min-w-0">
                      <div className="truncate font-medium text-slate-950">{product.title}</div>
                      <div className="truncate text-xs text-slate-500">
                        {product.brand || "No brand"} - {product.variants.length} variants
                      </div>
                    </div>
                  </div>
                </td>
                <td className="px-4 py-3 text-slate-600">
                  {getCategoryName(categories, product.category_id)}
                </td>
                <td className="px-4 py-3 font-medium text-slate-800">
                  {formatPriceRange(product)}
                </td>
                <td className="px-4 py-3 text-slate-600">{getTotalStock(product)}</td>
                <td className="px-4 py-3">
                  <ProductStatusBadge status={product.status} />
                </td>
                <td className="px-4 py-3 text-right">
                  <Link
                    to={`/seller/products/${product.product_id}/edit`}
                    className="inline-flex h-8 items-center gap-1.5 rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
                  >
                    <Pencil className="h-3.5 w-3.5" aria-hidden="true" />
                    Edit
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 px-4 py-3 text-sm">
        <span className="text-slate-500">
          Page {page} - {total} total
        </span>
        <div className="flex items-center gap-2">
          <button
            type="button"
            aria-label="Previous page"
            title="Previous page"
            disabled={!hasPreviousPage}
            onClick={() => onPageChange(page - 1)}
            className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <ChevronLeft className="h-4 w-4" aria-hidden="true" />
          </button>
          <button
            type="button"
            aria-label="Next page"
            title="Next page"
            disabled={!hasNextPage}
            onClick={() => onPageChange(page + 1)}
            className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <ChevronRight className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
      </div>
    </div>
  );
}
