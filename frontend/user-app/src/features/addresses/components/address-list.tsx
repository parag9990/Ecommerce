import { AddressCard } from './address-card';
import type { Address } from '../types';

type AddressListProps = {
  addresses: Address[];
  busyAddressId?: string | undefined;
  onDelete: (address: Address) => void;
  onEdit: (address: Address) => void;
};

export function AddressList({
  addresses,
  busyAddressId,
  onDelete,
  onEdit,
}: AddressListProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      {addresses.map((address) => (
        <AddressCard
          address={address}
          isBusy={address.address_id === busyAddressId}
          key={address.address_id ?? `${address.name}-${address.postal_code}`}
          onDelete={onDelete}
          onEdit={onEdit}
        />
      ))}
    </div>
  );
}
