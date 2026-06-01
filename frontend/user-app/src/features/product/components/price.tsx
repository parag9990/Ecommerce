import type { Money } from '../types';

type PriceProps = {
  value?: Money | undefined;
};

function formatPrice(amount: number, currency: string) {
  try {
    return new Intl.NumberFormat('en-IN', {
      currency,
      maximumFractionDigits: 0,
      style: 'currency',
    }).format(amount);
  } catch {
    return `${currency} ${amount.toLocaleString('en-IN')}`;
  }
}

export function Price({ value }: PriceProps) {
  if (value?.amount === undefined) {
    return <span className="text-sm text-slate-500">Price unavailable</span>;
  }

  return (
    <span className="font-semibold text-slate-950">
      {formatPrice(value.amount, value.currency ?? 'INR')}
    </span>
  );
}
