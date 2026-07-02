import type { LoginRequest, OtpChannel, OtpPurpose } from './types';
import { env } from '../../lib/env';

const otpPurposes: readonly OtpPurpose[] = [
  'email_verify',
  'login',
  'password_reset',
  'phone_verify',
  'signup',
];
const LOCAL_MAILPIT_URL = 'http://localhost:8025';

export function getDeviceInfo(): NonNullable<LoginRequest['device']> {
  return {
    channel: 'web',
    locale: navigator.language,
  };
}

export function detectOtpChannel(identifier: string): OtpChannel {
  return identifier.includes('@') ? 'email' : 'phone';
}

export function localEmailOtpNotice(target: string) {
  if (env.appEnv !== 'local' || detectOtpChannel(target) !== 'email') {
    return null;
  }

  return `Local email OTPs are captured in Mailpit at ${LOCAL_MAILPIT_URL}.`;
}

export function otpDeliverySuccessMessage(
  target: string,
  expiresInSeconds: number,
  label = 'OTP',
) {
  const expiry = `It expires in ${expiresInSeconds}s.`;

  if (env.appEnv === 'local' && detectOtpChannel(target) === 'email') {
    return `New ${label} sent to Mailpit at ${LOCAL_MAILPIT_URL}. ${expiry}`;
  }

  return `New ${label} sent. ${expiry}`;
}

export function otpDeliveryErrorMessage(
  target: string,
  error: unknown,
  fallback: string,
) {
  if (
    env.appEnv === 'local' &&
    detectOtpChannel(target) === 'phone' &&
    isApiErrorCode(error, 'OTP_DELIVERY_UNAVAILABLE')
  ) {
    return 'Phone OTP delivery is not enabled in the local stack. Use an email address for local testing.';
  }

  return error instanceof Error ? error.message : fallback;
}

export function isOtpPurpose(value: string | null): value is OtpPurpose {
  return value !== null && otpPurposes.includes(value as OtpPurpose);
}

export function safeRedirectPath(value: string | null, fallback = '/') {
  if (!value || !value.startsWith('/') || value.startsWith('//')) {
    return fallback;
  }

  return value;
}

export function toTrimmedValue(value: string) {
  return value.trim();
}

function isApiErrorCode(error: unknown, code: string) {
  return (
    typeof error === 'object' &&
    error !== null &&
    'code' in error &&
    error.code === code
  );
}
