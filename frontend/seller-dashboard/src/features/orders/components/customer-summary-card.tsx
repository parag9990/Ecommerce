import { MapPin, UserRound } from "lucide-react";

import type { CustomerSummary, ShippingAddress } from "../types";

type CustomerSummaryCardProps = {
  customer?: CustomerSummary;
  shippingAddress?: ShippingAddress;
};

function getAddressLines(address?: ShippingAddress) {
  if (!address) {
    return [];
  }

  return [
    address.name,
    address.line1,
    address.line2,
    [address.city, address.state, address.postal_code].filter(Boolean).join(", "),
    address.country,
    address.phone,
  ].filter((line): line is string => Boolean(line));
}

export function CustomerSummaryCard({
  customer,
  shippingAddress,
}: CustomerSummaryCardProps) {
  const addressLines = getAddressLines(shippingAddress);

  return (
    <section className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex items-center gap-2">
        <UserRound className="h-4 w-4 text-slate-500" aria-hidden="true" />
        <h2 className="text-sm font-semibold text-slate-950">Customer</h2>
      </div>

      <dl className="mt-3 grid gap-2 text-sm">
        <div className="flex justify-between gap-3">
          <dt className="text-slate-500">Name</dt>
          <dd className="truncate font-medium text-slate-800">{customer?.name ?? "-"}</dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-slate-500">Email</dt>
          <dd className="truncate font-medium text-slate-800">{customer?.email ?? "-"}</dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-slate-500">Phone</dt>
          <dd className="truncate font-medium text-slate-800">{customer?.phone ?? "-"}</dd>
        </div>
      </dl>

      <div className="mt-4 border-t border-slate-100 pt-4">
        <div className="flex items-center gap-2">
          <MapPin className="h-4 w-4 text-slate-500" aria-hidden="true" />
          <h3 className="text-sm font-semibold text-slate-950">Shipping</h3>
        </div>
        {addressLines.length === 0 ? (
          <p className="mt-3 text-sm text-slate-500">Shipping address unavailable.</p>
        ) : (
          <div className="mt-3 space-y-1 text-sm text-slate-700">
            {addressLines.map((line) => (
              <p key={line}>{line}</p>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
