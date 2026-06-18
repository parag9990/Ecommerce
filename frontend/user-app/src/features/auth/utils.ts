import type { LoginRequest, OtpChannel, OtpPurpose } from './types';

const otpPurposes: readonly OtpPurpose[] = [
  'email_verify',
  'login',
  'password_reset',
  'phone_verify',
  'signup',
];

export function getDeviceInfo(): NonNullable<LoginRequest['device']> {
  return {
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    user_agent: navigator.userAgent,
  };
}

export function detectOtpChannel(identifier: string): OtpChannel {
  return identifier.includes('@') ? 'email' : 'phone';
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
