import { cn } from "../../../lib/cn";
import type { Money } from "../types";
import { formatMoney } from "../utils/order-formatters";

type MoneyCellProps = {
  money?: Money | null;
  className?: string;
};

export function MoneyCell({ money, className }: MoneyCellProps) {
  return (
    <span className={cn("font-medium tabular-nums text-slate-800", className)}>
      {formatMoney(money)}
    </span>
  );
}
