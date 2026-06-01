import { z } from 'zod';

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const phonePattern = /^\+?[0-9][0-9\s()-]{6,19}$/;

export const passwordRequirements = [
  {
    id: 'length',
    label: 'At least 8 characters',
    test: (value: string) => value.length >= 8,
  },
  {
    id: 'uppercase',
    label: 'One uppercase letter',
    test: (value: string) => /[A-Z]/.test(value),
  },
  {
    id: 'lowercase',
    label: 'One lowercase letter',
    test: (value: string) => /[a-z]/.test(value),
  },
  {
    id: 'number',
    label: 'One number',
    test: (value: string) => /[0-9]/.test(value),
  },
] as const;

function isEmailOrPhone(value: string): boolean {
  return emailPattern.test(value) || phonePattern.test(value);
}

export const passwordSchema = z
  .string()
  .min(8, 'Password must be at least 8 characters.')
  .max(128, 'Password must be 128 characters or fewer.')
  .regex(/[A-Z]/, 'Password must include an uppercase letter.')
  .regex(/[a-z]/, 'Password must include a lowercase letter.')
  .regex(/[0-9]/, 'Password must include a number.');

export const loginSchema = z.object({
  identifier: z
    .string()
    .trim()
    .min(1, 'Email or phone is required.')
    .max(254, 'Email or phone is too long.')
    .refine(isEmailOrPhone, 'Enter a valid email or phone number.'),
  password: z.string().min(1, 'Password is required.'),
});

export const signupSchema = z
  .object({
    acceptTerms: z
      .boolean()
      .refine((value) => value, 'You must accept the terms and privacy policy.'),
    confirmPassword: z.string().min(1, 'Confirm your password.'),
    email: z
      .string()
      .trim()
      .min(1, 'Email is required.')
      .email('Enter a valid email address.'),
    full_name: z
      .string()
      .trim()
      .min(2, 'Full name must be at least 2 characters.')
      .max(120, 'Full name must be 120 characters or fewer.'),
    password: passwordSchema,
    phone: z
      .string()
      .trim()
      .max(24, 'Phone number is too long.')
      .refine(
        (value) => value.length === 0 || phonePattern.test(value),
        'Enter a valid phone number.',
      ),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match.',
    path: ['confirmPassword'],
  });

export const otpSchema = z.object({
  challenge_id: z.string().trim().min(1, 'OTP challenge is missing.'),
  otp: z.string().trim().regex(/^[0-9]{6}$/, 'Enter the 6 digit OTP.'),
});

export const forgotPasswordSchema = z.object({
  identifier: z
    .string()
    .trim()
    .min(1, 'Email or phone is required.')
    .max(254, 'Email or phone is too long.')
    .refine(isEmailOrPhone, 'Enter a valid email or phone number.'),
});

export const resetPasswordSchema = z
  .object({
    challenge_id: z.string().trim().min(1, 'OTP challenge is missing.'),
    confirmPassword: z.string().min(1, 'Confirm your new password.'),
    new_password: passwordSchema,
    otp: z.string().trim().regex(/^[0-9]{6}$/, 'Enter the 6 digit OTP.'),
  })
  .refine((data) => data.new_password === data.confirmPassword, {
    message: 'Passwords do not match.',
    path: ['confirmPassword'],
  });

export type LoginFormValues = z.infer<typeof loginSchema>;
export type SignupFormValues = z.infer<typeof signupSchema>;
export type OtpFormValues = z.infer<typeof otpSchema>;
export type ForgotPasswordFormValues = z.infer<typeof forgotPasswordSchema>;
export type ResetPasswordFormValues = z.infer<typeof resetPasswordSchema>;
