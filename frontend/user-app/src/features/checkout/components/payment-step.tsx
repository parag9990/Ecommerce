import { CreditCard } from 'lucide-react';

import { env } from '../../../lib/env';

type PaymentStepProps = {
  onProviderChange: (provider: string) => void;
  selectedProvider: string;
};

const providerLabels: Record<string, string> = {
  razorpay: 'UPI / Razorpay',
  stripe: 'Card / Stripe',
};

function formatProviderLabel(provider: string) {
  return providerLabels[provider] ?? provider.replaceAll('_', ' ').toUpperCase();
}

export function PaymentStep({
  onProviderChange,
  selectedProvider,
}: PaymentStepProps) {
  return (
    <section>
      <div className="flex items-center gap-2">
        <CreditCard aria-hidden="true" className="h-5 w-5 text-blue-700" />
        <h2 className="text-base font-semibold text-slate-950">
          Payment method
        </h2>
      </div>

      <div className="mt-3 grid gap-3" role="radiogroup">
        {env.paymentProviders.map((provider) => (
          <label
            className="flex cursor-pointer gap-3 rounded-md border border-slate-200 bg-white p-4 transition has-[:checked]:border-blue-600 has-[:checked]:ring-2 has-[:checked]:ring-blue-100"
            key={provider}
          >
            <input
              checked={selectedProvider === provider}
              className="mt-1 h-4 w-4 border-slate-300 text-blue-600 focus:ring-blue-500"
              name="payment_provider"
              onChange={() => {
                onProviderChange(provider);
              }}
              type="radio"
              value={provider}
            />
            <span className="text-sm font-semibold text-slate-950">
              {formatProviderLabel(provider)}
            </span>
          </label>
        ))}
      </div>
    </section>
  );
}
