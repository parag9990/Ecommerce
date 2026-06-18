import { Edit3, Star, Trash2 } from 'lucide-react';

import type { Address } from '../types';

type AddressCardProps = {
  address: Address;
  isBusy?: boolean | undefined;
  onDelete: (address: Address) => void;
  onEdit: (address: Address) => void;
};

export function AddressCard({
  address,
  isBusy = false,
  onDelete,
  onEdit,
}: AddressCardProps) {
  return (
    <article className="rounded-md border border-slate-200 bg-white p-4">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="font-semibold text-slate-950">{address.name}</h2>
            {address.is_default ? (
              <span className="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-1 text-xs font-semibold text-emerald-700">
                <Star aria-hidden="true" className="h-3 w-3" />
                Default
              </span>
            ) : null}
          </div>

          <address className="mt-2 not-italic text-sm leading-6 text-slate-600">
            {address.line1}
            {address.line2 ? (
              <>
                <br />
                {address.line2}
              </>
            ) : null}
            <br />
            {address.city}, {address.state} {address.postal_code}
            <br />
            {address.country}
          </address>

          {address.phone ? (
            <p className="mt-2 text-sm text-slate-500">Phone: {address.phone}</p>
          ) : null}
        </div>

        <div className="flex shrink-0 gap-1">
          <button
            aria-label={`Edit address for ${address.name}`}
            className="inline-flex h-9 w-9 items-center justify-center rounded-md text-slate-600 transition hover:bg-slate-100 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:text-slate-300"
            disabled={isBusy}
            onClick={() => {
              onEdit(address);
            }}
            title="Edit address"
            type="button"
          >
            <Edit3 aria-hidden="true" className="h-4 w-4" />
          </button>
          <button
            aria-label={`Delete address for ${address.name}`}
            className="inline-flex h-9 w-9 items-center justify-center rounded-md text-red-700 transition hover:bg-red-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600 disabled:cursor-not-allowed disabled:text-slate-300"
            disabled={isBusy}
            onClick={() => {
              onDelete(address);
            }}
            title="Delete address"
            type="button"
          >
            <Trash2 aria-hidden="true" className="h-4 w-4" />
          </button>
        </div>
      </div>
    </article>
  );
}
