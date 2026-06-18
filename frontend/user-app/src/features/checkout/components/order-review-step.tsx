import { ClipboardList } from 'lucide-react';

import { formatMoney } from '../../cart/components/price';
import type { Cart } from '../../cart/types';

type OrderReviewStepProps = {
  cart?: Cart | undefined;
  couponCode?: string | undefined;
};

export function OrderReviewStep({ cart, couponCode }: OrderReviewStepProps) {
  const items = cart?.items ?? [];

  return (
    <section>
      <div className="flex items-center gap-2">
        <ClipboardList aria-hidden="true" className="h-5 w-5 text-blue-700" />
        <h2 className="text-base font-semibold text-slate-950">
          Review order
        </h2>
      </div>

      <div className="mt-3 divide-y divide-slate-200 rounded-md border border-slate-200 bg-white">
        {items.map((item) => (
          <div
            className="flex justify-between gap-4 p-3 text-sm"
            key={item.item_id}
          >
            <div className="min-w-0">
              <p className="truncate font-medium text-slate-950">
                {item.title ?? 'Product'}
              </p>
              <p className="text-slate-600">Qty: {item.quantity}</p>
            </div>
            <p className="shrink-0 font-medium">
              {formatMoney(item.line_total ?? item.unit_price)}
            </p>
          </div>
        ))}
      </div>

      <dl className="mt-4 space-y-2 text-sm">
        <div className="flex justify-between gap-4">
          <dt className="text-slate-600">Subtotal</dt>
          <dd>{formatMoney(cart?.subtotal)}</dd>
        </div>
        <div className="flex justify-between gap-4">
          <dt className="text-slate-600">
            Discount {couponCode ? `(${couponCode})` : ''}
          </dt>
          <dd className="text-emerald-700">-{formatMoney(cart?.discount)}</dd>
        </div>
        <div className="flex justify-between gap-4 border-t border-slate-200 pt-2 font-semibold">
          <dt>Total</dt>
          <dd>{formatMoney(cart?.total)}</dd>
        </div>
      </dl>
    </section>
  );
}
