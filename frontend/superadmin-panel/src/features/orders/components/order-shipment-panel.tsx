import { Truck } from "lucide-react";

import { StatusBadge } from "../../../components/ui/status-badge";
import { formatDateTime } from "../../../lib/format";
import type { OrderShipment } from "../types";

export function OrderShipmentPanel({ shipments }: { shipments: OrderShipment[] }) {
  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="flex items-center gap-2">
        <Truck className="h-4 w-4 text-slate-600" aria-hidden="true" />
        <h2 className="text-base font-semibold text-slate-950">Shipments</h2>
      </div>

      {shipments.length === 0 ? (
        <p className="mt-3 text-sm text-slate-600">No shipment records were returned for this order.</p>
      ) : (
        <div className="mt-4 space-y-3">
          {shipments.map((shipment) => (
            <article key={shipment.shipment_id} className="rounded-lg border border-slate-200 p-3 text-sm">
              <div className="flex flex-wrap items-start justify-between gap-2">
                <div className="min-w-0">
                  <div className="truncate font-medium text-slate-950">{shipment.shipment_id}</div>
                  <div className="mt-1 text-xs text-slate-500">{shipment.seller_id ?? "Unknown seller"}</div>
                </div>
                <StatusBadge status={shipment.status} />
              </div>
              <dl className="mt-3 grid gap-2 text-slate-700">
                <div className="flex justify-between gap-3">
                  <dt className="text-slate-600">Carrier</dt>
                  <dd>{shipment.carrier ?? "Not added"}</dd>
                </div>
                <div className="flex justify-between gap-3">
                  <dt className="text-slate-600">Tracking</dt>
                  <dd className="break-all text-right">{shipment.tracking_number ?? "Not added"}</dd>
                </div>
                <div className="flex justify-between gap-3">
                  <dt className="text-slate-600">Shipped</dt>
                  <dd className="text-right">{formatDateTime(shipment.shipped_at)}</dd>
                </div>
                <div className="flex justify-between gap-3">
                  <dt className="text-slate-600">Delivered</dt>
                  <dd className="text-right">{formatDateTime(shipment.delivered_at)}</dd>
                </div>
              </dl>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}
