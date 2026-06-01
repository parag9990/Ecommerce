import { Link } from 'react-router-dom';

import { routePaths } from '../../../routes/route-paths';
import type { Cart } from '../types';
import { formatMoney } from './price';

type CartSummaryProps = {
  cart?: Cart | undefined;
  checkoutDisabled?: boolean | undefined;
};

export function CartSummary({
  cart,
  checkoutDisabled = false,
}: CartSummaryProps) {
  const itemCount = cart?.items?.reduce((sum, item) => sum + item.quantity, 0) ?? 0;

  return (
    <aside className="rounded-md border border-slate-200 bg-white p-4">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-base font-semibold text-slate-950">
          Order summary
        </h2>
        <span className="text-sm text-slate-500">
          {itemCount} {itemCount === 1 ? 'item' : 'items'}
        </span>
      </div>

      <dl className="mt-4 space-y-3 text-sm">
        <div className="flex justify-between gap-4">
          <dt className="text-slate-600">Subtotal</dt>
          <dd className="font-medium">{formatMoney(cart?.subtotal)}</dd>
        </div>
        <div className="flex justify-between gap-4">
          <dt className="text-slate-600">Discount</dt>
          <dd className="font-medium text-emerald-700">
            -{formatMoney(cart?.discount)}
          </dd>
        </div>
        <div className="border-t border-slate-200 pt-3">
          <div className="flex justify-between gap-4 text-base font-semibold">
            <dt>Total</dt>
            <dd>{formatMoney(cart?.total)}</dd>
          </div>
        </div>
      </dl>

      <Link
        aria-disabled={checkoutDisabled}
        className="mt-5 block rounded-md bg-blue-600 px-4 py-2.5 text-center text-sm font-semibold text-white transition hover:bg-blue-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 aria-disabled:pointer-events-none aria-disabled:bg-slate-400"
        onClick={(event) => {
          if (checkoutDisabled) {
            event.preventDefault();
          }
        }}
        to={routePaths.checkout}
      >
        Proceed to checkout
      </Link>
    </aside>
  );
}
