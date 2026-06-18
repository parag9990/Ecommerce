import { Trash2 } from 'lucide-react';

import { Button } from '../../../components/ui/button';
import type { CartItem } from '../types';
import { formatMoney } from './price';
import { QuantityStepper } from './quantity-stepper';

type CartItemRowProps = {
  disabled?: boolean | undefined;
  item: CartItem;
  onQuantityChange: (itemId: string, quantity: number) => void;
  onRemove: (itemId: string) => void;
};

function stockLabel(status: CartItem['stock_status']) {
  if (status === 'out_of_stock') {
    return { className: 'text-red-700', label: 'Out of stock' };
  }

  if (status === 'low_stock') {
    return { className: 'text-amber-700', label: 'Low stock' };
  }

  return undefined;
}

export function CartItemRow({
  disabled = false,
  item,
  onQuantityChange,
  onRemove,
}: CartItemRowProps) {
  const stock = stockLabel(item.stock_status);

  return (
    <article className="grid gap-4 border-b border-slate-200 py-4 last:border-b-0 sm:grid-cols-[6rem_1fr_auto]">
      <div className="aspect-square overflow-hidden rounded-md bg-slate-100">
        {item.image_url ? (
          <img
            alt={item.title ?? 'Cart item'}
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
        <h3 className="text-sm font-semibold text-slate-950">
          {item.title ?? 'Product'}
        </h3>
        {item.variant_label ? (
          <p className="mt-1 text-sm text-slate-600">{item.variant_label}</p>
        ) : null}
        {stock ? (
          <p className={`mt-2 text-sm font-medium ${stock.className}`}>
            {stock.label}
          </p>
        ) : null}
        <Button
          className="mt-3 h-8 px-2 text-red-700 hover:bg-red-50 hover:text-red-800"
          disabled={disabled}
          onClick={() => {
            onRemove(item.item_id);
          }}
          type="button"
          variant="ghost"
        >
          <Trash2 aria-hidden="true" className="mr-1 h-4 w-4" />
          Remove
        </Button>
      </div>

      <div className="flex items-center justify-between gap-4 sm:flex-col sm:items-end sm:justify-start">
        <QuantityStepper
          disabled={disabled || item.stock_status === 'out_of_stock'}
          onChange={(quantity) => {
            onQuantityChange(item.item_id, quantity);
          }}
          value={item.quantity}
        />
        <p className="text-sm font-semibold text-slate-950">
          {formatMoney(item.line_total ?? item.unit_price)}
        </p>
      </div>
    </article>
  );
}
