import { zodResolver } from "@hookform/resolvers/zod";
import { Save } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { FormError } from "../../../components/state/form-error";
import { useFulfillmentUpdate } from "../hooks/use-fulfillment-update";
import type { FulfillmentUpdateInput, Order, OrderStatus } from "../types";
import { formatOrderStatus } from "../utils/order-formatters";
import { nextFulfillmentStatuses } from "../utils/order-status-rules";

const fulfillmentStatuses = ["packed", "shipped", "delivered"] as const;

const shipmentSchema = z.object({
  status: z.enum(fulfillmentStatuses),
  carrier: z.string().trim().max(80, "Carrier must be 80 characters or fewer").optional(),
  tracking_number: z
    .string()
    .trim()
    .max(80, "Tracking number must be 80 characters or fewer")
    .optional(),
});

type ShipmentFormValues = z.infer<typeof shipmentSchema>;

type ShipmentUpdateFormProps = {
  order: Order;
  onUpdated?: (order: Order) => void;
};

function isFulfillmentStatus(status: OrderStatus): status is ShipmentFormValues["status"] {
  return fulfillmentStatuses.some((item) => item === status);
}

function cleanOptional(value?: string) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

export function ShipmentUpdateForm({ order, onUpdated }: ShipmentUpdateFormProps) {
  const allowedStatuses = nextFulfillmentStatuses(order.status).filter(isFulfillmentStatus);
  const defaultStatus = allowedStatuses[0] ?? "packed";
  const mutation = useFulfillmentUpdate(order.order_id, { onSuccess: onUpdated });

  const form = useForm<ShipmentFormValues>({
    resolver: zodResolver(shipmentSchema),
    defaultValues: {
      status: defaultStatus,
      carrier: order.shipments?.[0]?.carrier ?? "",
      tracking_number: order.shipments?.[0]?.tracking_number ?? "",
    },
  });

  useEffect(() => {
    form.reset({
      status: defaultStatus,
      carrier: order.shipments?.[0]?.carrier ?? "",
      tracking_number: order.shipments?.[0]?.tracking_number ?? "",
    });
  }, [defaultStatus, form, order.order_id, order.shipments]);

  async function handleSubmit(values: ShipmentFormValues) {
    const input: FulfillmentUpdateInput = {
      status: values.status,
      carrier: cleanOptional(values.carrier),
      tracking_number: cleanOptional(values.tracking_number),
    };

    await mutation.mutateAsync(input);
  }

  return (
    <form
      className="space-y-3 rounded-md border border-slate-200 bg-white p-4 shadow-sm"
      onSubmit={form.handleSubmit(handleSubmit)}
    >
      <div>
        <h2 className="text-sm font-semibold text-slate-950">Update fulfillment</h2>
      </div>

      <label className="grid gap-1 text-sm font-medium text-slate-700">
        Next status
        <select
          {...form.register("status")}
          disabled={allowedStatuses.length === 0}
          className="h-9 rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {allowedStatuses.map((status) => (
            <option key={status} value={status}>
              {formatOrderStatus(status)}
            </option>
          ))}
        </select>
      </label>

      <label className="grid gap-1 text-sm font-medium text-slate-700">
        Carrier
        <input
          {...form.register("carrier")}
          placeholder="Delhivery, Blue Dart, DTDC"
          className="h-9 rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
        />
      </label>
      {form.formState.errors.carrier ? (
        <p className="text-xs text-rose-600">{form.formState.errors.carrier.message}</p>
      ) : null}

      <label className="grid gap-1 text-sm font-medium text-slate-700">
        Tracking number
        <input
          {...form.register("tracking_number")}
          placeholder="Tracking number"
          className="h-9 rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
        />
      </label>
      {form.formState.errors.tracking_number ? (
        <p className="text-xs text-rose-600">
          {form.formState.errors.tracking_number.message}
        </p>
      ) : null}

      {mutation.isError ? (
        <FormError error={mutation.error} />
      ) : null}

      {mutation.isSuccess ? (
        <p className="rounded-md border border-emerald-200 bg-emerald-50 p-2 text-sm text-emerald-700">
          Fulfillment updated.
        </p>
      ) : null}

      <button
        type="submit"
        disabled={mutation.isPending || allowedStatuses.length === 0}
        className="inline-flex h-9 w-full items-center justify-center gap-2 rounded-md bg-slate-950 px-3 text-sm font-medium text-white transition hover:bg-slate-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-950 disabled:cursor-not-allowed disabled:opacity-60"
      >
        <Save className="h-4 w-4" aria-hidden="true" />
        {mutation.isPending ? "Updating..." : "Update shipment"}
      </button>
    </form>
  );
}
