import { describe, expect, test } from 'vitest';

import {
  forgotPasswordSchema,
  loginSchema,
  otpSchema,
  resetPasswordSchema,
  signupSchema,
} from './schemas';

describe('auth schemas', () => {
  test('rejects weak signup passwords and mismatched confirmation', () => {
    const result = signupSchema.safeParse({
      acceptTerms: true,
      confirmPassword: 'different-password',
      email: 'buyer@example.com',
      full_name: 'Buyer User',
      password: 'weak',
      phone: '',
    });

    expect(result.success).toBe(false);
  });

  test('accepts a valid buyer signup form', () => {
    const result = signupSchema.safeParse({
      acceptTerms: true,
      confirmPassword: 'Strongpass1',
      email: 'buyer@example.com',
      full_name: 'Buyer User',
      password: 'Strongpass1',
      phone: '+15551234567',
    });

    expect(result.success).toBe(true);
  });

  test('validates login identifiers and OTP formats', () => {
    expect(
      loginSchema.safeParse({
        identifier: 'buyer@example.com',
        password: 'secret',
      }).success,
    ).toBe(true);

    expect(
      forgotPasswordSchema.safeParse({
        identifier: '+15551234567',
      }).success,
    ).toBe(true);

    expect(
      otpSchema.safeParse({
        challenge_id: 'challenge-1',
        otp: '123456',
      }).success,
    ).toBe(true);

    expect(
      otpSchema.safeParse({
        challenge_id: 'challenge-1',
        otp: '12345a',
      }).success,
    ).toBe(false);
  });

  test('requires reset password confirmation to match', () => {
    const result = resetPasswordSchema.safeParse({
      challenge_id: 'challenge-1',
      confirmPassword: 'Strongpass2',
      new_password: 'Strongpass1',
      otp: '123456',
    });

    expect(result.success).toBe(false);
  });
});
