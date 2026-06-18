import { Plus } from 'lucide-react';
import { useState } from 'react';

import { Alert } from '../../../components/ui/alert';
import { Button } from '../../../components/ui/button';
import { EmptyState } from '../../../components/ui/empty-state';
import type { AddressFormValues } from '../address-schema';
import { AddressForm } from '../components/address-form';
import { AddressList } from '../components/address-list';
import { useAddressMutations } from '../hooks/use-address-mutations';
import { useAddressesQuery } from '../hooks/use-addresses-query';
import type { Address, AddressInput } from '../types';

function getErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

function cleanOptional(value?: string) {
  const trimmed = value?.trim();

  return trimmed ? trimmed : undefined;
}

function toAddressInput(values: AddressFormValues): AddressInput {
  return {
    city: values.city.trim(),
    country: values.country.trim(),
    is_default: values.is_default,
    line1: values.line1.trim(),
    line2: cleanOptional(values.line2),
    name: values.name.trim(),
    phone: cleanOptional(values.phone),
    postal_code: values.postal_code.trim(),
    state: values.state.trim(),
  };
}

function AddressListSkeleton() {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      {Array.from({ length: 4 }, (_, index) => (
        <div
          className="h-44 animate-pulse rounded-md border border-slate-200 bg-white p-4"
          key={index}
        />
      ))}
    </div>
  );
}

export function AddressesPage() {
  const addressesQuery = useAddressesQuery();
  const addressMutations = useAddressMutations();
  const [actionError, setActionError] = useState<string>();
  const [success, setSuccess] = useState<string>();
  const [editingAddress, setEditingAddress] = useState<Address>();
  const [deleteCandidate, setDeleteCandidate] = useState<Address>();
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [savingAddressId, setSavingAddressId] = useState<string>();

  async function saveAddress(values: AddressFormValues) {
    const addressId = editingAddress?.address_id;

    setSavingAddressId(addressId ?? 'new');
    setActionError(undefined);
    setSuccess(undefined);

    try {
      if (addressId) {
        await addressMutations.updateAddress.mutateAsync({
          addressId,
          body: toAddressInput(values),
        });
      } else {
        await addressMutations.createAddress.mutateAsync(toAddressInput(values));
      }

      setEditingAddress(undefined);
      setIsFormOpen(false);
      setSuccess(addressId ? 'Address saved.' : 'Address added.');
    } catch (error) {
      setActionError(getErrorMessage(error, 'Address could not be saved.'));
    } finally {
      setSavingAddressId(undefined);
    }
  }

  async function confirmDelete() {
    const addressId = deleteCandidate?.address_id;

    if (!addressId) {
      setDeleteCandidate(undefined);
      setActionError('Address id is missing.');
      return;
    }

    setSavingAddressId(addressId);
    setActionError(undefined);
    setSuccess(undefined);

    try {
      await addressMutations.deleteAddress.mutateAsync(addressId);
      setDeleteCandidate(undefined);
      setSuccess('Address deleted.');
    } catch (error) {
      setActionError(getErrorMessage(error, 'Address could not be deleted.'));
    } finally {
      setSavingAddressId(undefined);
    }
  }

  const addresses = addressesQuery.data ?? [];
  const error =
    addressesQuery.error instanceof Error
      ? addressesQuery.error.message
      : addressesQuery.isError
        ? 'Addresses could not be loaded.'
        : undefined;
  const isEmpty = !addressesQuery.isLoading && addresses.length === 0;

  return (
    <section className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-slate-950">
            Addresses
          </h1>
          <p className="mt-1 text-sm text-slate-600">
            Manage saved delivery addresses for checkout.
          </p>
        </div>
        <Button
          onClick={() => {
            setEditingAddress(undefined);
            setIsFormOpen(true);
          }}
          type="button"
        >
          <Plus aria-hidden="true" className="mr-2 h-4 w-4" />
          Add address
        </Button>
      </div>

      {error ? <Alert variant="error">{error}</Alert> : null}
      {actionError ? <Alert variant="error">{actionError}</Alert> : null}
      {success ? <Alert variant="success">{success}</Alert> : null}

      {isFormOpen ? (
        <AddressForm
          address={editingAddress}
          isSaving={Boolean(savingAddressId)}
          onCancel={() => {
            setEditingAddress(undefined);
            setIsFormOpen(false);
          }}
          onSubmit={saveAddress}
        />
      ) : null}

      {addressesQuery.isLoading ? <AddressListSkeleton /> : null}

      {isEmpty ? (
        <EmptyState
          action={
            <Button
              onClick={() => {
                setEditingAddress(undefined);
                setIsFormOpen(true);
              }}
              type="button"
            >
              Add first address
            </Button>
          }
          description="Save an address now so checkout can move faster later."
          title="No saved addresses"
        />
      ) : null}

      {!addressesQuery.isLoading && addresses.length > 0 ? (
        <AddressList
          addresses={addresses}
          busyAddressId={savingAddressId}
          onDelete={(address) => {
            setDeleteCandidate(address);
          }}
          onEdit={(address) => {
            setEditingAddress(address);
            setIsFormOpen(true);
          }}
        />
      ) : null}

      {deleteCandidate ? (
        <div
          aria-modal="true"
          className="fixed inset-0 z-40 grid place-items-center bg-slate-950/30 px-4"
          role="dialog"
        >
          <div className="w-full max-w-md rounded-md bg-white p-5 shadow-xl">
            <h2 className="text-lg font-semibold text-slate-950">
              Delete address?
            </h2>
            <p className="mt-2 text-sm leading-6 text-slate-600">
              This removes the address for {deleteCandidate.name}. Existing
              orders keep their original delivery snapshot.
            </p>
            <div className="mt-5 flex flex-wrap justify-end gap-3">
              <Button
                disabled={Boolean(savingAddressId)}
                onClick={() => {
                  setDeleteCandidate(undefined);
                }}
                type="button"
                variant="secondary"
              >
                Keep address
              </Button>
              <Button
                className="bg-red-600 hover:bg-red-700 focus-visible:outline-red-600"
                disabled={Boolean(savingAddressId)}
                onClick={() => {
                  void confirmDelete();
                }}
                type="button"
              >
                Delete
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </section>
  );
}
