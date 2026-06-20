import { formatMoney } from "../../../lib/format";
import type { AdminOrderItem } from "../types";

export function OrderItemsTable({ items }: { items: AdminOrderItem[] }) {
  return (
    <section className="overflow-hidden rounded-lg border border-slate-200 bg-white">
      <div className="border-b border-slate-200 px-4 py-3">
        <h2 className="text-base font-semibold text-slate-950">Order Items</h2>
        <p className="mt-1 text-sm text-slate-600">Immutable item snapshots from checkout.</p>
      </div>

      {items.length === 0 ? (
        <p className="p-4 text-sm text-slate-600">No order items were returned for this order.</p>
      ) : (
        <div className="overflow-auto">
          <table className="min-w-full border-separate border-spacing-0 text-left text-sm">
            <thead className="bg-slate-100 text-xs uppercase text-slate-600">
              <tr>
                <th className="border-b border-slate-200 px-4 py-3 font-semibold">Item</th>
                <th className="border-b border-slate-200 px-4 py-3 font-semibold">Product</th>
                <th className="border-b border-slate-200 px-4 py-3 font-semibold">Seller</th>
                <th className="border-b border-slate-200 px-4 py-3 font-semibold">Fulfillment</th>
                <th className="border-b border-slate-200 px-4 py-3 text-right font-semibold">Qty</th>
                <th className="border-b border-slate-200 px-4 py-3 text-right font-semibold">Unit</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.item_id} className="hover:bg-slate-50">
                  <td className="border-b border-slate-100 px-4 py-3">
                    <div className="max-w-[260px] truncate font-medium text-slate-950">{item.title}</div>
                    <div className="truncate text-xs text-slate-500">{item.item_id}</div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                    <div className="max-w-[180px] truncate">{item.product_id ?? "Unknown"}</div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3 text-slate-700">
                    <div className="max-w-[180px] truncate">{item.seller_id ?? "Unknown"}</div>
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3 capitalize text-slate-700">
                    {item.fulfillment_status?.replace(/_/g, " ") ?? "Unknown"}
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3 text-right text-slate-700">
                    {item.quantity}
                  </td>
                  <td className="border-b border-slate-100 px-4 py-3 text-right font-medium text-slate-950">
                    {formatMoney(item.unit_price)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
