import { ExternalLink, ShieldCheck } from 'lucide-react';

import { Alert } from '../../../components/ui/alert';
import { Button } from '../../../components/ui/button';
import { formatMoney } from '../../cart/components/price';
import type { PaymentIntentResponse } from '../types';

type ProviderPaymentWidgetProps = {
  paymentIntent: PaymentIntentResponse;
};

function getProviderRedirectUrl(paymentIntent: PaymentIntentResponse) {
  const candidate =
    paymentIntent.redirect_url ??
    paymentIntent.payment_url ??
    paymentIntent.next_action_url;

  if (!candidate) {
    return undefined;
  }

  try {
    const url = new URL(candidate, window.location.origin);

    if (url.protocol === 'https:' || url.protocol === 'http:') {
      return url.toString();
    }
  } catch {
    return undefined;
  }

  return undefined;
}

export function ProviderPaymentWidget({
  paymentIntent,
}: ProviderPaymentWidgetProps) {
  const providerRedirectUrl = getProviderRedirectUrl(paymentIntent);

  if (!paymentIntent.client_secret && !providerRedirectUrl) {
    return (
      <Alert title="Payment could not be started" variant="error">
        The payment provider response was incomplete. Please retry checkout.
      </Alert>
    );
  }

  return (
    <section className="rounded-md border border-slate-200 bg-white p-5">
      <div className="flex items-start gap-3">
        <ShieldCheck aria-hidden="true" className="mt-0.5 h-6 w-6 text-blue-700" />
        <div className="min-w-0">
          <h2 className="text-base font-semibold text-slate-950">
            Payment ready
          </h2>
          <dl className="mt-3 grid gap-2 text-sm sm:grid-cols-2">
            <div>
              <dt className="text-slate-500">Provider</dt>
              <dd className="font-medium text-slate-950">
                {paymentIntent.provider ?? 'Selected provider'}
              </dd>
            </div>
            <div>
              <dt className="text-slate-500">Amount</dt>
              <dd className="font-medium text-slate-950">
                {formatMoney(paymentIntent.amount)}
              </dd>
            </div>
            {paymentIntent.status ? (
              <div>
                <dt className="text-slate-500">Status</dt>
                <dd className="font-medium text-slate-950">
                  {paymentIntent.status}
                </dd>
              </div>
            ) : null}
          </dl>
        </div>
      </div>

      <div
        className="mt-5 rounded-md border border-dashed border-slate-300 bg-slate-50 p-4"
        data-client-secret-ready={Boolean(paymentIntent.client_secret)}
        data-payment-provider={paymentIntent.provider}
        id="payment-provider-mount"
      >
        <p className="text-sm text-slate-600">
          Secure payment UI will appear here when the configured provider SDK is
          available.
        </p>
      </div>

      {providerRedirectUrl ? (
        <Button
          className="mt-5"
          onClick={() => {
            window.location.assign(providerRedirectUrl);
          }}
          type="button"
        >
          <ExternalLink aria-hidden="true" className="mr-2 h-4 w-4" />
          Continue to payment
        </Button>
      ) : null}
    </section>
  );
}
