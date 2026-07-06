import type { Money } from '../types';

type PriceProps = {
  value?: Money | undefined;
};

function formatPrice(amount: number, currency: string) {
  const majorAmount = amount / 100;

  try {
    return new Intl.NumberFormat('en-IN', {
      currency,
      maximumFractionDigits: 2,
      minimumFractionDigits: 2,
      style: 'currency',
    }).format(majorAmount);
  } catch {
    return `${currency} ${majorAmount.toLocaleString('en-IN', {
      maximumFractionDigits: 2,
      minimumFractionDigits: 2,
    })}`;
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
