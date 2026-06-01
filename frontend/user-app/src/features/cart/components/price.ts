import type { Money } from '../types';

export function formatMoney(money?: Money) {
  const amount = money?.amount ?? 0;
  const currency = money?.currency ?? 'INR';

  try {
    return new Intl.NumberFormat('en-IN', {
      currency,
      maximumFractionDigits: 2,
      minimumFractionDigits: 2,
      style: 'currency',
    }).format(amount / 100);
  } catch {
    return `${currency} ${(amount / 100).toLocaleString('en-IN', {
      maximumFractionDigits: 2,
      minimumFractionDigits: 2,
    })}`;
  }
}
