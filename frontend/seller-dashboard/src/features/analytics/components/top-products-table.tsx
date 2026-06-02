import type { TopProduct } from "../types";
import {
  formatMoney,
  formatNumber,
  formatPercentage,
} from "../utils/analytics-formatters";
import { AnalyticsEmptyState } from "./analytics-empty-state";

type TopProductsTableProps = {
  products: TopProduct[];
};

export function TopProductsTable({ products }: TopProductsTableProps) {
  if (products.length === 0) {
    return (
      <AnalyticsEmptyState
        title="Product performance"
        description="No top product rows were returned for this range."
        className="min-h-80"
      />
    );
  }

  return (
    <section className="rounded-md border border-slate-200 bg-white shadow-sm">
      <div className="border-b border-slate-200 p-4">
        <h2 className="text-base font-semibold text-slate-950">Product performance</h2>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-slate-200 text-sm">
          <thead className="bg-slate-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">
                Product
              </th>
              <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-slate-500">
                Revenue
              </th>
              <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-slate-500">
                GMV
              </th>
              <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-slate-500">
                Orders
              </th>
              <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-slate-500">
                Units
              </th>
              <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-slate-500">
                Conversion
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {products.map((product) => (
              <tr key={product.product_id} className="hover:bg-slate-50">
                <td className="min-w-56 px-4 py-3">
                  <p className="font-medium text-slate-950">{product.name}</p>
                  {product.sku ? (
                    <p className="mt-0.5 text-xs text-slate-500">SKU: {product.sku}</p>
                  ) : null}
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-right text-slate-700">
                  {formatMoney(product.revenue)}
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-right text-slate-700">
                  {formatMoney(product.gmv)}
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-right text-slate-700">
                  {formatNumber(product.orders)}
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-right text-slate-700">
                  {formatNumber(product.units_sold)}
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-right text-slate-700">
                  {formatPercentage(product.conversion_rate)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
