import { apiGet, apiPatch } from '../../../lib/http';
import type { UpdateUserProfileInput, UserProfile } from '../types';

export function getMyProfile(signal?: AbortSignal) {
  return apiGet<UserProfile>('/api/v1/me', { auth: true, signal });
}

export function updateMyProfile(body: UpdateUserProfileInput) {
  return apiPatch<UserProfile, UpdateUserProfileInput>('/api/v1/me', body, {
    auth: true,
  });
}
