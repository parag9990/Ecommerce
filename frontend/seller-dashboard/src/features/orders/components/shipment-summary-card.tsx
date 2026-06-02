import { Truck } from "lucide-react";

import type { Shipment } from "../types";
import { formatDateTime } from "../utils/order-formatters";
import { OrderStatusBadge } from "./order-status-badge";

type ShipmentSummaryCardProps = {
  shipments: Shipment[];
};

export function ShipmentSummaryCard({ shipments }: ShipmentSummaryCardProps) {
  const latestShipment = shipments[0];

  return (
    <section className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex items-center gap-2">
        <Truck className="h-4 w-4 text-slate-500" aria-hidden="true" />
        <h2 className="text-sm font-semibold text-slate-950">Shipment</h2>
      </div>

      {!latestShipment ? (
        <p className="mt-3 text-sm text-slate-500">Shipment info has not been added.</p>
      ) : (
        <dl className="mt-3 grid gap-2 text-sm">
          {latestShipment.status ? (
            <div className="flex items-center justify-between gap-3">
              <dt className="text-slate-500">Status</dt>
              <dd>
                <OrderStatusBadge status={latestShipment.status} />
              </dd>
            </div>
          ) : null}
          <div className="flex justify-between gap-3">
            <dt className="text-slate-500">Carrier</dt>
            <dd className="truncate font-medium text-slate-800">
              {latestShipment.carrier ?? "-"}
            </dd>
          </div>
          <div className="flex justify-between gap-3">
            <dt className="text-slate-500">Tracking</dt>
            <dd className="truncate font-medium text-slate-800">
              {latestShipment.tracking_number ?? "-"}
            </dd>
          </div>
          <div className="flex justify-between gap-3">
            <dt className="text-slate-500">Shipped</dt>
            <dd className="font-medium text-slate-800">
              {formatDateTime(latestShipment.shipped_at)}
            </dd>
          </div>
          <div className="flex justify-between gap-3">
            <dt className="text-slate-500">Delivered</dt>
            <dd className="font-medium text-slate-800">
              {formatDateTime(latestShipment.delivered_at)}
            </dd>
          </div>
        </dl>
      )}
    </section>
  );
}
