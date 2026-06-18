import type { OrderItem } from "../types";
import { MoneyCell } from "./money-cell";

type OrderItemsTableProps = {
  items: OrderItem[];
};

export function OrderItemsTable({ items }: OrderItemsTableProps) {
  if (items.length === 0) {
    return <p className="text-sm text-slate-500">No items available.</p>;
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[640px] border-collapse text-left text-sm">
        <thead className="border-b border-slate-200 text-xs uppercase tracking-wide text-slate-500">
          <tr>
            <th className="py-2 pr-3 font-semibold">Product</th>
            <th className="px-3 py-2 font-semibold">SKU</th>
            <th className="px-3 py-2 font-semibold">Qty</th>
            <th className="px-3 py-2 text-right font-semibold">Unit</th>
            <th className="py-2 pl-3 text-right font-semibold">Total</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100">
          {items.map((item, index) => (
            <tr key={`${item.product_id ?? "item"}-${item.variant_id ?? index}`}>
              <td className="py-3 pr-3">
                <p className="font-medium text-slate-950">
                  {item.title ?? item.product_id ?? "Product"}
                </p>
                {item.variant_id ? (
                  <p className="mt-0.5 text-xs text-slate-500">{item.variant_id}</p>
                ) : null}
              </td>
              <td className="px-3 py-3 text-slate-600">{item.sku ?? "-"}</td>
              <td className="px-3 py-3 text-slate-600">{item.quantity ?? 1}</td>
              <td className="px-3 py-3 text-right">
                <MoneyCell money={item.unit_price} />
              </td>
              <td className="py-3 pl-3 text-right">
                <MoneyCell money={item.total} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
