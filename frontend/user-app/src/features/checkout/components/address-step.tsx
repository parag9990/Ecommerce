import { MapPin } from 'lucide-react';

import { Alert } from '../../../components/ui/alert';
import type { Address } from '../types';

type AddressStepProps = {
  addresses: Address[];
  onSelect: (addressId: string) => void;
  selectedAddressId?: string | undefined;
};

function formatAddress(address: Address) {
  return [
    address.line1,
    address.line2,
    address.city,
    address.state,
    address.postal_code,
    address.country,
  ]
    .filter(Boolean)
    .join(', ');
}

export function AddressStep({
  addresses,
  onSelect,
  selectedAddressId,
}: AddressStepProps) {
  if (addresses.length === 0) {
    return (
      <Alert title="No saved delivery address" variant="error">
        Add an address from your profile before checkout.
      </Alert>
    );
  }

  return (
    <section>
      <div className="flex items-center gap-2">
        <MapPin aria-hidden="true" className="h-5 w-5 text-blue-700" />
        <h2 className="text-base font-semibold text-slate-950">
          Delivery address
        </h2>
      </div>

      <div className="mt-3 grid gap-3" role="radiogroup">
        {addresses.map((address) => {
          const addressId = address.address_id;

          return (
            <label
              className="flex cursor-pointer gap-3 rounded-md border border-slate-200 bg-white p-4 transition has-[:checked]:border-blue-600 has-[:checked]:ring-2 has-[:checked]:ring-blue-100"
              key={addressId ?? formatAddress(address)}
            >
              <input
                checked={addressId !== undefined && selectedAddressId === addressId}
                className="mt-1 h-4 w-4 border-slate-300 text-blue-600 focus:ring-blue-500"
                disabled={!addressId}
                name="address_id"
                onChange={() => {
                  if (addressId) {
                    onSelect(addressId);
                  }
                }}
                type="radio"
                value={addressId}
              />
              <span className="text-sm leading-6">
                <span className="block font-semibold text-slate-950">
                  {address.name}
                  {address.is_default ? (
                    <span className="ml-2 rounded bg-blue-50 px-2 py-0.5 text-xs font-semibold text-blue-700">
                      Default
                    </span>
                  ) : null}
                </span>
                <span className="block text-slate-600">
                  {formatAddress(address)}
                </span>
                {address.phone ? (
                  <span className="block text-slate-600">{address.phone}</span>
                ) : null}
              </span>
            </label>
          );
        })}
      </div>
    </section>
  );
}
