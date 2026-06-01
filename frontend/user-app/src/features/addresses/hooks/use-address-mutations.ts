import { useMutation, useQueryClient } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import {
  createAddress,
  deleteAddress,
  updateAddress,
} from '../api/addresses.api';
import type { AddressInput } from '../types';

export function useAddressMutations() {
  const queryClient = useQueryClient();

  const refreshAddresses = () =>
    queryClient.invalidateQueries({ queryKey: queryKeys.profile.addresses() });

  const createAddressMutation = useMutation({
    mutationFn: (body: AddressInput) => createAddress(body),
    onSuccess: refreshAddresses,
  });

  const updateAddressMutation = useMutation({
    mutationFn: (input: { addressId: string; body: AddressInput }) =>
      updateAddress(input.addressId, input.body),
    onSuccess: refreshAddresses,
  });

  const deleteAddressMutation = useMutation({
    mutationFn: (addressId: string) => deleteAddress(addressId),
    onSuccess: refreshAddresses,
  });

  return {
    createAddress: createAddressMutation,
    deleteAddress: deleteAddressMutation,
    updateAddress: updateAddressMutation,
  };
}
