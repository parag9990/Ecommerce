import { formatMoney } from "../../lib/format";

export function AmountText({
  money,
  className
}: {
  money?: { amount?: number | null; currency?: string | null } | null;
  className?: string;
}) {
  return <span className={className}>{formatMoney(money)}</span>;
}
