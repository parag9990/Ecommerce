import { describe, expect, test } from 'vitest';

import {
  notificationPreferencesSchema,
  profileSchema,
} from './profile-schema';

describe('profile schemas', () => {
  test('validates profile edits', () => {
    expect(
      profileSchema.safeParse({
        avatar_url: '',
        full_name: 'Buyer User',
        phone: '+15551234567',
      }).success,
    ).toBe(true);

    expect(
      profileSchema.safeParse({
        avatar_url: 'not-a-url',
        full_name: 'B',
        phone: '',
      }).success,
    ).toBe(false);
  });

  test('validates notification preferences', () => {
    expect(
      notificationPreferencesSchema.safeParse({
        email_enabled: true,
        marketing_enabled: false,
        push_enabled: true,
        sms_enabled: false,
      }).success,
    ).toBe(true);
  });
});
