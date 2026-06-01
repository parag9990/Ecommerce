import { CheckCircle2, RotateCcw, XCircle } from 'lucide-react';
import { Link } from 'react-router-dom';

import { Button } from '../../../components/ui/button';
import { routePaths } from '../../../routes/route-paths';
import { formatMoney } from '../../cart/components/price';
import type { Order } from '../types';

type PaymentStatusCardProps = {
  error?: string | undefined;
  isLoading?: boolean | undefined;
  onRetry?: (() => void) | undefined;
  order?: Order | undefined;
  result: 'success' | 'failure';
  retrying?: boolean | undefined;
};

export function PaymentStatusCard({
  error,
  isLoading = false,
  onRetry,
  order,
  result,
  retrying = false,
}: PaymentStatusCardProps) {
  const isSuccess = result === 'success';

  return (
    <section className="mx-auto max-w-2xl rounded-md border border-slate-200 bg-white p-6 text-center">
      {isSuccess ? (
        <CheckCircle2 className="mx-auto h-12 w-12 text-emerald-600" />
      ) : (
        <XCircle className="mx-auto h-12 w-12 text-red-600" />
      )}

      <h1 className="mt-4 text-2xl font-semibold text-slate-950">
        {isSuccess ? 'Order placed' : 'Payment failed'}
      </h1>
      <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-slate-600">
        {isSuccess
          ? 'We received the payment response and are checking the latest order status.'
          : 'The payment was not completed. You can retry payment or return to your cart.'}
      </p>

      <dl className="mx-auto mt-5 grid max-w-sm gap-2 rounded-md bg-slate-50 p-3 text-left text-sm">
        <div className="flex justify-between gap-4">
          <dt className="text-slate-500">Order status</dt>
          <dd className="font-medium text-slate-950">
            {isLoading ? 'Loading' : order?.status ?? 'Unavailable'}
          </dd>
        </div>
        <div className="flex justify-between gap-4">
          <dt className="text-slate-500">Total</dt>
          <dd className="font-medium text-slate-950">
            {formatMoney(order?.total)}
          </dd>
        </div>
      </dl>

      {error ? <p className="mt-4 text-sm text-amber-700">{error}</p> : null}

      <div className="mt-6 flex flex-col justify-center gap-3 sm:flex-row">
        {isSuccess ? (
          <Link
            className="inline-flex h-11 items-center justify-center rounded-md bg-blue-600 px-4 text-sm font-semibold text-white transition hover:bg-blue-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
            to={routePaths.home}
          >
            Continue shopping
          </Link>
        ) : (
          <>
            <Button
              disabled={!onRetry || retrying}
              onClick={() => {
                onRetry?.();
              }}
              type="button"
            >
              <RotateCcw aria-hidden="true" className="mr-2 h-4 w-4" />
              {retrying ? 'Retrying' : 'Retry payment'}
            </Button>
            <Link
              className="inline-flex h-11 items-center justify-center rounded-md border border-slate-300 bg-white px-4 text-sm font-semibold text-slate-700 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
              to={routePaths.cart}
            >
              Back to cart
            </Link>
          </>
        )}
      </div>
    </section>
  );
}
