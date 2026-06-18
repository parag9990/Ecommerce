import { apiDelete, apiGet, apiPatch, apiPost } from '../../../lib/http';
import type {
  Address,
  AddressInput,
  AddressListResponse,
  SuccessResponse,
} from '../types';

export async function listAddresses(signal?: AbortSignal) {
  const response = await apiGet<AddressListResponse>('/api/v1/me/addresses', {
    auth: true,
    signal,
  });

  return response.addresses ?? [];
}

export function createAddress(body: AddressInput) {
  return apiPost<Address, AddressInput>('/api/v1/me/addresses', body, {
    auth: true,
  });
}

export function updateAddress(addressId: string, body: AddressInput) {
  return apiPatch<Address, AddressInput>(
    `/api/v1/me/addresses/${encodeURIComponent(addressId)}`,
    body,
    { auth: true },
  );
}

export function deleteAddress(addressId: string) {
  return apiDelete<SuccessResponse>(
    `/api/v1/me/addresses/${encodeURIComponent(addressId)}`,
    { auth: true },
  );
}
