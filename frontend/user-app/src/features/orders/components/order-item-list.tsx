import { formatMoney } from '../../../lib/format-money';
import type { OrderItem } from '../types';

type OrderItemListProps = {
  items: OrderItem[];
};

export function OrderItemList({ items }: OrderItemListProps) {
  if (items.length === 0) {
    return (
      <p className="rounded-md border border-slate-200 bg-white p-4 text-sm text-slate-600">
        Item details are not available for this order.
      </p>
    );
  }

  return (
    <div className="divide-y divide-slate-100 rounded-md border border-slate-200 bg-white">
      {items.map((item, index) => (
        <div
          className="grid gap-4 p-4 sm:grid-cols-[4rem_minmax(0,1fr)_7rem]"
          key={`${item.product_id ?? 'item'}-${item.variant_id ?? index}`}
        >
          <div className="aspect-square overflow-hidden rounded-md bg-slate-100">
            {item.image_url ? (
              <img
                alt={item.title ?? 'Order item'}
                className="h-full w-full object-cover"
                src={item.image_url}
              />
            ) : (
              <div className="flex h-full items-center justify-center px-2 text-center text-xs text-slate-500">
                No image
              </div>
            )}
          </div>

          <div className="min-w-0">
            <h3 className="font-semibold text-slate-950">
              {item.title ?? 'Order item'}
            </h3>
            <p className="mt-1 text-sm text-slate-500">
              Quantity: {item.quantity ?? 1}
            </p>
            {item.variant_label ?? item.variant_id ? (
              <p className="mt-1 text-sm text-slate-500">
                Variant: {item.variant_label ?? item.variant_id}
              </p>
            ) : null}
          </div>

          <p className="text-left font-semibold text-slate-950 sm:text-right">
            {formatMoney(item.price)}
          </p>
        </div>
      ))}
    </div>
  );
}
