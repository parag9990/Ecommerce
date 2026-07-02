import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest';

describe('auth OTP delivery messages', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_API_BASE_URL', '');
    vi.stubEnv('VITE_APP_ENV', 'local');
    vi.stubEnv('VITE_GRPC_WEB_BASE_URL', '');
    vi.stubEnv('VITE_GRPC_WEB_TIMEOUT_MS', '');
    vi.stubEnv('VITE_PAYMENT_PROVIDERS', '');
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  test('points local email OTPs to Mailpit', async () => {
    const { localEmailOtpNotice, otpDeliverySuccessMessage } = await import(
      './utils'
    );

    expect(localEmailOtpNotice('buyer@example.com')).toBe(
      'Local email OTPs are captured in Mailpit at http://localhost:8025.',
    );
    expect(otpDeliverySuccessMessage('buyer@example.com', 300)).toBe(
      'New OTP sent to Mailpit at http://localhost:8025. It expires in 300s.',
    );
  });

  test('builds login device info accepted by auth service', async () => {
    const { getDeviceInfo } = await import('./utils');

    expect(getDeviceInfo()).toEqual({
      channel: 'web',
      locale: navigator.language,
    });
    expect(getDeviceInfo()).not.toHaveProperty('timezone');
    expect(getDeviceInfo()).not.toHaveProperty('user_agent');
  });

  test('uses generic delivery copy outside local mode', async () => {
    vi.stubEnv('VITE_APP_ENV', 'production');
    vi.resetModules();

    const { localEmailOtpNotice, otpDeliverySuccessMessage } = await import(
      './utils'
    );

    expect(localEmailOtpNotice('buyer@example.com')).toBeNull();
    expect(otpDeliverySuccessMessage('buyer@example.com', 300)).toBe(
      'New OTP sent. It expires in 300s.',
    );
  });

  test('explains local phone delivery failures', async () => {
    const { otpDeliveryErrorMessage } = await import('./utils');

    expect(
      otpDeliveryErrorMessage(
        '+15551234567',
        { code: 'OTP_DELIVERY_UNAVAILABLE' },
        'Could not resend OTP.',
      ),
    ).toBe(
      'Phone OTP delivery is not enabled in the local stack. Use an email address for local testing.',
    );
  });
});
