import { cn } from "../../../lib/cn";
import type { OrderStatus } from "../types";
import { formatOrderStatus } from "../utils/order-formatters";

type OrderStatusBadgeProps = {
  status: OrderStatus;
  className?: string;
};

const statusStyles: Record<OrderStatus, string> = {
  created: "border-slate-200 bg-slate-100 text-slate-700",
  pending_payment: "border-amber-200 bg-amber-50 text-amber-700",
  paid: "border-emerald-200 bg-emerald-50 text-emerald-700",
  packed: "border-blue-200 bg-blue-50 text-blue-700",
  shipped: "border-violet-200 bg-violet-50 text-violet-700",
  delivered: "border-emerald-300 bg-emerald-100 text-emerald-800",
  cancelled: "border-rose-200 bg-rose-50 text-rose-700",
  refunded: "border-orange-200 bg-orange-50 text-orange-700",
};

export function OrderStatusBadge({ status, className }: OrderStatusBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex h-6 items-center rounded border px-2 text-xs font-medium capitalize leading-none",
        statusStyles[status],
        className,
      )}
    >
      {formatOrderStatus(status)}
    </span>
  );
}
