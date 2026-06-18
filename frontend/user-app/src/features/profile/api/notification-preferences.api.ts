import { apiGet, apiPatch } from '../../../lib/http';
import type { NotificationPreference } from '../types';

export function getNotificationPreferences(signal?: AbortSignal) {
  return apiGet<NotificationPreference>(
    '/api/v1/me/notification-preferences',
    {
      auth: true,
      signal,
    },
  );
}

export function updateNotificationPreferences(body: NotificationPreference) {
  return apiPatch<NotificationPreference, NotificationPreference>(
    '/api/v1/me/notification-preferences',
    body,
    { auth: true },
  );
}
