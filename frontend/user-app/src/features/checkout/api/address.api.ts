import { apiGet } from '../../../lib/http';
import type { AddressListResponse } from '../types';

export async function listCheckoutAddresses(signal?: AbortSignal) {
  const response = await apiGet<AddressListResponse>('/api/v1/me/addresses', {
    auth: true,
    signal,
  });

  return response.addresses ?? [];
}
