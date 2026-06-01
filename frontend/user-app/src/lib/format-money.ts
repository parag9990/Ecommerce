export type Money = {
  amount?: number | undefined;
  currency?: string | undefined;
};

export function formatMoney(money?: Money) {
  if (money?.amount === undefined) {
    return 'N/A';
  }

  const currency = money.currency ?? 'INR';
  const displayAmount = money.amount / 100;

  try {
    return new Intl.NumberFormat('en-IN', {
      currency,
      maximumFractionDigits: 2,
      minimumFractionDigits: 2,
      style: 'currency',
    }).format(displayAmount);
  } catch {
    return `${currency} ${displayAmount.toLocaleString('en-IN', {
      maximumFractionDigits: 2,
      minimumFractionDigits: 2,
    })}`;
  }
}
